package htmx

import (
	"context"
	"embed"
	"html/template"
	"io"
	"net/http"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/formatters/json"
	"github.com/go-go-golems/glazed/pkg/middlewares/row"
	"github.com/go-go-golems/parka/pkg/glazed/handlers"
	parka_middlewares "github.com/go-go-golems/parka/pkg/glazed/middlewares"
	"github.com/go-go-golems/parka/pkg/render"
	"github.com/go-go-golems/parka/pkg/render/layout"
	"github.com/kucherenkovova/safegroup"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

// FormData describes the data passed to the HTMX form template
type FormData struct {
	Command *cmds.CommandDescription
	// LongDescription is the HTML of the rendered markdown
	LongDescription template.HTML

	Layout *layout.Layout

	// Stream provides results
	ResultStream chan template.HTML
	ErrorStream  chan string

	// Additional data for template rendering
	AdditionalData map[string]interface{}
	// Metadata about the command
	CommandMetadata map[string]interface{}
}

func NewFormData() *FormData {
	return &FormData{
		AdditionalData:  make(map[string]interface{}),
		CommandMetadata: make(map[string]interface{}),
	}
}

//go:embed templates/*
var templateFS embed.FS

func NewFormTemplateLookup() *render.LookupTemplateFromFS {
	l := render.NewLookupTemplateFromFS(
		render.WithFS(templateFS),
		render.WithBaseDir("templates/"),
		render.WithPatterns("**/*.tmpl.html"),
	)

	_ = l.Reload()

	return l
}

type FormHandler struct {
	cmd         cmds.GlazeCommand
	middlewares []middlewares.Middleware

	templateName string
	lookup       render.TemplateLookup

	fd *FormData
}

type FormHandlerOption func(h *FormHandler)

func NewFormHandler(
	cmd cmds.GlazeCommand,
	options ...FormHandlerOption,
) *FormHandler {
	h := &FormHandler{
		cmd:          cmd,
		fd:           NewFormData(),
		lookup:       NewFormTemplateLookup(),
		templateName: "form.tmpl.html",
	}

	for _, option := range options {
		option(h)
	}

	return h
}

func WithFormData(fd *FormData) FormHandlerOption {
	return func(h *FormHandler) {
		h.fd = fd
	}
}

func WithMiddlewares(middlewares ...middlewares.Middleware) FormHandlerOption {
	return func(h *FormHandler) {
		h.middlewares = middlewares
	}
}

func WithTemplateLookup(lookup render.TemplateLookup) FormHandlerOption {
	return func(h *FormHandler) {
		h.lookup = lookup
	}
}

func WithTemplateName(templateName string) FormHandlerOption {
	return func(h *FormHandler) {
		h.templateName = templateName
	}
}

func WithAdditionalData(data map[string]interface{}) FormHandlerOption {
	return func(h *FormHandler) {
		h.fd.AdditionalData = data
	}
}

// ResultsData is used for the results-only template
type ResultsData struct {
	ResultStream chan template.HTML
	ErrorStream  chan string
}

func (h *FormHandler) RegisterRoutes(g *echo.Group) {
	// Add redirect from /example-htmx to /example-htmx/
	g.GET("", func(c echo.Context) error {
		if c.Request().URL.Path[len(c.Request().URL.Path)-1] != '/' {
			return c.Redirect(http.StatusMovedPermanently, c.Request().URL.Path+"/")
		}
		return h.HandleForm(c)
	})
	g.GET("/", h.HandleForm)
	// Results-only route for form submission
	g.GET("/submit", h.HandleResults)
}

// HandleForm renders the complete form template
func (h *FormHandler) HandleForm(c echo.Context) error {
	description := h.cmd.Description()
	parsedLayers := layers.NewParsedLayers()

	// Process middlewares including query parameters
	err := middlewares.ExecuteMiddlewares(description.Layers, parsedLayers,
		append(
			h.middlewares,
			parka_middlewares.UpdateFromQueryParameters(c, nil),
			middlewares.SetFromDefaults(),
		)...,
	)

	if err != nil {
		return h.handleError(err, c.Response(), parsedLayers)
	}

	// Get command metadata if available
	if cm, ok := h.cmd.(cmds.CommandWithMetadata); ok {
		h.fd.CommandMetadata, err = cm.Metadata(c.Request().Context(), parsedLayers)
		if err != nil {
			return err
		}
	}

	return h.renderTemplate(parsedLayers, c.Response(), "form.tmpl.html", h.fd)
}

// HandleResults processes the form submission and renders only the results
func (h *FormHandler) HandleResults(c echo.Context) error {
	description := h.cmd.Description()
	parsedLayers := layers.NewParsedLayers()

	resultC := make(chan string, 100)
	resultsData := &ResultsData{
		ResultStream: make(chan template.HTML, 100),
		ErrorStream:  make(chan string, 1),
	}

	// Process middlewares including query parameters
	err := middlewares.ExecuteMiddlewares(description.Layers, parsedLayers,
		append(
			h.middlewares,
			parka_middlewares.UpdateFromQueryParameters(c),
			middlewares.SetFromDefaults(),
		)...,
	)

	if err != nil {
		resultsData.ErrorStream <- err.Error()
		close(resultsData.ErrorStream)
		return h.renderTemplate(parsedLayers, c.Response(), "results.tmpl.html", resultsData)
	}

	// Setup output formatter
	of := json.NewOutputFormatter(json.WithOutputIndividualRows(true))

	// Create processor for results
	gp, err := handlers.CreateTableProcessorWithOutput(parsedLayers, "table", "ascii")
	if err != nil {
		return err
	}

	// Add row middleware for output
	gp.ReplaceTableMiddleware()
	gp.AddRowMiddleware(row.NewOutputChannelMiddleware(of, resultC))

	ctx := c.Request().Context()
	eg, ctx2 := safegroup.WithContext(ctx)

	// Process results
	eg.Go(func() error {
		defer close(resultsData.ResultStream)
		for {
			select {
			case <-ctx2.Done():
				return ctx2.Err()
			case result, ok := <-resultC:
				if !ok {
					return nil
				}
				resultsData.ResultStream <- template.HTML(result)
			}
		}
	})

	// Run command
	eg.Go(func() error {
		err := h.cmd.RunIntoGlazeProcessor(ctx2, parsedLayers, gp)
		if err != nil {
			resultsData.ErrorStream <- err.Error()
		}
		close(resultsData.ErrorStream)
		close(resultC)
		return nil
	})

	// Render results template
	eg.Go(func() error {
		return h.renderTemplate(parsedLayers, c.Response(), "results.tmpl.html", resultsData)
	})

	err = eg.Wait()
	if err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}

func (h *FormHandler) handleError(err error, w io.Writer, parsedLayers *layers.ParsedLayers) error {
	h.fd.ErrorStream <- err.Error()
	close(h.fd.ErrorStream)
	return h.renderTemplate(parsedLayers, w, h.templateName, h.fd)
}

func (h *FormHandler) renderTemplate(parsedLayers *layers.ParsedLayers, w io.Writer, templateName string, data interface{}) error {
	t, err := h.lookup.Lookup(templateName)
	if err != nil {
		return err
	}

	if formData, ok := data.(*FormData); ok {
		layout_, err := layout.ComputeLayout(h.cmd, parsedLayers)
		if err != nil {
			return err
		}

		description := h.cmd.Description()
		longHTML, err := render.RenderMarkdownToHTML(description.Long)
		if err != nil {
			return err
		}

		formData.Layout = layout_
		formData.LongDescription = template.HTML(longHTML)
		formData.Command = description
	}

	return t.Execute(w, data)
}
