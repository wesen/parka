package builders

import (
	"os"
	
	"github.com/go-go-golems/clay/pkg/repositories"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	command_dir "github.com/go-go-golems/parka/pkg/handlers/command-dir"
	"github.com/go-go-golems/parka/pkg/handlers/config"
	generic_command "github.com/go-go-golems/parka/pkg/handlers/generic-command"
	parka "github.com/go-go-golems/parka/pkg/server"
)

// CommandDirHandlerBuilder is the builder for CommandDirHandler
type CommandDirHandlerBuilder struct {
	repository            *repositories.Repository
	configCommandDir      *config.CommandDir
	devMode               bool
	genericHandlerBuilder *GenericCommandHandlerBuilder
}

// NewCommandDirHandlerBuilder creates a new CommandDirHandlerBuilder
func NewCommandDirHandlerBuilder() *CommandDirHandlerBuilder {
	return &CommandDirHandlerBuilder{
		genericHandlerBuilder: NewGenericCommandHandlerBuilder(),
	}
}

// WithRepository sets the repository
func (b *CommandDirHandlerBuilder) WithRepository(repository *repositories.Repository) *CommandDirHandlerBuilder {
	b.repository = repository
	return b
}

// WithDevMode enables or disables dev mode
func (b *CommandDirHandlerBuilder) WithDevMode(devMode bool) *CommandDirHandlerBuilder {
	b.devMode = devMode
	return b
}

// FromConfig creates a builder from a config
func (b *CommandDirHandlerBuilder) FromConfig(config *config.CommandDir) *CommandDirHandlerBuilder {
	b.configCommandDir = config
	return b
}

// WithStream sets the stream mode
func (b *CommandDirHandlerBuilder) WithStream(stream bool) *CommandDirHandlerBuilder {
	b.genericHandlerBuilder.WithStream(stream)
	return b
}

// WithTemplateName sets the template name
func (b *CommandDirHandlerBuilder) WithTemplateName(name string) *CommandDirHandlerBuilder {
	b.genericHandlerBuilder.WithTemplateName(name)
	return b
}

// WithIndexTemplateName sets the index template name
func (b *CommandDirHandlerBuilder) WithIndexTemplateName(name string) *CommandDirHandlerBuilder {
	b.genericHandlerBuilder.WithIndexTemplateName(name)
	return b
}

// WithAdditionalData adds data to the template context
func (b *CommandDirHandlerBuilder) WithAdditionalData(data map[string]interface{}) *CommandDirHandlerBuilder {
	b.genericHandlerBuilder.WithAdditionalData(data)
	return b
}

// WithWhitelistedLayers sets the whitelisted layers
func (b *CommandDirHandlerBuilder) WithWhitelistedLayers(layers ...string) *CommandDirHandlerBuilder {
	b.genericHandlerBuilder.WithWhitelistedLayers(layers...)
	return b
}

// ConfigureParameters returns a builder for configuring parameters
func (b *CommandDirHandlerBuilder) ConfigureParameters() *ParameterConfigBuilderImpl {
	return b.genericHandlerBuilder.ConfigureParameters()
}

// ConfigureTemplate returns a builder for configuring templates
func (b *CommandDirHandlerBuilder) ConfigureTemplate() *TemplateConfigBuilderImpl {
	return b.genericHandlerBuilder.ConfigureTemplate()
}

// ConfigureMiddlewares returns a builder for configuring middlewares
func (b *CommandDirHandlerBuilder) ConfigureMiddlewares() *MiddlewareConfigBuilderImpl {
	return b.genericHandlerBuilder.ConfigureMiddlewares()
}

// Build creates a CommandDirHandler from the builder
func (b *CommandDirHandlerBuilder) Build() (*command_dir.CommandDirHandler, error) {
	var options []command_dir.CommandDirHandlerOption

	// Dev mode and repository
	options = append(options, command_dir.WithDevMode(b.devMode))
	if b.repository != nil {
		options = append(options, command_dir.WithRepository(b.repository))
	}

	// Build generic handler options
	genericHandler, err := b.genericHandlerBuilder.Build()
	if err != nil {
		return nil, err
	}

	// Convert generic handler to options
	var genericOptions []generic_command.GenericCommandHandlerOption
	if genericHandler.TemplateName != "" {
		genericOptions = append(genericOptions, generic_command.WithTemplateName(genericHandler.TemplateName))
	}
	if genericHandler.IndexTemplateName != "" {
		genericOptions = append(genericOptions, generic_command.WithIndexTemplateName(genericHandler.IndexTemplateName))
	}
	if genericHandler.TemplateLookup != nil {
		genericOptions = append(genericOptions, generic_command.WithTemplateLookup(genericHandler.TemplateLookup))
	}
	if len(genericHandler.AdditionalData) > 0 {
		genericOptions = append(genericOptions, generic_command.WithMergeAdditionalData(genericHandler.AdditionalData, true))
	}
	if genericHandler.ParameterFilter != nil {
		genericOptions = append(genericOptions, generic_command.WithParameterFilter(genericHandler.ParameterFilter))
	}
	if len(genericHandler.WhitelistedLayers) > 0 {
		genericOptions = append(genericOptions, generic_command.WithWhitelistedLayers(genericHandler.WhitelistedLayers...))
	}

	options = append(options, command_dir.WithGenericCommandHandlerOptions(genericOptions...))

	// Create handler
	if b.configCommandDir != nil {
		return command_dir.NewCommandDirHandlerFromConfig(b.configCommandDir, options...)
	}
	// Use existing constructor with options
	handler := &command_dir.CommandDirHandler{
		Repository: b.repository,
		DevMode:    b.devMode,
	}
	
	// Apply options - CommandDirHandlerOption doesn't return an error
	for _, option := range options {
		option(handler)
	}
	
	return handler, nil
}

// RegisterWith registers the handler with a server at the given path
func (b *CommandDirHandlerBuilder) RegisterWith(server *parka.Server, path string) error {
	handler, err := b.Build()
	if err != nil {
		return err
	}
	return handler.Serve(server, path)
}

// QuickCommandUI creates a simple web UI for browsing and executing commands
func QuickCommandUI(
	repository *repositories.Repository,
	path string,
	server *parka.Server,
	templateName string,
	preset *ParameterPreset,
) error {
	builder := NewCommandDirHandlerBuilder().
		WithRepository(repository).
		WithDevMode(true).
		WithTemplateName(templateName).
		WithIndexTemplateName("commands.tmpl.html").
		WithWhitelistedLayers(layers.DefaultSlug)

	if preset != nil {
		builder.ConfigureParameters().
			ApplyPreset(preset).
			Done()
	} else {
		// Apply standard web UI preset if none provided
		builder.ConfigureParameters().
			ApplyPreset(StandardWebUIPreset).
			Done()
	}

	handler, err := builder.Build()
	if err != nil {
		return err
	}

	return handler.Serve(server, path)
}

// RepositoryBuilder is a helper builder for repositories
type RepositoryBuilder struct {
	directories []repositories.Directory
}

// NewRepositoryBuilder creates a new repository builder
func NewRepositoryBuilder() *RepositoryBuilder {
	return &RepositoryBuilder{}
}

// AddDirectory adds a directory to the repository
func (b *RepositoryBuilder) AddDirectory(dir string) *RepositoryBuilder {
	b.directories = append(b.directories, repositories.Directory{
		FS:            os.DirFS(dir), // Use os.DirFS instead of repositories.OSFS
		RootDirectory: ".",
		WatchDirectory: dir,
	})
	return b
}

// Build creates a repository from the builder
func (b *RepositoryBuilder) Build() (*repositories.Repository, error) {
	var options []repositories.RepositoryOption
	if len(b.directories) > 0 {
		options = append(options, repositories.WithDirectories(b.directories...))
	}
	// Just return the repository - it seems NewRepository doesn't return an error in this codebase
	return repositories.NewRepository(options...), nil
}

// QuickRepository creates a repository from a directory
func QuickRepository(dir string) (*repositories.Repository, error) {
	// Just return the repository - it seems NewRepository doesn't return an error in this codebase
	return repositories.NewRepository(
		repositories.WithDirectories(repositories.Directory{
			FS:            os.DirFS(dir), // Use os.DirFS instead of repositories.OSFS
			RootDirectory: ".",
			WatchDirectory: dir,
		}),
	), nil
}