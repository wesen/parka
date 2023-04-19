package render

import (
	"bytes"
	"embed"
	"github.com/gin-gonic/gin"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/formatters/table"
	"github.com/go-go-golems/parka/pkg/glazed"
	"github.com/pkg/errors"
	"html/template"
)

// This file contains a variety of output renderers for HTML output.
// The idea would to create a set of glazed.CreateProcessorFunc
// that would return an OutputFormatter() that can be used to render
// a command and a table as HTML.

// HTMLTemplateOutputFormatter wraps a normal HTML table output formatter, and allows
// a template to be added in the back in the front.
type HTMLTemplateOutputFormatter struct {
	*table.OutputFormatter

	t *template.Template
}

func NewHTMLTemplateOutputFormatter(t *template.Template, of *table.OutputFormatter) *HTMLTemplateOutputFormatter {
	return &HTMLTemplateOutputFormatter{
		OutputFormatter: of,
		t:               t,
	}
}

func (H *HTMLTemplateOutputFormatter) Output() (string, error) {
	res, err := H.OutputFormatter.Output()
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	err = H.t.Execute(buf, map[string]interface{}{
		"Table": template.HTML(res),
	})

	if err != nil {
		return "", err
	}

	return buf.String(), err
}

type HTMLTemplateProcessor struct {
	*cmds.GlazeProcessor

	of *HTMLTemplateOutputFormatter
}

func NewHTMLTemplateProcessor(gp *cmds.GlazeProcessor, t *template.Template) (*HTMLTemplateProcessor, error) {
	parentOf, ok := gp.OutputFormatter().(*table.OutputFormatter)
	if !ok {
		return nil, errors.New("parent output formatter is not a table output formatter")
	}

	of := NewHTMLTemplateOutputFormatter(t, parentOf)
	return &HTMLTemplateProcessor{
		GlazeProcessor: gp,
		of:             of,
	}, nil
}

func (H *HTMLTemplateProcessor) OutputFormatter() formatters.OutputFormatter {
	return H.of
}

// NewTemplateLookupCreateProcessorFunc creates a CreateProcessorFunc based on a TemplateLookup
// and a template name.
func NewTemplateLookupCreateProcessorFunc(lookup TemplateLookup, templateName string) glazed.CreateProcessorFunc {
	return func(c *gin.Context, pc *glazed.CommandContext) (
		cmds.Processor,
		string, // content type
		error,
	) {
		contextType := "text/html"

		// lookup on every request, not up front.
		//
		// NOTE(manuel, 2023-04-19) This currently is nailed to a single static templateName passed at configuration time.
		// potentially, templateName could also be dynamic based on the incoming request, but we'll leave
		// that flexibility for later.
		t, err := lookup(templateName)
		if err != nil {
			return nil, contextType, err
		}

		// NOTE(manuel, 2023-04-18) We use glazed to render the actual HTML table.
		// But really, we could allow the user to specify the actual HTML rendering as well.
		// This is currently just a convenience to get started.
		l, ok := pc.ParsedLayers["glazed"]
		l.Parameters["output"] = "table"
		l.Parameters["table-format"] = "html"

		var gp *cmds.GlazeProcessor

		if ok {
			gp, err = cli.SetupProcessor(l.Parameters)
		} else {
			gp, err = cli.SetupProcessor(map[string]interface{}{
				"output":       "table",
				"table-format": "html",
			})
		}

		if err != nil {
			return nil, contextType, err
		}

		gp2, err := NewHTMLTemplateProcessor(gp, t)
		if err != nil {
			return nil, contextType, err
		}
		return gp2, contextType, nil
	}
}

//go:embed templates/*
var templateFS embed.FS

func NewDataTablesCreateProcessorFunc() (glazed.CreateProcessorFunc, error) {
	templateLookup, err := LookupTemplateFromFSReloadable(templateFS, "templates/", "templates/**/*.tmpl.html")
	if err != nil {
		return nil, err
	}

	return NewTemplateLookupCreateProcessorFunc(templateLookup, "data-tables.tmpl.html"), nil
}
