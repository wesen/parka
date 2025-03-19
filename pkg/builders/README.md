# Parka Builder Pattern

This package provides a more intuitive and discoverable API for configuring Parka handlers using the builder pattern.

## Overview

The builder pattern implementation offers:

1. **Method Chaining**: Discoverable configuration through method chaining
2. **Contextual Configuration**: Nested builders for logical grouping of related settings
3. **Configuration Presets**: Reusable configuration patterns
4. **Quick Start Functions**: Simple constructors for common use cases

## Builder Types

The package provides the following builder types:

- `GenericCommandHandlerBuilder`: Base builder for generic command handlers
- `CommandHandlerBuilder`: Builder for single command handlers
- `CommandDirHandlerBuilder`: Builder for directory-based command handlers
- `RepositoryBuilder`: Helper builder for repository configuration

## Usage Examples

### Basic Usage

```go
// Create a command handler with chained configuration
handler, err := NewCommandHandlerBuilder().
    WithCommand(myCommand).
    WithDevMode(true).
    WithTemplateName("command.tmpl.html").
    Build()
```

### Contextual Configuration

```go
// Use nested builders for contextual configuration
handler, err := NewCommandDirHandlerBuilder().
    WithRepository(repository).
    WithDevMode(true).
    ConfigureTemplate().
        WithName("command.tmpl.html").
        WithIndexName("index.tmpl.html").
        Done().
    ConfigureParameters().
        WithDefault("limit", "100").
        WhitelistParameters("limit", "offset", "format").
        Done().
    Build()
```

### Using Presets

```go
// Create a custom preset
securityPreset := NewParameterPreset().
    BlacklistParameters("debug", "verbose").
    BlacklistLayerParameters("sql-connection", "password")

// Apply preset to a handler
handler, err := NewCommandHandlerBuilder().
    WithCommand(myCommand).
    ConfigureParameters().
        ApplyPreset(securityPreset).
        WithDefault("limit", "50"). // Override preset
        Done().
    Build()
```

### Quick Start Functions

```go
// Create a simple API endpoint
err := QuickCommandAPI(myCommand, "/api/reports", server, StandardAPIPreset)

// Create a web UI for browsing commands
repository, _ := QuickRepository("./commands")
err := QuickCommandUI(repository, "/commands", server, "command.tmpl.html", StandardWebUIPreset)
```

## Standard Presets

The package includes several standard presets:

- `StandardSecurityPreset`: Security-focused parameter filters
- `StandardAPIPreset`: Settings optimized for API endpoints
- `StandardWebUIPreset`: Settings optimized for web UI endpoints

## Backward Compatibility

The builder pattern implementation maintains backward compatibility with the existing functional options pattern. Builders internally convert their configuration to the equivalent functional options when building handlers.

## Registration

All builders provide a convenient `RegisterWith` method to register the handler with a server:

```go
// Build and register in one step
err := NewCommandHandlerBuilder().
    WithCommand(myCommand).
    RegisterWith(server, "/api/reports")
```

## Repository Builder

The package also includes a builder for repository configuration:

```go
// Create a repository builder
repository, err := NewRepositoryBuilder().
    AddDirectory("./commands").
    Build()
```

## Benefits

- **Reduced Cognitive Load**: Focus on what you want to configure, not how
- **Improved Discoverability**: Method chaining provides natural discovery
- **Contextual Configuration**: Logical grouping of related options
- **Code Reuse**: Presets enable easy reuse of common patterns
- **Simplified Configuration**: Quick start functions for common scenarios