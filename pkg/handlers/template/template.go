package template

import (
	"github.com/go-go-golems/parka/pkg/handlers/config"
	"github.com/go-go-golems/parka/pkg/render"
	"github.com/go-go-golems/parka/pkg/server"
	"io/fs"
	"strings"
)

type TemplateHandler struct {
	fs              fs.FS
	TemplateFile    string
	rendererOptions []render.RendererOption
	renderer        *render.Renderer
	alwaysReload    bool
	// TODO(manuel, 2023-06-20) Allow to pass in additional data from code, not just config file
}

type TemplateHandlerOption func(handler *TemplateHandler)

func WithDefaultFS(fs fs.FS) TemplateHandlerOption {
	return func(handler *TemplateHandler) {
		if handler.fs == nil {
			handler.fs = fs
		}
	}
}

func WithAlwaysReload(alwaysReload bool) TemplateHandlerOption {
	return func(handler *TemplateHandler) {
		handler.alwaysReload = alwaysReload
	}
}

func WithAppendRendererOptions(rendererOptions ...render.RendererOption) TemplateHandlerOption {
	return func(handler *TemplateHandler) {
		handler.rendererOptions = append(handler.rendererOptions, rendererOptions...)
	}
}

func NewTemplateHandler(templateFile string, options ...TemplateHandlerOption) *TemplateHandler {
	handler := &TemplateHandler{
		TemplateFile: templateFile,
	}
	for _, option := range options {
		option(handler)
	}
	return handler
}

func NewTemplateHandlerFromConfig(
	path string,
	t *config.Template,
	options ...TemplateHandlerOption,
) (*TemplateHandler, error) {
	handler := &TemplateHandler{
		TemplateFile: t.TemplateFile,
	}
	for _, option := range options {
		option(handler)
	}

	path = strings.TrimPrefix(path, "/")
	if path == "" || strings.HasSuffix(path, "/") {
		path += "index.tmpl.md"
	}

	templateLookup := render.NewLookupTemplateFromFile(handler.TemplateFile, path)
	err := templateLookup.Reload()
	if err != nil {
		return nil, err
	}

	rendererOptions := append(handler.rendererOptions,
		render.WithPrependTemplateLookups(templateLookup),
	)
	// TODO(manuel, 2023-06-20) We need to pass the base template renderer to render out markdown
	r, err := render.NewRenderer(rendererOptions...)
	if err != nil {
		return nil, err
	}
	handler.renderer = r

	return handler, nil
}

func (t *TemplateHandler) Serve(server_ *server.Server, path string) error {

	server_.Router.Pre(t.renderer.HandleWithTrimPrefixMiddleware(path, nil))

	return nil
}
