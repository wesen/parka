package builders

import (
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/parka/pkg/handlers/config"
	"github.com/go-go-golems/parka/pkg/render"
)

// Builder is a common interface for all builders
type Builder interface {
	Build() (interface{}, error)
}

// ParameterConfigBuilder is a nested builder for parameter configuration
type ParameterConfigBuilder struct {
	parent interface{}
	config *config.ParameterFilter
}

// TemplateConfigBuilder is a nested builder for template configuration
type TemplateConfigBuilder struct {
	parent interface{}
	name   *string
	index  *string
	lookup render.TemplateLookup
	data   map[string]interface{}
}

// MiddlewareConfigBuilder is a nested builder for middleware configuration
type MiddlewareConfigBuilder struct {
	parent         interface{}
	preMiddlewares []middlewares.Middleware
	postMiddlewares []middlewares.Middleware
}

// ParameterPreset encapsulates common parameter configuration settings
type ParameterPreset struct {
	defaults        map[string]interface{}
	layerDefaults   map[string]map[string]interface{}
	overrides       map[string]interface{}
	layerOverrides  map[string]map[string]interface{}
	whitelistParams []string
	whitelistLayers []string
	blacklistParams []string
	blacklistLayers []string
	whitelistLayerParams map[string][]string
	blacklistLayerParams map[string][]string
}

// NewParameterPreset creates a new parameter preset
func NewParameterPreset() *ParameterPreset {
	return &ParameterPreset{
		defaults:        make(map[string]interface{}),
		layerDefaults:   make(map[string]map[string]interface{}),
		overrides:       make(map[string]interface{}),
		layerOverrides:  make(map[string]map[string]interface{}),
		whitelistLayerParams: make(map[string][]string),
		blacklistLayerParams: make(map[string][]string),
	}
}

// WithDefault adds a default parameter
func (p *ParameterPreset) WithDefault(name string, value interface{}) *ParameterPreset {
	p.defaults[name] = value
	return p
}

// WithDefaults adds multiple default parameters
func (p *ParameterPreset) WithDefaults(defaults map[string]interface{}) *ParameterPreset {
	for k, v := range defaults {
		p.defaults[k] = v
	}
	return p
}

// WithDefaultLayer adds a default layer
func (p *ParameterPreset) WithDefaultLayer(name string, values map[string]interface{}) *ParameterPreset {
	if _, ok := p.layerDefaults[name]; !ok {
		p.layerDefaults[name] = make(map[string]interface{})
	}
	for k, v := range values {
		p.layerDefaults[name][k] = v
	}
	return p
}

// WithOverride adds an override parameter
func (p *ParameterPreset) WithOverride(name string, value interface{}) *ParameterPreset {
	p.overrides[name] = value
	return p
}

// WithOverrides adds multiple override parameters
func (p *ParameterPreset) WithOverrides(overrides map[string]interface{}) *ParameterPreset {
	for k, v := range overrides {
		p.overrides[k] = v
	}
	return p
}

// WithOverrideLayer adds an override layer
func (p *ParameterPreset) WithOverrideLayer(name string, values map[string]interface{}) *ParameterPreset {
	if _, ok := p.layerOverrides[name]; !ok {
		p.layerOverrides[name] = make(map[string]interface{})
	}
	for k, v := range values {
		p.layerOverrides[name][k] = v
	}
	return p
}

// WhitelistParameters adds parameters to the whitelist
func (p *ParameterPreset) WhitelistParameters(params ...string) *ParameterPreset {
	p.whitelistParams = append(p.whitelistParams, params...)
	return p
}

// WhitelistLayers adds layers to the whitelist
func (p *ParameterPreset) WhitelistLayers(layers ...string) *ParameterPreset {
	p.whitelistLayers = append(p.whitelistLayers, layers...)
	return p
}

// WhitelistLayerParameters adds parameters to a layer's whitelist
func (p *ParameterPreset) WhitelistLayerParameters(layer string, params ...string) *ParameterPreset {
	if _, ok := p.whitelistLayerParams[layer]; !ok {
		p.whitelistLayerParams[layer] = []string{}
	}
	p.whitelistLayerParams[layer] = append(p.whitelistLayerParams[layer], params...)
	return p
}

// BlacklistParameters adds parameters to the blacklist
func (p *ParameterPreset) BlacklistParameters(params ...string) *ParameterPreset {
	p.blacklistParams = append(p.blacklistParams, params...)
	return p
}

// BlacklistLayers adds layers to the blacklist
func (p *ParameterPreset) BlacklistLayers(layers ...string) *ParameterPreset {
	p.blacklistLayers = append(p.blacklistLayers, layers...)
	return p
}

// BlacklistLayerParameters adds parameters to a layer's blacklist
func (p *ParameterPreset) BlacklistLayerParameters(layer string, params ...string) *ParameterPreset {
	if _, ok := p.blacklistLayerParams[layer]; !ok {
		p.blacklistLayerParams[layer] = []string{}
	}
	p.blacklistLayerParams[layer] = append(p.blacklistLayerParams[layer], params...)
	return p
}

// ToParameterFilterOptions converts the preset to parameter filter options
func (p *ParameterPreset) ToParameterFilterOptions() []config.ParameterFilterOption {
	var options []config.ParameterFilterOption

	// Add defaults
	if len(p.defaults) > 0 {
		options = append(options, config.WithDefaultParameters(p.defaults))
	}

	// Add layer defaults
	for name, values := range p.layerDefaults {
		options = append(options, config.WithMergeDefaultLayer(name, values))
	}

	// Add overrides
	if len(p.overrides) > 0 {
		options = append(options, config.WithOverrideParameters(p.overrides))
	}

	// Add layer overrides
	for name, values := range p.layerOverrides {
		options = append(options, config.WithMergeOverrideLayer(name, values))
	}

	// Add whitelists
	if len(p.whitelistParams) > 0 {
		options = append(options, config.WithWhitelistParameters(p.whitelistParams...))
	}
	if len(p.whitelistLayers) > 0 {
		options = append(options, config.WithWhitelistLayers(p.whitelistLayers...))
	}
	for layer, params := range p.whitelistLayerParams {
		options = append(options, config.WithWhitelistLayerParameters(layer, params...))
	}

	// Add blacklists
	if len(p.blacklistParams) > 0 {
		options = append(options, config.WithBlacklistParameters(p.blacklistParams...))
	}
	if len(p.blacklistLayers) > 0 {
		options = append(options, config.WithBlacklistLayers(p.blacklistLayers...))
	}
	for layer, params := range p.blacklistLayerParams {
		options = append(options, config.WithBlacklistLayerParameters(layer, params...))
	}

	return options
}

// Standard presets
var (
	// StandardSecurityPreset provides a preset with security-focused parameter filters
	StandardSecurityPreset = NewParameterPreset().
		BlacklistParameters("debug", "verbose").
		BlacklistLayerParameters("sql-connection", "password", "secret")

	// StandardAPIPreset provides a preset for API endpoints
	StandardAPIPreset = NewParameterPreset().
		WithDefault("format", "json").
		WithDefault("limit", "100").
		WhitelistParameters("limit", "offset", "format", "filter")

	// StandardWebUIPreset provides a preset for web UI endpoints
	StandardWebUIPreset = NewParameterPreset().
		WithDefault("format", "table").
		WithDefault("limit", "20").
		WhitelistParameters("limit", "offset", "format", "filter", "sort")
)