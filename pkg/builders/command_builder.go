package builders

import (
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/loaders"
	"github.com/go-go-golems/parka/pkg/handlers/command"
	"github.com/go-go-golems/parka/pkg/handlers/config"
	generic_command "github.com/go-go-golems/parka/pkg/handlers/generic-command"
	parka "github.com/go-go-golems/parka/pkg/server"
)

// CommandHandlerBuilder is the builder for CommandHandler
type CommandHandlerBuilder struct {
	command                cmds.Command
	configCommand          *config.Command
	loader                 loaders.CommandLoader
	devMode                bool
	genericHandlerBuilder  *GenericCommandHandlerBuilder
}

// NewCommandHandlerBuilder creates a new CommandHandlerBuilder
func NewCommandHandlerBuilder() *CommandHandlerBuilder {
	return &CommandHandlerBuilder{
		genericHandlerBuilder: NewGenericCommandHandlerBuilder(),
	}
}

// WithCommand sets the command
func (b *CommandHandlerBuilder) WithCommand(command cmds.Command) *CommandHandlerBuilder {
	b.command = command
	return b
}

// WithDevMode enables or disables dev mode
func (b *CommandHandlerBuilder) WithDevMode(devMode bool) *CommandHandlerBuilder {
	b.devMode = devMode
	return b
}

// FromConfig creates a builder from a config
func (b *CommandHandlerBuilder) FromConfig(config *config.Command, loader loaders.CommandLoader) *CommandHandlerBuilder {
	b.configCommand = config
	b.loader = loader
	return b
}

// WithStream sets the stream mode
func (b *CommandHandlerBuilder) WithStream(stream bool) *CommandHandlerBuilder {
	b.genericHandlerBuilder.WithStream(stream)
	return b
}

// WithTemplateName sets the template name
func (b *CommandHandlerBuilder) WithTemplateName(name string) *CommandHandlerBuilder {
	b.genericHandlerBuilder.WithTemplateName(name)
	return b
}

// WithIndexTemplateName sets the index template name
func (b *CommandHandlerBuilder) WithIndexTemplateName(name string) *CommandHandlerBuilder {
	b.genericHandlerBuilder.WithIndexTemplateName(name)
	return b
}

// WithAdditionalData adds data to the template context
func (b *CommandHandlerBuilder) WithAdditionalData(data map[string]interface{}) *CommandHandlerBuilder {
	b.genericHandlerBuilder.WithAdditionalData(data)
	return b
}

// WithWhitelistedLayers sets the whitelisted layers
func (b *CommandHandlerBuilder) WithWhitelistedLayers(layers ...string) *CommandHandlerBuilder {
	b.genericHandlerBuilder.WithWhitelistedLayers(layers...)
	return b
}

// ConfigureParameters returns a builder for configuring parameters
func (b *CommandHandlerBuilder) ConfigureParameters() *ParameterConfigBuilderImpl {
	return b.genericHandlerBuilder.ConfigureParameters()
}

// ConfigureTemplate returns a builder for configuring templates
func (b *CommandHandlerBuilder) ConfigureTemplate() *TemplateConfigBuilderImpl {
	return b.genericHandlerBuilder.ConfigureTemplate()
}

// ConfigureMiddlewares returns a builder for configuring middlewares
func (b *CommandHandlerBuilder) ConfigureMiddlewares() *MiddlewareConfigBuilderImpl {
	return b.genericHandlerBuilder.ConfigureMiddlewares()
}

// Build creates a CommandHandler from the builder
func (b *CommandHandlerBuilder) Build() (*command.CommandHandler, error) {
	var options []command.CommandHandlerOption

	// Dev mode
	options = append(options, command.WithDevMode(b.devMode))

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

	options = append(options, command.WithGenericCommandHandlerOptions(genericOptions...))

	// Create handler
	if b.configCommand != nil {
		return command.NewCommandHandlerFromConfig(b.configCommand, b.loader, options...)
	}
	return command.NewCommandHandler(b.command, options...)
}

// RegisterWith registers the handler with a server at the given path
func (b *CommandHandlerBuilder) RegisterWith(server *parka.Server, path string) error {
	handler, err := b.Build()
	if err != nil {
		return err
	}
	return handler.Serve(server, path)
}

// QuickCommandAPI creates a simple API endpoint for a command
func QuickCommandAPI(
	command cmds.Command,
	path string,
	server *parka.Server,
	preset *ParameterPreset,
) error {
	builder := NewCommandHandlerBuilder().
		WithCommand(command).
		WithTemplateName("data-tables.tmpl.html")

	if preset != nil {
		builder.ConfigureParameters().
			ApplyPreset(preset).
			Done()
	} else {
		// Apply standard API preset if none provided
		builder.ConfigureParameters().
			ApplyPreset(StandardAPIPreset).
			Done()
	}

	handler, err := builder.Build()
	if err != nil {
		return err
	}

	return handler.Serve(server, path)
}