package pkg

import (
	"embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-go-golems/parka/pkg/render"
	"html/template"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed "web/src/templates/*"
var templateFS embed.FS

//go:embed "web/dist/*"
var distFS embed.FS

// StaticPath allows you to serve static files from a http.FileSystem under a given URL path urlPath.
type StaticPath struct {
	fs      http.FileSystem
	urlPath string
}

func NewStaticPath(fs http.FileSystem, urlPath string) StaticPath {
	return StaticPath{
		fs:      fs,
		urlPath: urlPath,
	}
}

// Server is the main class that parka uses to serve static and templated content.
// It is a wrapper around gin.Engine.
//
// It is meant to be quite flexible, allowing you to add static paths and template lookups
// that can provide different fs and template backends.
//
// Router is the gin.Engine that is used to serve the content, and it is exposed so that you
// can use it as just a gin.Engine if you want to.
type Server struct {
	Router *gin.Engine

	StaticPaths     []StaticPath
	TemplateLookups []render.TemplateLookup

	Port    uint16
	Address string
}

type ServerOption = func(*Server)

// WithPrependTemplateLookups will prepend the given template lookups to the list of lookups,
// ensuring that they will be found before whatever templates might already be in the list.
func WithPrependTemplateLookups(lookups ...render.TemplateLookup) ServerOption {
	return func(s *Server) {
		// prepend lookups to the list
		s.TemplateLookups = append(lookups, s.TemplateLookups...)
	}
}

// WithAppendTemplateLookups will append the given template lookups to the list of lookups,
// but they will be found after whatever templates might already be in the list. This is great
// for providing fallback templates.
func WithAppendTemplateLookups(lookups ...render.TemplateLookup) ServerOption {
	return func(s *Server) {
		// append lookups to the list
		s.TemplateLookups = append(s.TemplateLookups, lookups...)
	}
}

// WithReplaceTemplateLookups will replace any existing template lookups with the given ones.
func WithReplaceTemplateLookups(lookups ...render.TemplateLookup) ServerOption {
	return func(s *Server) {
		s.TemplateLookups = lookups
	}
}

// WithStaticPaths will add the given static paths to the list of static paths.
// If a path with the same URL path already exists, it will be replaced.
func WithStaticPaths(paths ...StaticPath) ServerOption {
	return func(s *Server) {
		// prepend paths to the list
	pathLoop:
		for _, path := range paths {
			for i, existingPath := range s.StaticPaths {
				if existingPath.urlPath == path.urlPath {
					s.StaticPaths[i] = path
					continue pathLoop
				}
			}
			s.StaticPaths = append(s.StaticPaths, path)
		}
	}
}

// WithPort will set the port that the server will listen on.
func WithPort(port uint16) ServerOption {
	return func(s *Server) {
		s.Port = port
	}
}

// WithAddress will set the address that the server will listen on.
func WithAddress(address string) ServerOption {
	return func(s *Server) {
		s.Address = address
	}
}

// NewServer will create a new Server with the given options.
// This loads a fixed set of files and templates from the embed.FS.
// These files provide tailwind support for Markdown rendering and a standard index and base page template.
// NOTE(manuel, 2023-04-16) This is definitely ripe to be removed.
func NewServer(options ...ServerOption) (*Server, error) {
	router := gin.Default()

	parkaLookup, err := render.LookupTemplateFromFS(templateFS, "web/src/templates", "**/*.tmpl.*")
	if err != nil {
		return nil, err
	}

	s := &Server{
		Router: router,
		StaticPaths: []StaticPath{
			NewStaticPath(NewEmbedFileSystem(distFS, "web/dist"), "/dist"),
		},
		TemplateLookups: []render.TemplateLookup{
			parkaLookup,
		},
	}

	for _, option := range options {
		option(s)
	}

	return s, nil
}

// EmbedFileSystem is a helper to make an embed FS work as a http.FS,
// which allows us to serve embed.FS using gin's `Static` middleware.
type EmbedFileSystem struct {
	f           http.FileSystem
	stripPrefix string
}

// NewEmbedFileSystem will create a new EmbedFileSystem that will serve the given embed.FS
// under the given URL path. stripPrefix will be added to the beginning of all paths when
// looking up files in the embed.FS.
func NewEmbedFileSystem(f fs.FS, stripPrefix string) *EmbedFileSystem {
	if !strings.HasSuffix(stripPrefix, "/") {
		stripPrefix += "/"
	}
	return &EmbedFileSystem{
		f:           http.FS(f),
		stripPrefix: stripPrefix,
	}
}

// Open will open the file with the given name from the embed.FS. The name will be prefixed
// with the stripPrefix that was given when creating the EmbedFileSystem.
func (e *EmbedFileSystem) Open(name string) (http.File, error) {
	name = strings.TrimPrefix(name, "/")
	return e.f.Open(e.stripPrefix + name)
}

// Exists will check if the given path exists in the embed.FS. The path will be prefixed
// with the stripPrefix that was given when creating the EmbedFileSystem, while prefix will
// be removed from the path.
func (e *EmbedFileSystem) Exists(prefix string, path string) bool {
	if len(path) < len(prefix) {
		return false
	}

	// remove prefix from path
	path = path[len(prefix):]

	f, err := e.f.Open(e.stripPrefix + path)
	if err != nil {
		return false
	}
	defer f.Close()
	return true
}

// LookupTemplate will iterate through the template lookups until it finds one of the
// templates given in name.
func (s *Server) LookupTemplate(name ...string) (*template.Template, error) {
	var t *template.Template

	for _, lookup := range s.TemplateLookups {
		t, err := lookup(name...)
		if err == nil {
			return t, nil
		}
	}

	return t, nil
}

// serverMarkdownTemplatePage is an internal helper function to look up a markdown or HTML file
// and serve it.
//
// It first looks for a markdown file or template called either page.md or page.tmpl.md,
// and render it as a template, passing it the given data.
// It will use base.tmpl.html as the base template for serving the resulting markdown HTML.
// page.md is rendered as a plain markdown file, while page.tmpl.md is rendered as a template.
//
// If no markdown file or template is found, it will look for a HTML file or template called
// either page.html or page.tmpl.html and serve it as a template, passing it the given data.
// page.html is served as a plain HTML file, while page.tmpl.html is served as a template.
func (s *Server) serveMarkdownTemplatePage(c *gin.Context, page string, data interface{}) {
	t, err := s.LookupTemplate(page+".tmpl.md", page+".md")
	if err != nil {
		c.String(http.StatusInternalServerError, "Error rendering template")
		return
	}

	if t != nil {
		markdown, err := render.RenderMarkdownTemplateToHTML(t, nil)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error rendering markdown")
			return
		}

		baseTemplate, err := s.LookupTemplate("base.tmpl.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Error rendering template")
			return
		}

		err = baseTemplate.Execute(
			c.Writer,
			map[string]interface{}{
				"markdown": template.HTML(markdown),
			})
		if err != nil {
			c.String(http.StatusInternalServerError, "Error rendering template")
			return
		}
	} else {
		t, err = s.LookupTemplate(page+".tmpl.html", page+".html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Error rendering template")
			return
		}
		if t == nil {
			c.String(http.StatusInternalServerError, "Error rendering template")
			return
		}

		err := t.Execute(c.Writer, data)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error rendering template")
			return
		}
	}
}

// Run will start the server and listen on the given address and port.
func (s *Server) Run() error {
	for _, path := range s.StaticPaths {
		s.Router.StaticFS(path.urlPath, path.fs)
	}

	s.Router.GET("/", func(c *gin.Context) {
		s.serveMarkdownTemplatePage(c, "index", nil)
	})
	s.Router.GET("/:page", func(c *gin.Context) {
		page := c.Param("page")
		s.serveMarkdownTemplatePage(c, page, nil)
	})

	addr := fmt.Sprintf("%s:%d", s.Address, s.Port)
	return s.Router.Run(addr)
}
