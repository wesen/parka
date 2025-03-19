# Parka Framework Refactoring RFC

## Current State Analysis

### 1. Complexity Assessment

The Parka framework, while powerful and flexible, presents a significant complexity burden for developers. The current implementation relies heavily on the functional options pattern with numerous `With*` options that create several challenges:

1. **Cognitive Overload**: Developers must understand and remember dozens of option functions across multiple handler types.

2. **Discoverability Issues**: It's difficult to discover which options are available without extensive documentation or IDE support.

3. **Composability Challenges**: The nesting of options (e.g., `WithGenericCommandHandlerOptions`) creates multiple layers that are hard to trace mentally.

4. **Configuration Duplication**: Common configuration patterns need to be reapplied across multiple handlers.

5. **Steep Learning Curve**: New developers face a significant learning curve to understand the available options and their interactions.

Example of current complexity:

```go
handler := command_dir.NewCommandDirHandler(
    command_dir.WithRepository(repository),
    command_dir.WithDevMode(true),
    command_dir.WithGenericCommandHandlerOptions(
        generic_command.WithTemplateName("command.tmpl.html"),
        generic_command.WithIndexTemplateName("index.tmpl.html"),
        generic_command.WithParameterFilter(
            config.NewParameterFilter(
                config.WithDefaultParameter("limit", "100"),
                config.WithMergeDefaultLayer("glazed", map[string]interface{}{
                    "filter": []string{"id", "name"},
                }),
                config.WithWhitelistParameters("limit", "offset", "format"),
                config.WithBlacklistLayerParameters("sql-connection", "password"),
            ),
        ),
        generic_command.WithWhitelistedLayers("default", "glazed"),
        generic_command.WithPreMiddlewares(middleware1, middleware2),
    ),
)
```

This pattern, while flexible, creates code that is difficult to read, maintain, and understand at a glance.

### 2. Documentation Burden

The current approach creates a heavy documentation burden:

- Each option function needs its own documentation
- The interactions between options must be documented
- Common patterns and best practices need extensive examples
- Configuration examples balloon in size and complexity

### 3. Architectural Assessment

The core issue is not with the architecture itself, which follows solid principles, but rather with its exposed API surface. The internal components are well-structured, but the developer-facing API introduces unnecessary complexity through its heavily nested option pattern.

## Proposed Solution: Builder Pattern and Presets

I propose a two-pronged approach to simplify Parka's API while maintaining its power and flexibility:

### 1. Builder Pattern Implementation

Replace the complex functional options pattern with a more intuitive builder pattern that:

- Provides discoverable, method-chained configuration
- Separates concerns into logical groupings
- Offers sensible defaults
- Self-documents through method names and fluent API

Example of proposed builder pattern:

```go
handler := NewCommandDirHandlerBuilder().
    WithRepository(repository).
    WithDevMode(true).
    ConfigureTemplate().
        WithName("command.tmpl.html").
        WithIndexName("index.tmpl.html").
        Done().
    ConfigureParameters().
        WithDefault("limit", "100").
        WithDefaults(map[string]interface{}{
            "format": "table",
        }).
        WithDefaultLayer("glazed", map[string]interface{}{
            "filter": []string{"id", "name"},
        }).
        WhitelistParameters("limit", "offset", "format").
        Done().
    ConfigureMiddlewares().
        WithPreMiddleware(middleware1).
        WithPreMiddleware(middleware2).
        Done().
    Build()
```

Benefits:
- Method chaining provides a natural, discoverable API
- Nested configuration groups logically organize options
- `Done()` methods clearly delineate configuration sections
- Type safety and IDE completion improve usability
- Clearer mental model of the configuration structure

### 2. Configuration Presets

Introduce a preset system that allows common configurations to be packaged and reused:

```go
// Define standard presets
securePreset := NewParameterPreset().
    BlacklistParameters("debug", "verbose").
    BlacklistLayerParameters("sql-connection", "password").
    WhitelistParameters("limit", "offset", "format")

developmentPreset := NewParameterPreset().
    WithDefault("limit", "100").
    WithDefaultLayer("sql-connection", map[string]interface{}{
        "host": "localhost",
        "port": 5432,
    })

// Use presets in handler configuration
handler := NewCommandDirHandlerBuilder().
    WithRepository(repository).
    ConfigureParameters().
        ApplyPreset(securePreset).
        ApplyPreset(developmentPreset).
        Done().
    Build()
```

Benefits:
- Encapsulates common configuration patterns
- Promotes consistency across handlers
- Simplifies configuration code
- Makes best practices easier to follow
- Enables organization-specific presets

### 3. Quick Start Configuration

Introduce "quick start" constructors for common use cases:

```go
// Simple API endpoint for a command
handler := QuickCommandAPI(
    myCommand,             // Command to serve
    "/api/v1/reports",     // Base path
    StandardSecurityPreset // Standard security preset
)

// Web UI for browsing and executing commands
handler := QuickCommandUI(
    repository,            // Repository of commands
    "/commands",           // Base path
    "command.tmpl.html",   // Template name
    StandardWebUIPreset    // Standard web UI preset
)
```

Benefits:
- Gets developers running quickly with minimal code
- Promotes standard patterns
- Can be extended with the builder pattern when needed
- Significantly reduces the learning curve

## Implementation Strategy

### 1. Builder Classes Implementation

Create builder classes for each handler type that encapsulate the functional options:

```go
type CommandDirHandlerBuilder struct {
    repository *repositories.Repository
    devMode    bool
    template   *TemplateConfig
    parameters *ParameterConfig
    middlewares *MiddlewareConfig
}

func (b *CommandDirHandlerBuilder) WithRepository(repo *repositories.Repository) *CommandDirHandlerBuilder {
    b.repository = repo
    return b
}

func (b *CommandDirHandlerBuilder) WithDevMode(devMode bool) *CommandDirHandlerBuilder {
    b.devMode = devMode
    return b
}

func (b *CommandDirHandlerBuilder) ConfigureTemplate() *TemplateConfigBuilder {
    if b.template == nil {
        b.template = &TemplateConfig{}
    }
    return &TemplateConfigBuilder{
        parent: b,
        config: b.template,
    }
}

// ... other methods

func (b *CommandDirHandlerBuilder) Build() (*CommandDirHandler, error) {
    // Convert builder config to functional options
    var options []command_dir.Option
    
    if b.repository != nil {
        options = append(options, command_dir.WithRepository(b.repository))
    }
    
    if b.devMode {
        options = append(options, command_dir.WithDevMode(true))
    }
    
    // Convert template config to options...
    // Convert parameter config to options...
    // Convert middleware config to options...
    
    return command_dir.NewCommandDirHandler(options...)
}
```

### 2. Preset Implementation

Create a preset system that can generate and combine options:

```go
type ParameterPreset struct {
    defaults     map[string]interface{}
    layerDefaults map[string]map[string]interface{}
    overrides    map[string]interface{}
    layerOverrides map[string]map[string]interface{}
    whitelist    []string
    blacklist    []string
    // ...
}

func (p *ParameterPreset) WithDefault(name string, value interface{}) *ParameterPreset {
    if p.defaults == nil {
        p.defaults = make(map[string]interface{})
    }
    p.defaults[name] = value
    return p
}

// ... other methods

func (p *ParameterPreset) ToParameterFilter() *config.ParameterFilter {
    options := []config.ParameterFilterOption{}
    
    if p.defaults != nil {
        options = append(options, config.WithDefaultParameters(p.defaults))
    }
    
    // ... convert other preset properties to options
    
    return config.NewParameterFilter(options...)
}
```

### 3. Migration Strategy

1. **Backward Compatibility**: Implement the builder pattern alongside the existing functional options to ensure a smooth transition.

2. **Documentation Update**: Create comprehensive documentation for the new API with examples of common patterns.

3. **Gradual Migration**: Encourage developers to migrate to the new API while maintaining support for the old API.

4. **New Features**: Implement new features using the new API pattern only.

## Code Examples

### Example 1: Simple Command API

```go
// Current approach
cmd := &MyReportCommand{}
handler, err := command.NewCommandHandler(
    cmd,
    command.WithGenericCommandHandlerOptions(
        generic_command.WithParameterFilter(
            config.NewParameterFilter(
                config.WithDefaultParameter("limit", "100"),
                config.WithWhitelistParameters("limit", "offset", "format"),
            ),
        ),
    ),
)
handler.Serve(server, "/api/reports")

// Proposed approach
handler := QuickCommandAPI(cmd, "/api/reports", StandardAPIPreset)
handler.RegisterWith(server)

// Or with builder for more control
handler := NewCommandHandlerBuilder().
    WithCommand(cmd).
    ConfigureParameters().
        WithDefault("limit", "100").
        WhitelistParameters("limit", "offset", "format").
        Done().
    Build()
handler.RegisterWith(server, "/api/reports")
```

### Example 2: Command Directory with Web UI

```go
// Current approach
repository := repositories.NewRepository(
    repositories.WithDirectories(repositories.Directory{
        FS:            os.DirFS(dir),
        RootDirectory: ".",
        WatchDirectory: dir,
    }),
)
handler, err := command_dir.NewCommandDirHandler(
    command_dir.WithRepository(repository),
    command_dir.WithDevMode(true),
    command_dir.WithGenericCommandHandlerOptions(
        generic_command.WithTemplateName("command.tmpl.html"),
        generic_command.WithIndexTemplateName("commands/index.tmpl.html"),
        generic_command.WithParameterFilter(
            config.NewParameterFilter(
                config.WithDefaultParameter("limit", "100"),
                config.WithWhitelistParameters("limit", "offset", "format"),
            ),
        ),
    ),
)
handler.Serve(server, "/commands")

// Proposed approach
repository := QuickRepository(dir)
handler := QuickCommandUI(repository, "/commands", "command.tmpl.html", StandardWebUIPreset)
handler.RegisterWith(server)

// Or with builder for more control
repository := NewRepositoryBuilder().
    AddDirectory(dir).
    Build()

handler := NewCommandDirHandlerBuilder().
    WithRepository(repository).
    WithDevMode(true).
    ConfigureTemplate().
        WithName("command.tmpl.html").
        WithIndexName("commands/index.tmpl.html").
        Done().
    ConfigureParameters().
        ApplyPreset(StandardSecurityPreset).
        WithDefault("limit", "100").
        Done().
    Build()
handler.RegisterWith(server, "/commands")
```

### Example 3: Custom Configuration with Presets

```go
// Create custom presets
securityPreset := NewParameterPreset().
    BlacklistParameters("debug", "verbose").
    BlacklistLayerParameters("sql-connection", "password").
    WhitelistParameters("limit", "offset", "format")

databasePreset := NewParameterPreset().
    WithDefaultLayer("sql-connection", map[string]interface{}{
        "host": os.Getenv("DB_HOST"),
        "port": os.Getenv("DB_PORT"),
        "database": "reports",
    })

// Apply presets to handler
handler := NewCommandHandlerBuilder().
    WithCommand(reportCommand).
    ConfigureParameters().
        ApplyPreset(securityPreset).
        ApplyPreset(databasePreset).
        // Override or add specific settings
        WithDefault("limit", "50").
        Done().
    Build()
```

## Benefits of the Proposed Approach

1. **Reduced Cognitive Load**: Developers can focus on what they want to configure rather than how to configure it.

2. **Improved Discoverability**: Method chaining provides natural discovery of available options.

3. **Contextual Configuration**: Nested builders provide logical grouping of related configuration.

4. **Code Reuse**: Presets enable easy reuse of common configuration patterns.

5. **Simplified Documentation**: Documentation can focus on high-level concepts rather than individual options.

6. **Faster Onboarding**: New developers can be productive with minimal learning.

7. **Maintainability**: Code is more maintainable with clear structure and organization.

## Conclusion

The current Parka framework provides powerful functionality but suffers from an unnecessarily complex API surface. By implementing a builder pattern with logical grouping, presets for common configurations, and quick start constructors, we can maintain all the power and flexibility while significantly improving the developer experience.

This refactoring proposal aims to:

1. Simplify the API for common use cases
2. Maintain full flexibility for complex scenarios
3. Improve code readability and maintainability
4. Reduce the learning curve for new developers

The proposed changes should be implemented in a backward-compatible way to ensure existing code continues to work while encouraging migration to the new, more user-friendly API.