package builders_test

// Package builders_test demonstrates how to use the Parka builders
// These are simplified examples and don't compile as real tests.
// They serve as documentation examples only.

/*
ExampleGenericCommandHandlerBuilder shows how to configure a generic command handler.

Example usage:

```go
builder := builders.NewGenericCommandHandlerBuilder().
	WithTemplateName("command.tmpl.html").
	WithIndexTemplateName("index.tmpl.html")

builder.ConfigureParameters().
	WithDefault("limit", "100").
	WithDefaults(map[string]interface{}{
		"format": "table",
	}).
	WithDefaultLayer("glazed", map[string]interface{}{
		"filter": []string{"id", "name"},
	}).
	WhitelistParameters("limit", "offset", "format").
	Done()

handler, err := builder.Build()
```
*/
func ExampleGenericCommandHandlerBuilder() {
	// Output:
}

/*
ExampleCommandHandlerBuilder shows how to configure a command handler.

Example usage:

```go
builder := builders.NewCommandHandlerBuilder().
	WithCommand(myCommand).
	WithDevMode(true).
	WithTemplateName("command.tmpl.html")

builder.ConfigureParameters().
	WithDefault("limit", "100").
	WhitelistParameters("limit", "offset", "format").
	Done()

builder.ConfigureTemplate().
	WithName("command.tmpl.html").
	WithIndexName("index.tmpl.html").
	WithData("title", "My Command").
	Done()

handler, err := builder.Build()
```
*/
func ExampleCommandHandlerBuilder() {
	// Output:
}

/*
ExampleCommandDirHandlerBuilder shows how to configure a command directory handler.

Example usage:

```go
repository := buildRepository()

builder := builders.NewCommandDirHandlerBuilder().
	WithRepository(repository).
	WithDevMode(true)

builder.ConfigureParameters().
	ApplyPreset(builders.StandardSecurityPreset).
	WithDefault("limit", "50").
	Done()

handler, err := builder.Build()
handler.Serve(server, "/commands")
```
*/
func ExampleCommandDirHandlerBuilder() {
	// Output:
}

/*
ExampleQuickCommandAPI shows how to quickly create an API endpoint.

Example usage:

```go
// Create a simple API endpoint
err := builders.QuickCommandAPI(myCommand, "/api/reports", server, builders.StandardAPIPreset)
```
*/
func ExampleQuickCommandAPI() {
	// Output:
}

/*
ExampleQuickCommandUI shows how to quickly create a web UI.

Example usage:

```go
// Create a web UI for browsing commands
repository, _ := builders.QuickRepository("./commands")
err := builders.QuickCommandUI(repository, "/commands", server, "command.tmpl.html", nil)
```
*/
func ExampleQuickCommandUI() {
	// Output:
}

/*
ExampleCustomParameterPreset shows how to create and use custom presets.

Example usage:

```go
// Create custom presets
securityPreset := builders.NewParameterPreset().
	BlacklistParameters("debug", "verbose").
	BlacklistLayerParameters("sql-connection", "password")

databasePreset := builders.NewParameterPreset().
	WithDefaultLayer("sql-connection", map[string]interface{}{
		"host": os.Getenv("DB_HOST"),
		"port": os.Getenv("DB_PORT"),
		"database": "reports",
	})

// Apply presets to handler
handler, _ := builders.NewCommandHandlerBuilder().
	WithCommand(myCommand).
	ConfigureParameters().
		ApplyPreset(securityPreset).
		ApplyPreset(databasePreset).
		WithDefault("limit", "50"). // Override preset
		Done().
	Build()
```
*/
func ExampleCustomParameterPreset() {
	// Output:
}