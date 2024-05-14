package generic_command

import (
	"fmt"
	"github.com/go-go-golems/clay/pkg/repositories"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/parka/pkg/glazed/handlers/datatables"
	"github.com/go-go-golems/parka/pkg/glazed/handlers/json"
	output_file "github.com/go-go-golems/parka/pkg/glazed/handlers/output-file"
	"github.com/go-go-golems/parka/pkg/glazed/handlers/sse"
	"github.com/go-go-golems/parka/pkg/glazed/handlers/text"
	"github.com/go-go-golems/parka/pkg/handlers/config"
	"github.com/go-go-golems/parka/pkg/render"
	parka "github.com/go-go-golems/parka/pkg/server"
	"github.com/go-go-golems/parka/pkg/utils"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"net/http"
	"path/filepath"
	"strings"
)

type GenericCommandHandler struct {
	// If true, all glazed outputs will try to use a row output if possible.
	// This means that "ragged" objects (where columns might not all be present)
	// will have missing columns, only the fields of the first object will be used
	// as rows.
	//
	// This is true per default, and needs to be explicitly set to false to use
	// a normal TableMiddleware oriented output.
	Stream bool

	// AdditionalData is passed to the template being rendered.
	AdditionalData map[string]interface{}

	ParameterFilter *config.ParameterFilter

	// TemplateName is the name of the template that is lookup up through the given TemplateLookup
	// used to render the glazed command.
	TemplateName string
	// IndexTemplateName is the name of the template that is looked up through TemplateLookup to render
	// command indexes. Leave empty to not render index pages at all.
	IndexTemplateName string
	// TemplateLookup is used to look up both TemplateName and IndexTemplateName
	TemplateLookup render.TemplateLookup

	// path under which the command handler is served
	BasePath string

	middlewares []middlewares.Middleware
}

func NewGenericCommandHandler(options ...GenericCommandHandlerOption) *GenericCommandHandler {
	handler := &GenericCommandHandler{
		AdditionalData:  map[string]interface{}{},
		ParameterFilter: &config.ParameterFilter{},
	}

	for _, opt := range options {
		opt(handler)
	}

	return handler
}

type GenericCommandHandlerOption func(handler *GenericCommandHandler)

func WithTemplateName(name string) GenericCommandHandlerOption {
	return func(handler *GenericCommandHandler) {
		handler.TemplateName = name
	}
}

func WithParameterFilter(overridesAndDefaults *config.ParameterFilter) GenericCommandHandlerOption {
	return func(handler *GenericCommandHandler) {
		handler.ParameterFilter = overridesAndDefaults
	}
}

func WithParameterFilterOptions(opts ...config.ParameterFilterOption) GenericCommandHandlerOption {
	return func(handler *GenericCommandHandler) {
		for _, opt := range opts {
			opt(handler.ParameterFilter)
		}
	}
}

func WithDefaultTemplateName(name string) GenericCommandHandlerOption {
	return func(handler *GenericCommandHandler) {
		if handler.TemplateName == "" {
			handler.TemplateName = name
		}
	}
}

func WithIndexTemplateName(name string) GenericCommandHandlerOption {
	return func(handler *GenericCommandHandler) {
		handler.IndexTemplateName = name
	}
}

func WithDefaultIndexTemplateName(name string) GenericCommandHandlerOption {
	return func(handler *GenericCommandHandler) {
		if handler.IndexTemplateName == "" {
			handler.IndexTemplateName = name
		}
	}
}

// WithMergeAdditionalData merges the passed in map with the handler's AdditionalData map.
// If a value is already set in the AdditionalData map and override is true, it will get overwritten.
func WithMergeAdditionalData(data map[string]interface{}, override bool) GenericCommandHandlerOption {
	return func(handler *GenericCommandHandler) {
		if handler.AdditionalData == nil {
			handler.AdditionalData = data
		} else {
			for k, v := range data {
				if _, ok := handler.AdditionalData[k]; !ok || override {
					handler.AdditionalData[k] = v
				}
			}
		}
	}
}

func WithTemplateLookup(lookup render.TemplateLookup) GenericCommandHandlerOption {
	return func(handler *GenericCommandHandler) {
		handler.TemplateLookup = lookup
	}
}

func (gch *GenericCommandHandler) ServeSingleCommand(server *parka.Server, basePath string, command cmds.Command) error {
	gch.BasePath = basePath

	gch.middlewares = gch.ParameterFilter.ComputeMiddlewares(gch.Stream)
	server.Router.GET(basePath+"/data", func(c echo.Context) error {
		return gch.ServeData(c, command)
	})
	server.Router.GET(basePath+"/text", func(c echo.Context) error {
		return gch.ServeText(c, command)
	})
	server.Router.GET(basePath+"/stream", func(c echo.Context) error {
		return gch.ServeStreaming(c, command)
	})
	server.Router.GET(basePath+"/download/*", func(c echo.Context) error {
		return gch.ServeDownload(c, command)
	})
	// don't use a specific datatables path here
	server.Router.GET(basePath, func(c echo.Context) error {
		return gch.ServeDataTables(c, command, basePath+"/download")
	})

	return nil
}

func (gch *GenericCommandHandler) ServeRepository(server *parka.Server, basePath string, repository *repositories.Repository) error {
	basePath = strings.TrimSuffix(basePath, "/")
	gch.BasePath = basePath

	gch.middlewares = gch.ParameterFilter.ComputeMiddlewares(gch.Stream)

	server.Router.GET(basePath+"/data/*", func(c echo.Context) error {
		commandPath := c.Param("*")
		commandPath = strings.TrimPrefix(commandPath, "/")
		command, err := getRepositoryCommand(repository, commandPath)
		if err != nil {
			log.Debug().
				Str("commandPath", commandPath).
				Str("basePath", basePath).
				Msg("could not find command")
			return err
		}

		return gch.ServeData(c, command)
	})

	server.Router.GET(basePath+"/text/*", func(c echo.Context) error {
		commandPath := c.Param("*")
		commandPath = strings.TrimPrefix(commandPath, "/")
		command, err := getRepositoryCommand(repository, commandPath)
		if err != nil {
			log.Debug().
				Str("commandPath", commandPath).
				Str("basePath", basePath).
				Msg("could not find command")
			return err
		}

		return gch.ServeText(c, command)
	})

	server.Router.GET(basePath+"/streaming/*", func(c echo.Context) error {
		commandPath := c.Param("*")
		commandPath = strings.TrimPrefix(commandPath, "/")
		command, err := getRepositoryCommand(repository, commandPath)
		if err != nil {
			log.Debug().
				Str("commandPath", commandPath).
				Str("basePath", basePath).
				Msg("could not find command")
			return err
		}

		return gch.ServeStreaming(c, command)
	})

	server.Router.GET(basePath+"/datatables/*", func(c echo.Context) error {
		commandPath := c.Param("*")
		commandPath = strings.TrimPrefix(commandPath, "/")

		// Get repository command
		command, err := getRepositoryCommand(repository, commandPath)
		if err != nil {
			log.Debug().
				Str("commandPath", commandPath).
				Str("basePath", basePath).
				Msg("could not find command")
			return err
		}

		return gch.ServeDataTables(c, command, basePath+"/download/"+commandPath)
	})

	server.Router.GET(basePath+"/download/*", func(c echo.Context) error {
		commandPath := c.Param("*")
		commandPath = strings.TrimPrefix(commandPath, "/")
		// strip file name from path
		index := strings.LastIndex(commandPath, "/")
		if index == -1 {
			return errors.New("could not find file name")
		}
		if index >= len(commandPath)-1 {
			return errors.New("could not find file name")
		}
		commandPath = commandPath[:index]

		command, err := getRepositoryCommand(repository, commandPath)
		if err != nil {
			log.Debug().
				Str("commandPath", commandPath).
				Str("basePath", basePath).
				Msg("could not find command")
			return err
		}

		return gch.ServeDownload(c, command)
	})

	return nil
}

func (gch *GenericCommandHandler) ServeData(c echo.Context, command cmds.Command) error {
	switch v := command.(type) {
	case cmds.GlazeCommand:
		return json.CreateJSONQueryHandler(v, json.WithMiddlewares(gch.middlewares...))(c)
	default:
		return text.CreateQueryHandler(v)(c)
	}
}

func (gch *GenericCommandHandler) ServeText(c echo.Context, command cmds.Command) error {
	return text.CreateQueryHandler(command, gch.middlewares...)(c)
}

func (gch *GenericCommandHandler) ServeStreaming(c echo.Context, command cmds.Command) error {
	return sse.CreateQueryHandler(command, gch.middlewares...)(c)
}

func (gch *GenericCommandHandler) ServeDataTables(c echo.Context, command cmds.Command, downloadPath string) error {
	switch v := command.(type) {
	case cmds.GlazeCommand:
		options := []datatables.QueryHandlerOption{
			datatables.WithMiddlewares(gch.middlewares...),
			datatables.WithTemplateLookup(gch.TemplateLookup),
			datatables.WithTemplateName(gch.TemplateName),
			datatables.WithAdditionalData(gch.AdditionalData),
			datatables.WithStreamRows(gch.Stream),
		}

		return datatables.CreateDataTablesHandler(v, gch.BasePath, downloadPath, options...)(c)
	default:
		return c.JSON(http.StatusInternalServerError, utils.H{"error": "command is not a glazed command"})
	}
}

func (gch *GenericCommandHandler) ServeDownload(c echo.Context, command cmds.Command) error {
	path_ := c.Request().URL.Path
	index := strings.LastIndex(path_, "/")
	if index == -1 {
		return c.JSON(http.StatusInternalServerError, utils.H{"error": "could not find file name"})
	}
	if index >= len(path_)-1 {
		return c.JSON(http.StatusInternalServerError, utils.H{"error": "could not find file name"})
	}
	fileName := path_[index+1:]

	switch v := command.(type) {
	case cmds.GlazeCommand:
		return output_file.CreateGlazedFileHandler(
			v,
			fileName,
			gch.middlewares...,
		)(c)

	case cmds.WriterCommand:
		handler := text.NewQueryHandler(command)

		baseName := filepath.Base(fileName)
		c.Response().Header().Set("Content-Disposition", "attachment; filename="+baseName)

		err := handler.Handle(c)
		if err != nil {
			return err
		}

		return nil

	default:
		return c.JSON(http.StatusInternalServerError, utils.H{"error": "command is not a glazed/writer command"})
	}

}

// getRepositoryCommand lookups a command in the given repository and return success as bool and the given command,
// or sends an error code over HTTP using the gin.Context.
func getRepositoryCommand(r *repositories.Repository, commandPath string) (cmds.Command, error) {
	path := strings.Split(commandPath, "/")
	commands := r.CollectCommands(path, false)
	if len(commands) == 0 {
		return nil, CommandNotFound{CommandPath: commandPath}
	}

	if len(commands) > 1 {
		err := &AmbiguousCommand{
			CommandPath: commandPath,
		}
		for _, command := range commands {
			description := command.Description()
			err.PotentialCommands = append(err.PotentialCommands, strings.Join(description.Parents, " ")+" "+description.Name)
		}
		return nil, err
	}

	// NOTE(manuel, 2023-05-15) Check if this is actually an alias, and populate the defaults from the alias flags
	// This could potentially be moved to the repository code itself

	return commands[0], nil
}

type CommandNotFound struct {
	CommandPath string
}

func (e CommandNotFound) Error() string {
	return fmt.Sprintf("command %s not found", e.CommandPath)
}

type AmbiguousCommand struct {
	CommandPath       string
	PotentialCommands []string
}

func (e AmbiguousCommand) Error() string {
	return fmt.Sprintf("command %s is ambiguous, could be one of: %s", e.CommandPath, strings.Join(e.PotentialCommands, ", "))

}
