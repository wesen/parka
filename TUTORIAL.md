# Parka Builder Pattern Tutorial

## Introduction to the Builder Pattern in Parka

The Parka framework has been refactored to use the builder pattern to provide a more intuitive, discoverable, and flexible API. This tutorial will guide you through understanding and using the new builder pattern in your Parka applications.

### Understanding the Builder Pattern

The builder pattern is a design pattern that separates the construction of complex objects from their representation. In the context of Parka, this means:

1. **Method Chaining**: Configure handlers with chained method calls that return the builder itself
2. **Nested Builders**: Organize related configurations into logical groups using nested builders
3. **Configuration Presets**: Apply reusable configuration patterns with presets
4. **Clear API Surface**: Discover available options directly through IDE autocompletion

### Core Components of the Builder System

The Parka builder pattern consists of several key components:

1. **Builder Interfaces**: Core interfaces that define the builder pattern structure
   - `Builder`: Base interface for all builders
   - `ParameterConfigBuilder`: Interface for parameter configuration
   - `TemplateConfigBuilder`: Interface for template configuration
   - `MiddlewareConfigBuilder`: Interface for middleware configuration

2. **Handler Builders**: Implementation for different handler types
   - `GenericCommandHandlerBuilder`: Base builder for all command handlers
   - `CommandHandlerBuilder`: Builder for single command handlers
   - `CommandDirHandlerBuilder`: Builder for directory-based command handlers

3. **Helper Builders**: Additional builders for related components
   - `RepositoryBuilder`: Builder for repository configuration

4. **Preset System**: Reusable configuration patterns
   - `ParameterPreset`: Encapsulates parameter configuration settings
   - Standard presets for common use cases

5. **Quick Start Functions**: Simple API for common use cases
   - `QuickCommandAPI`: Rapidly configure API endpoints
   - `QuickCommandUI`: Rapidly configure web UIs
   - `QuickRepository`: Quickly create repository configurations

## Detailed Builder Pattern Usage

### Basic Handler Configuration

Let's start with the simplest use case - configuring a command handler:

```go
import (
    "github.com/wesen/parka/pkg/builders"
)

// Create a basic command handler
handler, err := builders.NewCommandHandlerBuilder().
    WithCommand(myCommand).           // Set the command to serve
    WithDevMode(true).                // Enable development mode
    WithTemplateName("command.tmpl.html"). // Set template for rendering
    Build()

if err != nil {
    // Handle error
}

// Register with your Echo server
handler.RegisterWith(server, "/api/my-command")
```

The method chaining pattern makes it clear what's being configured, and each method returns the builder itself to allow for further configuration.

### Using Nested Builders for Complex Configuration

For more complex configuration scenarios, Parka uses nested builders that group related settings:

```go
handler, err := builders.NewCommandDirHandlerBuilder().
    WithRepository(repository).
    WithDevMode(true).
    // Enter the template configuration context
    ConfigureTemplate().
        WithName("command.tmpl.html").
        WithIndexName("index.tmpl.html").
        WithData("title", "Command Explorer").
        Done(). // Return to parent builder
    // Enter the parameter configuration context
    ConfigureParameters().
        WithDefault("limit", "100").
        WithDefaults(map[string]interface{}{
            "format": "table",
            "quiet": "false",
        }).
        WithDefaultLayer("glazed", map[string]interface{}{
            "filter": []string{"id", "name"},
        }).
        WhitelistParameters("limit", "offset", "format").
        Done(). // Return to parent builder
    // Enter the middleware configuration context
    ConfigureMiddlewares().
        WithPreMiddleware(loggingMiddleware).
        WithPostMiddleware(analyticsMiddleware).
        Done(). // Return to parent builder
    Build()
```

The `Done()` method returns you to the parent builder, making it clear when you're exiting a configuration context. This approach:

1. Groups related settings together
2. Makes the configuration structure visually apparent
3. Prevents configuration options from becoming a flat, overwhelming list
4. Improves code readability and maintainability

### Parameter Filtering System

One of the most powerful features of Parka is its parameter filtering system, which allows you to control:

- Default parameter values
- Parameter overrides
- Whitelisted parameters (allowed)
- Blacklisted parameters (disallowed)
- Layer-specific configurations

The builder pattern makes this system more intuitive:

```go
builder.ConfigureParameters().
    // Set default parameter values
    WithDefault("limit", "100").
    WithDefault("offset", "0").
    
    // Set multiple defaults at once
    WithDefaults(map[string]interface{}{
        "format": "table",
        "quiet": "false",
    }).
    
    // Configure layer-specific defaults
    WithDefaultLayer("glazed", map[string]interface{}{
        "filter": []string{"id", "name"},
        "sort": "id",
    }).
    
    // Override parameters (takes precedence over user input)
    WithOverride("debug", "false").
    
    // Control which parameters are allowed
    WhitelistParameters("limit", "offset", "format", "filter").
    
    // Block sensitive parameters
    BlacklistParameters("password", "secret", "token").
    
    // Block parameters in specific layers
    BlacklistLayerParameters("database", "password", "user").
    BlacklistLayerParameters("aws", "secret_key").
    
    // Set whitelisted layers (only these layers are accessible)
    WhitelistLayers("default", "glazed", "format").
    
    Done() // Return to parent builder
```

### Using the Preset System

To promote reusability and consistency, the builder pattern includes a preset system that allows you to package configuration patterns for reuse:

```go
// Create a security-focused preset
securityPreset := builders.NewParameterPreset().
    BlacklistParameters("debug", "verbose").
    BlacklistLayerParameters("sql-connection", "password", "user").
    BlacklistLayerParameters("aws", "secret_key", "access_key").
    WhitelistParameters("limit", "offset", "format", "filter")

// Create a development preset
developmentPreset := builders.NewParameterPreset().
    WithDefault("limit", "100").
    WithDefaultLayer("sql-connection", map[string]interface{}{
        "host": "localhost",
        "port": 5432,
        "database": "dev",
    })

// Apply presets to a handler
handler, err := builders.NewCommandHandlerBuilder().
    WithCommand(myCommand).
    ConfigureParameters().
        // Apply presets (order matters - later presets can override earlier ones)
        ApplyPreset(securityPreset).
        ApplyPreset(developmentPreset).
        // Add specific overrides after presets
        WithDefault("limit", "50"). // Override the preset's default
        Done().
    Build()
```

The framework also provides standard presets for common scenarios:

```go
// Use a standard security preset
builder.ConfigureParameters().
    ApplyPreset(builders.StandardSecurityPreset).
    Done()

// Use a standard API preset
builder.ConfigureParameters().
    ApplyPreset(builders.StandardAPIPreset).
    Done()

// Use a standard Web UI preset
builder.ConfigureParameters().
    ApplyPreset(builders.StandardWebUIPreset).
    Done()
```

### Repository Configuration

For command directory handlers, you need to configure a repository. The builder pattern includes a dedicated builder for repositories:

```go
// Create a repository with builder pattern
repository, err := builders.NewRepositoryBuilder().
    // Add directories to search for command definitions
    AddDirectory("./commands").
    AddDirectory("./plugins").
    // Configure watching for changes
    WithWatchMode(true).
    // Build the repository
    Build()

if err != nil {
    // Handle error
}

// Use the repository with a command directory handler
handler, err := builders.NewCommandDirHandlerBuilder().
    WithRepository(repository).
    // Other configuration...
    Build()
```

### Quick Start Functions

For the most common use cases, the builder system provides quick start functions that handle the typical configuration for you:

```go
// Quick API endpoint for a single command
err := builders.QuickCommandAPI(
    myCommand,              // Command to serve
    "/api/reports",         // Base path
    server,                 // Echo server
    builders.StandardAPIPreset, // Optional preset
)

// Quick repository setup
repository, err := builders.QuickRepository("./commands")

// Quick web UI for browsing and executing commands
err := builders.QuickCommandUI(
    repository,             // Repository of commands
    "/commands",            // Base path
    server,                 // Echo server
    "command.tmpl.html",    // Template name
    builders.StandardWebUIPreset, // Optional preset
)
```

These functions internally use the builders but provide a simpler API for common scenarios.

## Real-World Use Cases

### Building an API Endpoint for a Single Command

```go
package main

import (
    "github.com/labstack/echo/v4"
    "github.com/wesen/parka/pkg/builders"
    "myapp/commands"
)

func main() {
    // Create an Echo server
    server := echo.New()
    
    // Create your command
    myCommand := &commands.GenerateReportCommand{}
    
    // Configure and register a command handler
    err := builders.NewCommandHandlerBuilder().
        WithCommand(myCommand).
        WithDevMode(false).
        ConfigureParameters().
            // Apply standard API preset
            ApplyPreset(builders.StandardAPIPreset).
            // Add specific configuration
            WithDefault("format", "json").
            WithDefault("limit", "100").
            WhitelistParameters("format", "limit", "offset", "filter").
            Done().
        // Register with server in one step
        RegisterWith(server, "/api/v1/reports")
    
    if err != nil {
        panic(err)
    }
    
    // Start server
    server.Start(":8080")
}
```

### Creating a Web UI for Command Exploration

```go
package main

import (
    "github.com/labstack/echo/v4"
    "github.com/wesen/parka/pkg/builders"
)

func main() {
    // Create an Echo server
    server := echo.New()
    
    // Create repository from commands directory
    repository, err := builders.NewRepositoryBuilder().
        AddDirectory("./commands").
        WithWatchMode(true). // Auto-reload when commands change
        Build()
    
    if err != nil {
        panic(err)
    }
    
    // Configure and register a command directory handler
    err = builders.NewCommandDirHandlerBuilder().
        WithRepository(repository).
        WithDevMode(true).
        ConfigureTemplate().
            WithName("command.tmpl.html").
            WithIndexName("commands/index.tmpl.html").
            WithData("title", "Command Explorer").
            WithData("version", "1.0.0").
            Done().
        ConfigureParameters().
            ApplyPreset(builders.StandardWebUIPreset).
            Done().
        RegisterWith(server, "/commands")
    
    if err != nil {
        panic(err)
    }
    
    // Start server
    server.Start(":8080")
}
```

### Customizing Parameter Filtering with Organization-Specific Presets

```go
package main

import (
    "github.com/labstack/echo/v4"
    "github.com/wesen/parka/pkg/builders"
    "os"
)

// Define organization-specific presets
func createOrganizationPresets() *builders.ParameterPreset {
    return builders.NewParameterPreset().
        // Security settings
        BlacklistParameters("debug", "verbose", "internal").
        BlacklistLayerParameters("database", "password", "user", "connection_string").
        BlacklistLayerParameters("aws", "secret_key", "access_key").
        // Default settings
        WithDefault("tenant", os.Getenv("DEFAULT_TENANT")).
        WithDefault("region", os.Getenv("DEFAULT_REGION")).
        // Layer defaults
        WithDefaultLayer("database", map[string]interface{}{
            "host": os.Getenv("DB_HOST"),
            "port": os.Getenv("DB_PORT"),
            "name": os.Getenv("DB_NAME"),
        })
}

func main() {
    // Create an Echo server
    server := echo.New()
    
    // Get organization presets
    orgPreset := createOrganizationPresets()
    
    // Create repository
    repository, _ := builders.QuickRepository("./commands")
    
    // Configure and register command directory handler
    builders.NewCommandDirHandlerBuilder().
        WithRepository(repository).
        ConfigureParameters().
            // Apply standard security preset first
            ApplyPreset(builders.StandardSecurityPreset).
            // Then apply organization preset (will override conflicting settings)
            ApplyPreset(orgPreset).
            Done().
        RegisterWith(server, "/commands")
    
    // Start server
    server.Start(":8080")
}
```

## Advanced Topics

### Combining Multiple Handlers

You can configure and combine multiple handlers in a single server:

```go
// Create Echo server
server := echo.New()

// Configure an API handler for a single command
builders.NewCommandHandlerBuilder().
    WithCommand(reportCommand).
    ConfigureParameters().
        ApplyPreset(builders.StandardAPIPreset).
        Done().
    RegisterWith(server, "/api/reports")

// Configure a web UI for exploring all commands
repository, _ := builders.QuickRepository("./commands")

builders.NewCommandDirHandlerBuilder().
    WithRepository(repository).
    ConfigureTemplate().
        WithName("command.tmpl.html").
        Done().
    RegisterWith(server, "/commands")

// Start server
server.Start(":8080")
```

### Custom Middleware Configuration

The builder pattern also makes it easy to configure custom middlewares:

```go
// Configure middlewares
builders.NewCommandHandlerBuilder().
    WithCommand(myCommand).
    ConfigureMiddlewares().
        // Add pre-middlewares (run before parameter processing)
        WithPreMiddleware(loggingMiddleware).
        WithPreMiddleware(authMiddleware).
        // Add parameter filters
        WithParameterFilter(customFilter).
        // Add post-middlewares (run after parameter processing)
        WithPostMiddleware(analyticsMiddleware).
        Done().
    RegisterWith(server, "/api/command")
```

### Error Handling in Builders

The builder pattern in Parka is designed to delay error checking until the `Build()` method is called. This allows for fluid method chaining without interruption:

```go
// Configure a handler with potential errors
handler, err := builders.NewCommandHandlerBuilder().
    WithCommand(myCommand).
    WithTemplateName("command.tmpl.html").
    Build()

// Check for errors after building
if err != nil {
    log.Fatalf("Error building handler: %v", err)
}
```

For direct registration with a server, the error is returned from the `RegisterWith` method:

```go
// Build and register in one step
err := builders.NewCommandHandlerBuilder().
    WithCommand(myCommand).
    RegisterWith(server, "/api/command")

if err != nil {
    log.Fatalf("Error registering handler: %v", err)
}
```

## Best Practices

### Organizing Builder Configuration

1. **Logical Grouping**: Use nested builders to group related configuration
2. **Preset First, Specifics Later**: Apply presets first, then add specific overrides
3. **Clear Context Boundaries**: Use `Done()` to clearly mark the end of a context

### Reusing Configuration Patterns

1. **Create Organization Presets**: Package common settings into reusable presets
2. **Use Standard Presets**: Leverage the built-in presets for common scenarios
3. **Document Custom Presets**: When creating custom presets, document their purpose and effect

### Error Handling

1. **Check Build Errors**: Always check the error returned from `Build()`
2. **Validate Early**: Perform basic validation before building when possible
3. **Use Meaningful Error Messages**: When errors occur, provide clear context

## Migrating from Functional Options

If you're migrating from the previous functional options pattern, here's how the patterns map to each other:

### Command Handler Example

**Previous Approach:**
```go
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
```

**New Builder Approach:**
```go
handler, err := builders.NewCommandHandlerBuilder().
    WithCommand(cmd).
    ConfigureParameters().
        WithDefault("limit", "100").
        WhitelistParameters("limit", "offset", "format").
        Done().
    Build()
```

### Command Directory Handler Example

**Previous Approach:**
```go
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
```

**New Builder Approach:**
```go
handler, err := builders.NewCommandDirHandlerBuilder().
    WithRepository(repository).
    WithDevMode(true).
    WithTemplateName("command.tmpl.html").
    WithIndexTemplateName("commands/index.tmpl.html").
    ConfigureParameters().
        WithDefault("limit", "100").
        WhitelistParameters("limit", "offset", "format").
        Done().
    Build()
```

## Conclusion

The Parka builder pattern provides a more intuitive, discoverable, and flexible API for configuring Parka handlers. By using method chaining, nested builders, and presets, you can create more readable and maintainable code while reducing cognitive load and improving developer productivity.

This tutorial has covered the basics of using the builder pattern, from simple handler configuration to advanced topics like custom presets and middleware configuration. As you continue to work with Parka, you'll discover that the builder pattern makes even complex configurations more manageable and intuitive.