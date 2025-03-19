# Parka Refactor Preliminary Information

This document contains the detailed information about the Parka codebase needed to implement the builder pattern refactor proposed in REFACTOR-RFC.md.

## 1. Handler Types Inventory

### Core Handlers

1. **GenericCommandHandler**
   - Base handler that all command handlers extend
   - Constructor: `NewGenericCommandHandler(options ...GenericCommandHandlerOption)`
   - Location: `pkg/handlers/generic-command/generic.go`

2. **CommandHandler**
   - Embeds GenericCommandHandler
   - Constructor: `NewCommandHandler(command cmds.Command, options ...CommandHandlerOption)`
   - Config-based constructor: `NewCommandHandlerFromConfig(config_ *config.Command, loader loaders.CommandLoader, options ...CommandHandlerOption)`
   - Location: `pkg/handlers/command/command.go`

3. **CommandDirHandler**
   - Embeds GenericCommandHandler
   - Constructor: `NewCommandDirHandler(options ...CommandDirHandlerOption)`
   - Config-based constructor: `NewCommandDirHandlerFromConfig(config_ *config.CommandDir, options ...CommandDirHandlerOption)`
   - Location: `pkg/handlers/command-dir/command-dir.go`

4. **TemplateHandler**
   - Constructor: `NewTemplateHandler(templateFile string, options ...TemplateHandlerOption)`
   - Config-based constructor: `NewTemplateHandlerFromConfig(t *config.Template, options ...TemplateHandlerOption)`
   - Location: `pkg/handlers/template/template.go`

5. **TemplateDirHandler**
   - Constructor: `NewTemplateDirHandler(options ...TemplateDirHandlerOption)`
   - Config-based constructor: `NewTemplateDirHandlerFromConfig(td *config.TemplateDir, options ...TemplateDirHandlerOption)`
   - Location: `pkg/handlers/template-dir/template-dir.go`

6. **StaticDirHandler**
   - Constructor: `NewStaticDirHandler(options ...StaticDirHandlerOption)`
   - Config-based constructor: `NewStaticDirHandlerFromConfig(sh *config.Static, options ...StaticDirHandlerOption)`
   - Location: `pkg/handlers/static-dir/static.go`

7. **StaticFileHandler**
   - Constructor: `NewStaticFileHandler(options ...StaticFileHandlerOption)`
   - Config-based constructor: `NewStaticFileHandlerFromConfig(shf *config.StaticFile, options ...StaticFileHandlerOption)`
   - Location: `pkg/handlers/static-file/static-file.go`

### Glazed Handlers (Output Format Specific)

1. **JSON Handler**
   - Constructor: `NewQueryHandler(cmd cmds.Command, options ...QueryHandlerOption)`
   - Factory functions: `CreateJSONQueryHandler`, `CreateJSONBodyHandler`
   - Location: `pkg/glazed/handlers/json/json.go`

2. **DataTables Handler**
   - Constructor: `NewQueryHandler(cmd cmds.GlazeCommand, options ...QueryHandlerOption)`
   - Factory function: `CreateDataTablesHandler`
   - Location: `pkg/glazed/handlers/datatables/datatables.go`

3. **Text Handler**
   - Constructor: `NewHandler(cmd cmds.Command, options ...HandleOption)`
   - Location: `pkg/glazed/handlers/text/text.go`

4. **SSE Handler**
   - Constructor: `NewHandler(cmd cmds.Command, options ...HandlerOption)`
   - Location: `pkg/glazed/handlers/sse/sse.go`

5. **Output File Handler**
   - Constructor: `NewHandler(cmd cmds.Command, options ...HandlerOption)`
   - Location: `pkg/glazed/handlers/output-file/output-file.go`

## 2. Handler Option Types (With Signatures)

### GenericCommandHandler Options

```go
// GenericCommandHandlerOption defines options for the GenericCommandHandler
type GenericCommandHandlerOption func(*GenericCommandHandler) error

// WithTemplateName sets the template name
func WithTemplateName(name string) GenericCommandHandlerOption

// WithDefaultTemplateName sets the template name if not already set
func WithDefaultTemplateName(name string) GenericCommandHandlerOption

// WithIndexTemplateName sets the index template name
func WithIndexTemplateName(name string) GenericCommandHandlerOption

// WithDefaultIndexTemplateName sets the index template name if not already set
func WithDefaultIndexTemplateName(name string) GenericCommandHandlerOption

// WithMergeAdditionalData adds data to the template context
func WithMergeAdditionalData(data map[string]interface{}, override bool) GenericCommandHandlerOption

// WithTemplateLookup sets the template lookup
func WithTemplateLookup(lookup render.TemplateLookup) GenericCommandHandlerOption

// WithParameterFilter sets the parameter filter
func WithParameterFilter(overridesAndDefaults *config.ParameterFilter) GenericCommandHandlerOption

// WithParameterFilterOptions creates a parameter filter from options
func WithParameterFilterOptions(opts ...config.ParameterFilterOption) GenericCommandHandlerOption

// WithPreMiddlewares adds middlewares to run before parameter filtering
func WithPreMiddlewares(middlewares ...middlewares.Middleware) GenericCommandHandlerOption

// WithPostMiddlewares adds middlewares to run after parameter filtering
func WithPostMiddlewares(middlewares ...middlewares.Middleware) GenericCommandHandlerOption

// WithWhitelistedLayers sets which parameter layers can be modified
func WithWhitelistedLayers(layers ...string) GenericCommandHandlerOption
```

### CommandHandler Options

```go
// CommandHandlerOption defines options for the CommandHandler
type CommandHandlerOption func(*CommandHandler) error

// WithDevMode enables development mode (template reloading)
func WithDevMode(devMode bool) CommandHandlerOption

// WithGenericCommandHandlerOptions passes options to the embedded GenericCommandHandler
func WithGenericCommandHandlerOptions(options ...generic_command.GenericCommandHandlerOption) CommandHandlerOption
```

### CommandDirHandler Options

```go
// CommandDirHandlerOption defines options for the CommandDirHandler
type CommandDirHandlerOption func(*CommandDirHandler) error

// WithDevMode enables development mode
func WithDevMode(devMode bool) CommandDirHandlerOption

// WithRepository sets the command repository
func WithRepository(r *repositories.Repository) CommandDirHandlerOption

// WithGenericCommandHandlerOptions passes options to the embedded GenericCommandHandler
func WithGenericCommandHandlerOptions(options ...generic_command.GenericCommandHandlerOption) CommandDirHandlerOption
```

### ParameterFilter Options

```go
// ParameterFilterOption defines options for the ParameterFilter
type ParameterFilterOption func(*ParameterFilter) error

// Override options
func WithReplaceOverrides(overrides *LayerParameters) ParameterFilterOption
func WithMergeOverrides(overrides *LayerParameters) ParameterFilterOption
func WithOverrideParameter(name string, value interface{}) ParameterFilterOption
func WithOverrideParameters(params map[string]interface{}) ParameterFilterOption
func WithMergeOverrideLayer(name string, layer map[string]interface{}) ParameterFilterOption
func WithReplaceOverrideLayer(name string, layer map[string]interface{}) ParameterFilterOption
func WithOverrideLayers(layers map[string]map[string]interface{}) ParameterFilterOption

// Default options
func WithReplaceDefaults(defaults *LayerParameters) ParameterFilterOption
func WithMergeDefaults(defaults *LayerParameters) ParameterFilterOption
func WithDefaultParameter(name string, value interface{}) ParameterFilterOption
func WithDefaultParameters(params map[string]interface{}) ParameterFilterOption
func WithMergeDefaultLayer(name string, layer map[string]interface{}) ParameterFilterOption
func WithReplaceDefaultLayer(name string, layer map[string]interface{}) ParameterFilterOption
func WithDefaultLayers(layers map[string]map[string]interface{}) ParameterFilterOption

// Whitelist options
func WithWhitelist(whitelist *ParameterFilterList) ParameterFilterOption
func WithWhitelistParameters(params ...string) ParameterFilterOption
func WithWhitelistLayers(layers ...string) ParameterFilterOption
func WithWhitelistLayerParameters(layer string, params ...string) ParameterFilterOption

// Blacklist options
func WithBlacklist(blacklist *ParameterFilterList) ParameterFilterOption
func WithBlacklistParameters(params ...string) ParameterFilterOption
func WithBlacklistLayers(layers ...string) ParameterFilterOption
func WithBlacklistLayerParameters(layer string, params ...string) ParameterFilterOption
```

## 3. Component Relationships and Inheritance

### Inheritance Hierarchy

```
GenericCommandHandler
  ├── CommandHandler
  └── CommandDirHandler
```

### Key Component Dependencies

1. **Handler → Server**: All handlers register with a Server via `Serve(server *parka.Server, basePath string) error`

2. **CommandHandler → GenericCommandHandler**: CommandHandler embeds GenericCommandHandler and passes options via `WithGenericCommandHandlerOptions`

3. **CommandDirHandler → GenericCommandHandler**: CommandDirHandler embeds GenericCommandHandler and passes options via `WithGenericCommandHandlerOptions`

4. **Handler → ParameterFilter**: Command handlers use ParameterFilter to control parameter behavior

5. **TemplateLookup → Renderer**: Renderers use TemplateLookup to find templates

## 4. Common Patterns

### 1. Functional Options Pattern

```go
// Constructor with options
func NewXHandler(options ...XHandlerOption) (*XHandler, error) {
    handler := &XHandler{
        // Default values
    }
    
    // Apply all options
    for _, option := range options {
        if err := option(handler); err != nil {
            return nil, err
        }
    }
    
    return handler, nil
}

// Option function type
type XHandlerOption func(*XHandler) error

// Option function implementations
func WithFeatureEnabled(enabled bool) XHandlerOption {
    return func(h *XHandler) error {
        h.featureEnabled = enabled
        return nil
    }
}
```

### 2. Handler Registration Pattern

```go
func (h *XHandler) Serve(server *server.Server, basePath string) error {
    // Register routes with server
    server.Router.GET(basePath, h.handleRequest)
    // ... more route registrations
    return nil
}
```

### 3. Embedded Handler Pattern

```go
// Parent embedding
type ChildHandler struct {
    // Embed parent
    ParentHandler
    
    // Child-specific fields
    additionalField string
}

// Pass options to parent
func WithParentOptions(options ...ParentOption) ChildOption {
    return func(h *ChildHandler) error {
        // Convert child handler to parent handler reference
        for _, opt := range options {
            if err := opt(&h.ParentHandler); err != nil {
                return err
            }
        }
        return nil
    }
}
```

### 4. Config-Based Construction

```go
func NewXHandlerFromConfig(config *config.XConfig, options ...XHandlerOption) (*XHandler, error) {
    // Convert config to options
    configOptions := []XHandlerOption{
        WithFeature(config.Feature),
        WithEnabled(config.Enabled),
        // ... more options from config
    }
    
    // Combine with passed options (passed options take precedence)
    allOptions := append(configOptions, options...)
    
    // Use regular constructor with all options
    return NewXHandler(allOptions...)
}
```

## 5. Parameter System Details

### Parameter Filter Structure

```go
// ParameterFilter controls parameter handling
type ParameterFilter struct {
    Overrides *LayerParameters      // Values that cannot be changed by user
    Defaults  *LayerParameters      // Values used if not provided by user
    Whitelist *ParameterFilterList  // Parameters that are allowed
    Blacklist *ParameterFilterList  // Parameters that are blocked
}

// LayerParameters holds parameter values
type LayerParameters struct {
    Parameters map[string]string                  // Top-level parameters
    Layers     map[string]map[string]interface{}  // Layer-specific parameters
}

// ParameterFilterList specifies which parameters to filter
type ParameterFilterList struct {
    Layers          []string             // Layers to filter
    LayerParameters map[string][]string  // Parameters to filter in layers
    Parameters      []string             // Top-level parameters to filter
}
```

### Middleware System Flow

```
Request
  ├── Pre-Middlewares (logging, validation, etc.)
  ├── Parameter Filter Middlewares
  │    ├── Apply defaults
  │    ├── Apply whitelist/blacklist
  │    └── Apply overrides
  ├── Post-Middlewares (custom processing)
  └── Command Execution
```

## 6. Repository System

### Repository Structure

```go
// Repository stores commands
type Repository struct {
    CommandMap   map[string]cmds.Command  // Commands by ID
    CommandCache []cmds.Command          // All commands (for listing)
}

// Creation options
func WithDirectories(dirs ...Directory) RepositoryOption
func WithCommandLoader(loader clay.CommandLoader) RepositoryOption
func WithUpdateCallback(callback UpdateCallback) RepositoryOption
```

### Directory Configuration

```go
// Directory configures where to load commands from
type Directory struct {
    FS               fs.FS    // Filesystem
    RootDirectory    string   // Root directory for commands
    RootDocDirectory string   // Root directory for documentation
    WatchDirectory   string   // Directory to watch for changes
    Name             string   // Repository name
    SourcePrefix     string   // Prefix for command source
}
```

## 7. Builder Pattern Implementation Strategy

Based on the codebase structure, here's how the builder pattern implementation should be approached:

### 1. Core Builder Classes

```go
// GenericCommandHandlerBuilder - base builder for all command handlers
type GenericCommandHandlerBuilder struct {
    // Configuration state
    templateName      string
    indexTemplateName string
    additionalData    map[string]interface{}
    templateLookup    render.TemplateLookup
    parameterFilter   *config.ParameterFilter
    preMiddlewares    []middlewares.Middleware
    postMiddlewares   []middlewares.Middleware
    whitelistedLayers []string
}

// Feature-specific builder
type ParameterConfigBuilder struct {
    parent *GenericCommandHandlerBuilder
    // Parameter configuration state
    defaults  map[string]interface{}
    overrides map[string]interface{}
    whitelist []string
    blacklist []string
    // ... more state
}
```

### 2. Builder Methods

```go
// Direct configuration
func (b *GenericCommandHandlerBuilder) WithTemplateName(name string) *GenericCommandHandlerBuilder {
    b.templateName = name
    return b
}

// Enter context-specific configuration
func (b *GenericCommandHandlerBuilder) ConfigureParameters() *ParameterConfigBuilder {
    return &ParameterConfigBuilder{
        parent: b,
        // Initialize state
    }
}

// Context-specific configuration methods
func (b *ParameterConfigBuilder) WithDefault(name string, value interface{}) *ParameterConfigBuilder {
    if b.defaults == nil {
        b.defaults = make(map[string]interface{})
    }
    b.defaults[name] = value
    return b
}

// Return to parent
func (b *ParameterConfigBuilder) Done() *GenericCommandHandlerBuilder {
    return b.parent
}

// Build the actual handler
func (b *GenericCommandHandlerBuilder) Build() (*generic_command.GenericCommandHandler, error) {
    // Convert builder state to options
    var options []generic_command.GenericCommandHandlerOption
    
    if b.templateName != "" {
        options = append(options, generic_command.WithTemplateName(b.templateName))
    }
    
    // ... convert other state to options
    
    // Create parameter filter if needed
    if b.parameterConfig != nil {
        filterOptions := []config.ParameterFilterOption{}
        
        if len(b.parameterConfig.defaults) > 0 {
            filterOptions = append(filterOptions, config.WithDefaultParameters(b.parameterConfig.defaults))
        }
        
        // ... convert other parameter config to options
        
        filter := config.NewParameterFilter(filterOptions...)
        options = append(options, generic_command.WithParameterFilter(filter))
    }
    
    // Create handler with options
    return generic_command.NewGenericCommandHandler(options...)
}
```

### 3. Preset Implementation

```go
// ParameterPreset encapsulates common parameter settings
type ParameterPreset struct {
    defaults  map[string]interface{}
    layerDefaults map[string]map[string]interface{}
    whitelist []string
    blacklist []string
    // ... more state
}

// Configuration methods
func (p *ParameterPreset) WithDefault(name string, value interface{}) *ParameterPreset {
    if p.defaults == nil {
        p.defaults = make(map[string]interface{})
    }
    p.defaults[name] = value
    return p
}

// Apply to a builder
func (p *ParameterPreset) ApplyTo(builder *ParameterConfigBuilder) *ParameterConfigBuilder {
    // Apply all settings to the builder
    for name, value := range p.defaults {
        builder.WithDefault(name, value)
    }
    
    if len(p.whitelist) > 0 {
        builder.WhitelistParameters(p.whitelist...)
    }
    
    // ... apply other settings
    
    return builder
}
```

## 8. Quick Start Functions

Example blueprint for quick start functions:

```go
// Create a simple API endpoint for a command
func QuickCommandAPI(
    cmd cmds.Command,    // Command to serve
    path string,         // Path to serve at
    server *Server,      // Server to register with
    preset *ParameterPreset, // Optional preset
) error {
    // Create builder
    builder := NewCommandHandlerBuilder().
        WithCommand(cmd)
    
    // Apply preset if provided
    if preset != nil {
        builder.ConfigureParameters().
            ApplyPreset(preset).
            Done()
    }
    
    // Build handler
    handler, err := builder.Build()
    if err != nil {
        return err
    }
    
    // Register with server
    return handler.Serve(server, path)
}
```

## Conclusion

The information provided in this document should be sufficient for implementing the builder pattern refactor proposed in REFACTOR-RFC.md. The refactor should:

1. Create builder classes for each handler type
2. Group configuration by functionality
3. Implement presets for common configurations
4. Provide quick start functions for common use cases
5. Maintain backward compatibility with the existing API

By following the patterns outlined in this document, the implementation will provide a more intuitive and discoverable API while preserving all the current functionality and flexibility of the Parka framework.