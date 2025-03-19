# Parka Architecture Documentation

## Overview

Parka is a web framework built on top of the Echo web framework that enables exposing command-line tools and APIs through a web interface. It provides a flexible and extensible architecture for serving static content, templates, and command execution through HTTP endpoints.

## Core Components

### 1. Server

The central component that manages the HTTP server and routing infrastructure. Built on Echo, it provides:

- HTTP request/response handling
- Route registration and routing
- Middleware processing
- Template rendering

```go
server, err := server.NewServer(
    server.WithPort(8080),
    server.WithAddress("localhost"),
    server.WithGzip(),
    server.WithDefaultParkaRenderer(),
)
```

### 2. Handlers

Handlers process different types of content and implement a common pattern for registering routes with the server:

- **StaticFileHandler**: Serves individual static files
- **StaticDirHandler**: Serves directories of static files
- **TemplateHandler**: Renders individual templates
- **TemplateDirHandler**: Renders templates from directories
- **CommandHandler**: Handles individual command execution
- **CommandDirHandler**: Handles multiple commands from repositories
- **ConfigFileHandler**: Configures multiple handlers from a YAML file

### 3. Rendering System

A flexible template rendering system that supports:

- Markdown rendering with syntax highlighting
- HTML template rendering
- Layout templates and partial inclusion
- Data context for templates

### 4. Parameter System

A sophisticated parameter handling system with:

- Parameter filtering (whitelist/blacklist)
- Default values
- Override values
- Middleware processing

### 5. Repository System

Manages command repositories for dynamic command discovery and execution:

- Directory-based command loading
- Dynamic command reloading
- Command discovery and execution

## Command Handler Registration

Parka offers multiple ways to register command handlers, from programmatic to configuration-based approaches.

### 1. Generic Command Handler

The base handler that provides core functionality:

```go
handler := generic_command.NewGenericCommandHandler(
    generic_command.WithTemplateName("command.tmpl.html"),
    generic_command.WithIndexTemplateName("index.tmpl.html"),
    generic_command.WithParameterFilter(filter),
    generic_command.WithWhitelistedLayers("default", "glazed"),
    generic_command.WithPreMiddlewares(middleware1, middleware2),
    generic_command.WithPostMiddlewares(middleware3),
)
```

**Available Options:**
- `WithTemplateName`: Sets the command template
- `WithIndexTemplateName`: Sets the index page template
- `WithMergeAdditionalData`: Adds data to pass to templates
- `WithTemplateLookup`: Configures template lookup
- `WithParameterFilter`: Configures parameter filtering
- `WithWhitelistedLayers`: Restricts parameter layers
- `WithPreMiddlewares`/`WithPostMiddlewares`: Adds middlewares

### 2. Single Command Handler

For serving individual Glazed commands:

```go
handler := command.NewCommandHandler(
    myCommand,
    command.WithDevMode(true),
    command.WithGenericCommandHandlerOptions(
        generic_command.WithTemplateName("command.tmpl.html"),
    ),
)

// Register with server
handler.Serve(server, "/api/my-command")
```

**Available Options:**
- `WithDevMode`: Enables development mode
- `WithGenericCommandHandlerOptions`: Passes options to the underlying GenericCommandHandler

### 3. Command Directory Handler

For serving multiple commands from a repository:

```go
repository := repositories.NewRepository(
    repositories.WithDirectories(repositories.Directory{
        FS:            os.DirFS(dir),
        RootDirectory: ".",
        WatchDirectory: dir,
    }),
)

handler := command_dir.NewCommandDirHandler(
    command_dir.WithRepository(repository),
    command_dir.WithDevMode(true),
    command_dir.WithGenericCommandHandlerOptions(
        generic_command.WithTemplateName("command.tmpl.html"),
        generic_command.WithIndexTemplateName("commands/index.tmpl.html"),
    ),
)

// Register with server
handler.Serve(server, "/api/commands")
```

**Available Options:**
- `WithRepository`: Sets the command repository
- `WithDevMode`: Enables development mode
- `WithGenericCommandHandlerOptions`: Passes options to the underlying GenericCommandHandler

### 4. Configuration-Based Registration

The most flexible way to register multiple handlers through a YAML configuration file:

```yaml
routes:
  # Command Directory
  - path: "/api"
    commandDirectory:
      repositories:
        - "./commands"
      includeDefaultRepositories: true
      templateName: "command.tmpl.html"
      indexTemplateName: "commands/index.tmpl.html"
      defaults:
        flags:
          limit: 100
        layers:
          glazed:
            filter:
              - id
              - name
      whitelist:
        - limit
        - offset
      blacklist:
        - debug
  
  # Single Command
  - path: "/hello"
    command:
      file: "./commands/hello.yaml"
      templateName: "command.tmpl.html"
```

Loading a configuration file:

```go
configData, err := os.ReadFile("config.yaml")
configFile, err := config.ParseConfig(configData)

cfh := handlers.NewConfigFileHandler(
    configFile,
    handlers.WithDevMode(true),
    handlers.WithRepositoryFactory(repositoryFactory),
)

// Register with server
cfh.Serve(server)
```

## Parameter Configuration

Parameter configuration is a powerful feature that allows control over how command parameters are processed:

### Parameter Filter

```go
filter := config.NewParameterFilter(
    // Default values
    config.WithDefaultParameter("limit", "100"),
    config.WithMergeDefaultLayer("glazed", map[string]interface{}{
        "filter": []string{"id", "name"},
    }),
    
    // Override values
    config.WithOverrideParameter("format", "json"),
    config.WithMergeOverrideLayer("sql-connection", map[string]interface{}{
        "host": "localhost",
        "port": 5432,
    }),
    
    // Whitelist allowed parameters
    config.WithWhitelistParameters("limit", "offset", "format"),
    config.WithWhitelistLayers("glazed", "sql-connection"),
    
    // Blacklist sensitive parameters
    config.WithBlacklistParameters("debug", "verbose"),
    config.WithBlacklistLayerParameters("sql-connection", "password"),
)
```

### Parameter Filter Options

1. **Defaults**:
   - `WithDefaultParameter(name, value)`: Set default for a parameter
   - `WithDefaultParameters(map)`: Set defaults for multiple parameters
   - `WithMergeDefaultLayer(name, map)`: Set defaults for a layer
   - `WithDefaultLayers(map)`: Set defaults for multiple layers

2. **Overrides**:
   - `WithOverrideParameter(name, value)`: Override a parameter
   - `WithOverrideParameters(map)`: Override multiple parameters
   - `WithMergeOverrideLayer(name, map)`: Override values in a layer
   - `WithOverrideLayers(map)`: Override values in multiple layers

3. **Whitelist**:
   - `WithWhitelistParameters(...string)`: Allow specific parameters
   - `WithWhitelistLayers(...string)`: Allow specific layers
   - `WithWhitelistLayerParameters(layer, ...string)`: Allow specific parameters in a layer

4. **Blacklist**:
   - `WithBlacklistParameters(...string)`: Block specific parameters
   - `WithBlacklistLayers(...string)`: Block specific layers
   - `WithBlacklistLayerParameters(layer, ...string)`: Block specific parameters in a layer

## Middleware System

Parka's middleware system operates at two levels:

### Echo Middleware

Standard Echo middlewares for HTTP processing:

```go
router := echo.New()
router.Use(middleware.Recover())
router.Use(middleware.RequestLoggerWithConfig(...))
router.Use(middleware.Gzip())
```

### Command Parameter Middleware

Specialized middleware for processing command parameters:

```go
handler := NewGenericCommandHandler(
    WithPreMiddlewares(
        LoggingMiddleware(),      // Run first
        ValidationMiddleware(),   // Run second
    ),
    WithParameterFilter(filter),  // Parameter filter middlewares run third
    WithPostMiddlewares(
        MetricsMiddleware(),      // Run fourth
        AuditMiddleware(),        // Run last
    ),
)
```

The middlewares follow a nested execution pattern:
```
PreMiddlewares(
    ParameterFilterMiddlewares(
        PostMiddlewares(
            actuaCommandExecution
        )
    )
)
```

## Request Flow

A typical request through Parka follows this path:

1. **Request Reception**: 
   - Echo router receives HTTP request
   - Echo middleware processes request (logging, recovery, etc.)

2. **Route Matching**:
   - Request is matched to registered route
   - Appropriate handler is invoked

3. **Parameter Processing**:
   - Request parameters are extracted from query, form, or JSON body
   - Pre-middlewares are executed
   - Parameter filtering is applied (defaults, overrides, whitelist, blacklist)
   - Post-middlewares are executed

4. **Command Execution** (for command handlers):
   - Command is retrieved from repository (for command directory handlers)
   - Command parameters are processed and validated
   - Command is executed with processed parameters
   - Command output is captured

5. **Response Generation**:
   - Based on requested format (JSON, text, HTML tables, etc.)
   - Template rendering if applicable

6. **Response Delivery**:
   - Response is sent back to client
   - Echo middleware finalizes response processing

## Best Practices

1. **Parameter Security**:
   - Use whitelists in production environments
   - Always blacklist sensitive parameters
   - Override security-critical settings

2. **Configuration Layering**:
   - Use defaults for development-friendly values
   - Apply environment-specific overrides
   - Keep security parameters separate from functional ones

3. **Middleware Ordering**:
   - Apply blacklists before whitelists
   - Set defaults before overrides
   - Consider execution order for custom middlewares

4. **Repository Management**:
   - Organize commands logically
   - Consider command discoverability
   - Use clear naming conventions