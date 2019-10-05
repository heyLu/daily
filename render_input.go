package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

func RenderInput(w http.ResponseWriter, req *http.Request, typeName string) {
	data := map[string]interface{}{ "Title": "New entry - daily",
		"Type": typeName,
	}
	tmpl := tmplInputDefault

	if customTmpl, ok := inputTemplateFor[typeName]; ok {
		tmpl = customTmpl.Template
		for key, val := range customTmpl.Data {
			data[key] = val
		}
	}

	err := tmpl.Execute(w, data)
	if err != nil {
		log.Println(err)
		fmt.Fprint(w, err)
	}
}

var inputTemplateFor = map[string]templateDefinition{
	"mood": templateDefinition{
		Template: tmplInputMood,
		Data: map[string]interface{}{
			"Title": "mood - daily",
			"Stylesheet": "mood.css",
			"ValueLabel": "Mood",
		},
	},
}

type templateDefinition struct {
	Template *template.Template
	Data     map[string]interface{}
}

var tmplInputMood = template.Must(tmplBase.New("input-mood").Parse(`
{{- template "html-start" . }}
	<div id="mood-gradient" class="hidden"></div>

	{{ template "input-form" . }}

	<script defer async src="/static/mood.js"></script>
{{ template "html-end" }}
`))

var tmplInputDefault = template.Must(tmplBase.New("input-default").Parse(`
{{- template "html-start" . }}
	{{ template "input-form" . }}
{{ template "html-end" }}
`))

func RenderEdit(w http.ResponseWriter, req *http.Request, entry *Entry) {
	data := map[string]interface{}{
		"Title": "Edit entry - daily",
		"Entry": entry,
	}

	err := tmplEditDefault.Execute(w, data)
	if err != nil {
		log.Println(err)
		fmt.Fprint(w, err)
	}
}

var tmplEditDefault = template.Must(tmplBase.New("edit-default").Parse(`
{{- template "html-start" . }}
	{{ template "edit-form" . }}
{{ template "html-end" }}
`))

var tmplBase = template.Must(template.New("base").Funcs(tmplFuncs).Parse(`{{ define "input-form" }}
	<section id="content">
		<h1>Create entry</h1>

		<form method="POST" action="/new">
			<input name="type" value="{{ .Type }}" placeholder="type" required {{ if .Type }}hidden{{ end }} />
			<div class="field">
				<label for="value">{{ or .ValueLabel "Value" }}</label>
				<input id="entry-value" name="value" type="range" min="0" max="1" step="0.01" />
			</div>

			<div class="field">
				<label for="Note">Note</label>
				<input name="note" type="text" />
			</div>

			<h2>Additional data</h2>

			<div id="additional-fields"></div>

			<div class="field">
				<button id="add-field">Add field</button>
			</div>

			<input type="submit" value="Save" />
		</form>
	</section>
{{ end }}
{{ define "edit-form" }}
	<section id="content">
		<h1>Edit entry</h1>

		<form method="POST" action="/edit/{{ .Entry.ID }}">
			<div class="field">
				<input name="type" value="{{ .Entry.Type }}" disabled />
			</div>

			<div class="field">
				<label for="value">Value</label>
				<input id="entry-value" name="value" type="number" value="{{ .Entry.Value }}" />
			</div>

			<div class="field">
				<label for="note">Note</label>
				<input name="note" type="text" value="{{ .Entry.Note }}" />
			</div>

			<h2>Additional data</h2>

			<div id="additional-fields">
			{{ range $key, $value := .Entry.Data }}
				{{ if (isList $value) }}
					{{ range $singleValue := $value }}
					<div class="field">
						<input class="field-key" type="text" value="{{ $key }}" />
						<input class="field-value" type="text" name="{{ $key }}" value="{{ $singleValue }}" />
					</div>
					{{ end }}
				{{ else }}
				<div class="field">
					<input class="field-key" type="text" value="{{ $key }}" />
					<input class="field-value" type="text" name="{{ $key }}" value="{{ $value }}" />
				</div>
				{{ end }}
			{{ end }}
			</div>

			<div class="field">
				<button id="add-field">Add field</button>
			</div>

			<input type="submit" value="Save" />
		</form>
	</section>
{{ end }}

{{ define "html-start" }}
<!doctype html>
<html>
<head>
	<title>{{ .Title }}</title>
	<meta name="viewport" content="width=device-width, initial-scale=1" />

	<link rel="stylesheet" href="/static/default.css" />
	{{ if .Stylesheet }}
	<link rel="stylesheet" href="/static/{{ .Stylesheet }}" />
	{{ end }}

	<script defer async src="/static/fields.js"></script>
</head>

<body>
{{ end }}

{{ define "html-end" }}

</body>
</html>
{{ end }}
`))

var tmplFuncs = template.FuncMap{
	"isList": func(val interface{}) bool {
		switch val.(type) {
		case []interface{}:
			return true
		default:
			return false
		}
	},
}
