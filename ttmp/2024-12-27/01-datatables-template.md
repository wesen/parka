# Data Tables Template System Documentation

## Core Types

### Command Description
The `CommandDescription` struct is the central type that describes a command and its parameters:

```go
type CommandDescription struct {
    Name           string
    Short          string 
    Long           string
    Layout         []*layout.Section
    Layers         *layers.ParameterLayers
    AdditionalData map[string]interface{}
    Parents        []string
    Source         string
}
```

### Layout System
The layout system uses several nested types to define form structure:

```go
type Layout struct {
    Sections []*Section
}

type Section struct {
    Title       string
    Description string 
    Rows        []*Row
    Style       string
    Classes     string
}

type Row struct {
    Inputs  []*Input
    Classes string
    Style   string 
}

type Input struct {
    Name         string
    Label        string
    Options      []Option
    DefaultValue interface{}
    Help         string
    CSS         string
    Id          string
    Classes     string
    Template    string
    InputType   string
}
```

### Parameter System
Parameters are defined through:

```go
type ParameterDefinition struct {
    Name       string
    ShortFlag  string
    Type       ParameterType
    Help       string
    Default    *interface{}
    Choices    []string
    Required   bool
    IsArgument bool
}
```

Parameters are organized into layers through the `ParameterLayer` interface and `ParameterLayers` collection.

## Layout Computation

The layout is computed in `ComputeLayout()` which:

1. Takes a command and parsed layers as input
2. Processes parameter definitions from all layers
3. Creates sections, rows and inputs based on the parameters
4. Returns a complete `Layout` struct

## Template Rendering Flow

### Template Context Setup

The `QueryHandler.renderTemplate()` method sets up the template context by:

1. Looking up the template
2. Computing the layout
3. Rendering markdown description to HTML
4. Setting up data streams for rows
5. Passing the complete DataTables struct to the template

### Form Widget Template

The form widget template (`form-widget`) handles rendering different parameter types:

```html
{{ define "form-widget" }}
<div {{ if .Id }}id="{{.Id}}" {{end}}
     {{ if .Classes }}class="{{.Classes}}" {{end}}
     {{ if .CSS }}style="{{.CSS}}"{{end}}
     style="height: 100%">
    <div style="display: flex; flex-direction: column; align-items: end; height: 100%">
        <!-- Parameter type specific rendering -->
        {{ if eq .Type "string" }}
            <label for="{{.Name}}">{{.Help}}</label>
            <input type="text" name="{{.Name}}" value="{{.Value}}">
            
        {{ else if eq .Type "stringList" }}
            <label for="{{.Name}}">{{.Help}}</label>
            <input type="text" name="{{.Name}}" value='{{.Value | join "," }}'>
            
        <!-- Additional types... -->
        
        {{ end }}
    </div>
</div>
{{ end }}
```

The template supports these parameter types:

- string - Text input
- stringList - Comma-separated text input
- int/float - Number input
- bool - Checkbox
- date - Date picker
- choice - Select dropdown
- choiceList - Multi-select dropdown

### Form Layout Rendering

The main template renders the form structure:

1. Command header and description
2. Iterates through layout sections
3. For each section:
   - Renders section title/description
   - Iterates through rows
   - For each row renders inputs using form-widget template

```html
<form id="form" action="{{.Command.Name}}" method="get">
    <fieldset>
        {{range $section := .Layout.Sections}}
            <!-- Section header -->
            <div class="row">
                <div class="columns">
                    {{if $section.Title}}<h3>{{$section.Title}}</h3>{{end}}
                    {{if $section.ShortDescription}}
                        <p>{{$section.ShortDescription}}</p>
                    {{end}}
                </div>
            </div>
            
            <!-- Section rows -->
            {{range $row := $section.Rows}}
                <div class="row">
                    {{range $field := $row.Inputs}}
                        <div class="column">
                            {{ template "form-widget" $field }}
                        </div>
                    {{end}}
                </div>
            {{end}}
        {{end}}
    </fieldset>
</form>
```

### JavaScript Handling

The template includes JavaScript for:

1. Form submission handling
2. Query string management
3. Form validation
4. DataTables initialization (when enabled)
5. Error handling and display

## Key Features

1. **Flexible Layout**: The section/row/input structure allows flexible form layouts

2. **Parameter Type Support**: Comprehensive support for different parameter types

3. **Validation**: Built-in form validation

4. **Streaming**: Supports streaming results through channels

5. **Error Handling**: Structured error display and handling

6. **DataTables Integration**: Optional DataTables.net integration for results

7. **Download Links**: Automatic generation of download links for different formats

## Usage Example

To use the template system:

1. Create a CommandDescription with parameters
2. Define layout sections if custom layout needed
3. Create QueryHandler with the command
4. Pass to HTTP handler

The template will automatically render an appropriate form based on the parameters and layout.
