package builders

import (
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/parka/pkg/handlers/config"
	generic_command "github.com/go-go-golems/parka/pkg/handlers/generic-command"
	"github.com/go-go-golems/parka/pkg/render"
)

// GenericCommandHandlerBuilder is the builder for GenericCommandHandler
type GenericCommandHandlerBuilder struct {
	stream           bool
	templateName     string
	indexTemplateName string
	templateLookup   render.TemplateLookup
	additionalData   map[string]interface{}
	parameterFilter  *config.ParameterFilter
	preMiddlewares   []middlewares.Middleware
	postMiddlewares  []middlewares.Middleware
	whitelistedLayers []string
}

// NewGenericCommandHandlerBuilder creates a new GenericCommandHandlerBuilder
func NewGenericCommandHandlerBuilder() *GenericCommandHandlerBuilder {
	return &GenericCommandHandlerBuilder{
		stream:          true, // Default to true as in original implementation
		additionalData:  make(map[string]interface{}),
		parameterFilter: &config.ParameterFilter{},
	}
}

// WithStream configures streaming mode
func (b *GenericCommandHandlerBuilder) WithStream(stream bool) *GenericCommandHandlerBuilder {
	b.stream = stream
	return b
}

// WithTemplateName sets the template name
func (b *GenericCommandHandlerBuilder) WithTemplateName(name string) *GenericCommandHandlerBuilder {
	b.templateName = name
	return b
}

// WithIndexTemplateName sets the index template name
func (b *GenericCommandHandlerBuilder) WithIndexTemplateName(name string) *GenericCommandHandlerBuilder {
	b.indexTemplateName = name
	return b
}

// WithTemplateLookup sets the template lookup
func (b *GenericCommandHandlerBuilder) WithTemplateLookup(lookup render.TemplateLookup) *GenericCommandHandlerBuilder {
	b.templateLookup = lookup
	return b
}

// WithAdditionalData adds data to the template context
func (b *GenericCommandHandlerBuilder) WithAdditionalData(data map[string]interface{}) *GenericCommandHandlerBuilder {
	for k, v := range data {
		b.additionalData[k] = v
	}
	return b
}

// WithWhitelistedLayers sets the whitelisted layers
func (b *GenericCommandHandlerBuilder) WithWhitelistedLayers(layers ...string) *GenericCommandHandlerBuilder {
	b.whitelistedLayers = append(b.whitelistedLayers, layers...)
	return b
}

// WithPreMiddleware adds a middleware to run before parameter filtering
func (b *GenericCommandHandlerBuilder) WithPreMiddleware(middleware middlewares.Middleware) *GenericCommandHandlerBuilder {
	b.preMiddlewares = append(b.preMiddlewares, middleware)
	return b
}

// WithPostMiddleware adds a middleware to run after parameter filtering
func (b *GenericCommandHandlerBuilder) WithPostMiddleware(middleware middlewares.Middleware) *GenericCommandHandlerBuilder {
	b.postMiddlewares = append(b.postMiddlewares, middleware)
	return b
}

// ConfigureParameters returns a builder for configuring parameters
func (b *GenericCommandHandlerBuilder) ConfigureParameters() *ParameterConfigBuilderImpl {
	return &ParameterConfigBuilderImpl{
		parent: b,
		config: b.parameterFilter,
	}
}

// ConfigureTemplate returns a builder for configuring templates
func (b *GenericCommandHandlerBuilder) ConfigureTemplate() *TemplateConfigBuilderImpl {
	return &TemplateConfigBuilderImpl{
		parent: b,
		data:   b.additionalData,
	}
}

// ConfigureMiddlewares returns a builder for configuring middlewares
func (b *GenericCommandHandlerBuilder) ConfigureMiddlewares() *MiddlewareConfigBuilderImpl {
	return &MiddlewareConfigBuilderImpl{
		parent:         b,
		preMiddlewares: b.preMiddlewares,
		postMiddlewares: b.postMiddlewares,
	}
}

// Build creates a GenericCommandHandler from the builder
func (b *GenericCommandHandlerBuilder) Build() (*generic_command.GenericCommandHandler, error) {
	var options []generic_command.GenericCommandHandlerOption

	// Template configuration
	if b.templateName != "" {
		options = append(options, generic_command.WithTemplateName(b.templateName))
	}
	if b.indexTemplateName != "" {
		options = append(options, generic_command.WithIndexTemplateName(b.indexTemplateName))
	}
	if b.templateLookup != nil {
		options = append(options, generic_command.WithTemplateLookup(b.templateLookup))
	}
	if len(b.additionalData) > 0 {
		options = append(options, generic_command.WithMergeAdditionalData(b.additionalData, true))
	}

	// Parameter configuration
	options = append(options, generic_command.WithParameterFilter(b.parameterFilter))

	// Middleware configuration
	if len(b.preMiddlewares) > 0 {
		options = append(options, generic_command.WithPreMiddlewares(b.preMiddlewares...))
	}
	if len(b.postMiddlewares) > 0 {
		options = append(options, generic_command.WithPostMiddlewares(b.postMiddlewares...))
	}

	// Whitelisted layers
	if len(b.whitelistedLayers) > 0 {
		options = append(options, generic_command.WithWhitelistedLayers(b.whitelistedLayers...))
	}

	// Create handler
	return generic_command.NewGenericCommandHandler(options...)
}

// ParameterConfigBuilderImpl is the implementation of ParameterConfigBuilder
type ParameterConfigBuilderImpl struct {
	parent *GenericCommandHandlerBuilder
	config *config.ParameterFilter
}

// WithDefault sets a default parameter
func (b *ParameterConfigBuilderImpl) WithDefault(name string, value interface{}) *ParameterConfigBuilderImpl {
	config.WithDefaultParameter(name, value)(b.config)
	return b
}

// WithDefaults sets multiple default parameters
func (b *ParameterConfigBuilderImpl) WithDefaults(params map[string]interface{}) *ParameterConfigBuilderImpl {
	config.WithDefaultParameters(params)(b.config)
	return b
}

// WithDefaultLayer sets default layer parameters
func (b *ParameterConfigBuilderImpl) WithDefaultLayer(name string, values map[string]interface{}) *ParameterConfigBuilderImpl {
	config.WithMergeDefaultLayer(name, values)(b.config)
	return b
}

// WithOverride sets an override parameter
func (b *ParameterConfigBuilderImpl) WithOverride(name string, value interface{}) *ParameterConfigBuilderImpl {
	config.WithOverrideParameter(name, value)(b.config)
	return b
}

// WithOverrides sets multiple override parameters
func (b *ParameterConfigBuilderImpl) WithOverrides(params map[string]interface{}) *ParameterConfigBuilderImpl {
	config.WithOverrideParameters(params)(b.config)
	return b
}

// WithOverrideLayer sets override layer parameters
func (b *ParameterConfigBuilderImpl) WithOverrideLayer(name string, values map[string]interface{}) *ParameterConfigBuilderImpl {
	config.WithMergeOverrideLayer(name, values)(b.config)
	return b
}

// WhitelistParameters adds parameters to the whitelist
func (b *ParameterConfigBuilderImpl) WhitelistParameters(params ...string) *ParameterConfigBuilderImpl {
	config.WithWhitelistParameters(params...)(b.config)
	return b
}

// WhitelistLayers adds layers to the whitelist
func (b *ParameterConfigBuilderImpl) WhitelistLayers(layers ...string) *ParameterConfigBuilderImpl {
	config.WithWhitelistLayers(layers...)(b.config)
	return b
}

// WhitelistLayerParameters adds parameters to a layer's whitelist
func (b *ParameterConfigBuilderImpl) WhitelistLayerParameters(layer string, params ...string) *ParameterConfigBuilderImpl {
	config.WithWhitelistLayerParameters(layer, params...)(b.config)
	return b
}

// BlacklistParameters adds parameters to the blacklist
func (b *ParameterConfigBuilderImpl) BlacklistParameters(params ...string) *ParameterConfigBuilderImpl {
	config.WithBlacklistParameters(params...)(b.config)
	return b
}

// BlacklistLayers adds layers to the blacklist
func (b *ParameterConfigBuilderImpl) BlacklistLayers(layers ...string) *ParameterConfigBuilderImpl {
	config.WithBlacklistLayers(layers...)(b.config)
	return b
}

// BlacklistLayerParameters adds parameters to a layer's blacklist
func (b *ParameterConfigBuilderImpl) BlacklistLayerParameters(layer string, params ...string) *ParameterConfigBuilderImpl {
	config.WithBlacklistLayerParameters(layer, params...)(b.config)
	return b
}

// ApplyPreset applies a parameter preset
func (b *ParameterConfigBuilderImpl) ApplyPreset(preset *ParameterPreset) *ParameterConfigBuilderImpl {
	for _, opt := range preset.ToParameterFilterOptions() {
		opt(b.config)
	}
	return b
}

// Done returns to the parent builder
func (b *ParameterConfigBuilderImpl) Done() *GenericCommandHandlerBuilder {
	return b.parent
}

// TemplateConfigBuilderImpl is the implementation of TemplateConfigBuilder
type TemplateConfigBuilderImpl struct {
	parent *GenericCommandHandlerBuilder
	name   string
	index  string
	lookup render.TemplateLookup
	data   map[string]interface{}
}

// WithName sets the template name
func (b *TemplateConfigBuilderImpl) WithName(name string) *TemplateConfigBuilderImpl {
	b.name = name
	b.parent.templateName = name
	return b
}

// WithIndexName sets the index template name
func (b *TemplateConfigBuilderImpl) WithIndexName(name string) *TemplateConfigBuilderImpl {
	b.index = name
	b.parent.indexTemplateName = name
	return b
}

// WithLookup sets the template lookup
func (b *TemplateConfigBuilderImpl) WithLookup(lookup render.TemplateLookup) *TemplateConfigBuilderImpl {
	b.lookup = lookup
	b.parent.templateLookup = lookup
	return b
}

// WithData adds data to the template context
func (b *TemplateConfigBuilderImpl) WithData(key string, value interface{}) *TemplateConfigBuilderImpl {
	b.data[key] = value
	b.parent.additionalData[key] = value
	return b
}

// WithDataMap adds a map of data to the template context
func (b *TemplateConfigBuilderImpl) WithDataMap(data map[string]interface{}) *TemplateConfigBuilderImpl {
	for k, v := range data {
		b.data[k] = v
		b.parent.additionalData[k] = v
	}
	return b
}

// Done returns to the parent builder
func (b *TemplateConfigBuilderImpl) Done() *GenericCommandHandlerBuilder {
	return b.parent
}

// MiddlewareConfigBuilderImpl is the implementation of MiddlewareConfigBuilder
type MiddlewareConfigBuilderImpl struct {
	parent         *GenericCommandHandlerBuilder
	preMiddlewares []middlewares.Middleware
	postMiddlewares []middlewares.Middleware
}

// WithPreMiddleware adds a middleware to run before parameter filtering
func (b *MiddlewareConfigBuilderImpl) WithPreMiddleware(middleware middlewares.Middleware) *MiddlewareConfigBuilderImpl {
	b.preMiddlewares = append(b.preMiddlewares, middleware)
	b.parent.preMiddlewares = append(b.parent.preMiddlewares, middleware)
	return b
}

// WithPreMiddlewares adds multiple middlewares to run before parameter filtering
func (b *MiddlewareConfigBuilderImpl) WithPreMiddlewares(middlewares ...middlewares.Middleware) *MiddlewareConfigBuilderImpl {
	b.preMiddlewares = append(b.preMiddlewares, middlewares...)
	b.parent.preMiddlewares = append(b.parent.preMiddlewares, middlewares...)
	return b
}

// WithPostMiddleware adds a middleware to run after parameter filtering
func (b *MiddlewareConfigBuilderImpl) WithPostMiddleware(middleware middlewares.Middleware) *MiddlewareConfigBuilderImpl {
	b.postMiddlewares = append(b.postMiddlewares, middleware)
	b.parent.postMiddlewares = append(b.parent.postMiddlewares, middleware)
	return b
}

// WithPostMiddlewares adds multiple middlewares to run after parameter filtering
func (b *MiddlewareConfigBuilderImpl) WithPostMiddlewares(middlewares ...middlewares.Middleware) *MiddlewareConfigBuilderImpl {
	b.postMiddlewares = append(b.postMiddlewares, middlewares...)
	b.parent.postMiddlewares = append(b.parent.postMiddlewares, middlewares...)
	return b
}

// Done returns to the parent builder
func (b *MiddlewareConfigBuilderImpl) Done() *GenericCommandHandlerBuilder {
	return b.parent
}