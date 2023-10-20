package writer

import (
	"context"
	"embed"
	"github.com/gin-gonic/gin"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/parka/pkg/glazed"
	"github.com/go-go-golems/parka/pkg/glazed/parser"
	"github.com/go-go-golems/parka/pkg/render"
	"github.com/go-go-golems/parka/pkg/render/layout"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"html/template"
	"io"
)

type WriterData struct {
	Command *cmds.CommandDescription
	// LongDescription is the HTML of the rendered markdown of the long description of the command.
	LongDescription template.HTML

	Layout *layout.Layout
	Links  []layout.Link

	BasePath string

	// AdditionalData to be passed to the rendering engine
	AdditionalData  map[string]interface{}
	CommandMetadata map[string]interface{}
}

func NewWriterData() *WriterData {
	return &WriterData{
		AdditionalData:  map[string]interface{}{},
		CommandMetadata: map[string]interface{}{},
	}
}

//go:embed templates/*
var templateFS embed.FS

func NewLookupTemplate() *render.LookupTemplateFromFS {
	l := render.NewLookupTemplateFromFS(
		render.WithFS(templateFS),
		render.WithBaseDir("templates/"),
		render.WithPatterns("**/*.tmpl.html"),
	)

	_ = l.Reload()

	return l
}

func (dt *WriterData) Clone() *WriterData {
	ret := *dt
	return &ret
}

type QueryHandler struct {
	cmd                cmds.WriterCommand
	contextMiddlewares []glazed.ContextMiddleware
	parserOptions      []parser.ParserOption

	templateName string
	lookup       render.TemplateLookup

	wd *WriterData
}

type QueryHandlerOption func(qh *QueryHandler)

func NewQueryHandler(
	cmd cmds.WriterCommand,
	options ...QueryHandlerOption,
) *QueryHandler {
	qh := &QueryHandler{
		cmd:          cmd,
		wd:           NewWriterData(),
		lookup:       NewLookupTemplate(),
		templateName: "data-tables.tmpl.html",
	}

	for _, option := range options {
		option(qh)
	}

	return qh
}
func WithWriterData(wd *WriterData) QueryHandlerOption {
	return func(qh *QueryHandler) {
		qh.wd = wd
	}
}

func WithContextMiddlewares(middlewares ...glazed.ContextMiddleware) QueryHandlerOption {
	return func(h *QueryHandler) {
		h.contextMiddlewares = middlewares
	}
}

// WithParserOptions sets the parser options for the QueryHandler
func WithParserOptions(options ...parser.ParserOption) QueryHandlerOption {
	return func(h *QueryHandler) {
		h.parserOptions = options
	}
}

func WithTemplateLookup(lookup render.TemplateLookup) QueryHandlerOption {
	return func(h *QueryHandler) {
		h.lookup = lookup
	}
}

func WithTemplateName(templateName string) QueryHandlerOption {
	return func(h *QueryHandler) {
		h.templateName = templateName
	}
}

func WithAdditionalData(data map[string]interface{}) QueryHandlerOption {
	return func(h *QueryHandler) {
		h.wd.AdditionalData = data
	}
}

func (qh *QueryHandler) Handle(c *gin.Context, w io.Writer) error {
	pc := glazed.NewCommandContext(qh.cmd)

	qh.contextMiddlewares = append(
		qh.contextMiddlewares,
		glazed.NewContextParserMiddleware(
			qh.cmd,
			glazed.NewCommandQueryParser(qh.cmd, true, true, qh.parserOptions...),
		),
	)

	for _, h := range qh.contextMiddlewares {
		err := h.Handle(c, pc)
		if err != nil {
			return err
		}
	}

	wd_ := qh.wd.Clone()

	ctx := c.Request.Context()
	ctx2, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
	}()
	//defer cancel()
	eg, _ := errgroup.WithContext(ctx2)

	// actually run the command
	allParameters := pc.GetAllParameterValues()
	// TODO(manuel, 2023-10-20) Actually run the command
	//eg.Go(func() error {
	//	buf := &bytes.Buffer{}
	//	// NOTE(manuel, 2023-10-16) The GetAllParameterValues is a bit of a hack because really what we want is to only get those flags through the layers
	//	err := qh.cmd.RunIntoWriter(ctx3, pc.ParsedLayers, allParameters, buf)
	//	if err != nil {
	//		return err
	//	}
	//
	//	s := buf.String()
	//	log.Info().Str("data", s).Msg("data from command")
	//
	//	return nil
	//})

	eg.Go(func() error {
		// if qh.Cmd implements cmds.CommandWithMetadata, get Metadata
		if cm_, ok := qh.cmd.(cmds.CommandWithMetadata); ok {
			var err error
			wd_.CommandMetadata, err = cm_.Metadata(c, pc.ParsedLayers, allParameters)
			if err != nil {
				return err
			}
		}
		err := qh.renderTemplate(c, pc, w, wd_)
		if err != nil {
			return err
		}

		return nil
	})

	return eg.Wait()
}

func (qh *QueryHandler) renderTemplate(
	c *gin.Context,
	pc *glazed.CommandContext,
	w io.Writer,
	wd_ *WriterData,
) error {
	// Here, we use the parsed layer to configure the glazed middlewares.
	// We then use the created formatters.TableOutputFormatter as a basis for
	// our own output formatter that renders into an HTML template.
	var err error

	t, err := qh.lookup.Lookup(qh.templateName)
	if err != nil {
		return err
	}

	layout_, err := layout.ComputeLayout(pc)
	if err != nil {
		return err
	}

	description := pc.Cmd.Description()

	longHTML, err := render.RenderMarkdownToHTML(description.Long)
	if err != nil {
		return err
	}

	wd_.Layout = layout_
	wd_.LongDescription = template.HTML(longHTML)
	wd_.Command = description

	// start copying from rowC to HTML or JS stream

	err = t.Execute(w, wd_)
	if err != nil {
		return err
	}

	return nil
}

func CreateHandler(
	cmd cmds.WriterCommand,
	path string,
	commandPath string,
	options ...QueryHandlerOption,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		links := []layout.Link{}

		wd := NewWriterData()
		wd.Command = cmd.Description()
		wd.Links = links
		wd.BasePath = path

		options_ := []QueryHandlerOption{
			WithWriterData(wd),
		}
		options_ = append(options_, options...)

		handler := NewQueryHandler(cmd, options_...)

		err := handler.Handle(c, c.Writer)
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Error().Err(err).Msg("error handling query")
		}
	}
}
