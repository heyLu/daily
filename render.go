package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"strings"
	"time"
)

func (e Entry) Render(w io.Writer, contentType string) error {
	if strings.Contains(contentType, "html") {
		return e.RenderHTML(w)
	}
	return e.RenderJSON(w)
}

func (es Entries) Render(w io.Writer, contentType string) error {
	return es.RenderJSON(w)
}

func (e Entry) RenderJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(e)
}

func (e Entry) RenderJSONString() (string, error) {
	buf := new(bytes.Buffer)
	err := e.RenderJSON(buf)
	return buf.String(), err
}

func (es Entries) RenderJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(es)
}

func (es Entries) RenderJSONString() (string, error) {
	buf := new(bytes.Buffer)
	err := es.RenderJSON(buf)
	return buf.String(), err
}

func (e Entry) RenderHTML(w io.Writer) error {
	e.Date = e.Date.Round(time.Second)
	return tmplEntry.Execute(w, map[string]interface{}{
		"Entry": e,
		"Stylesheet": "entry.css",
	})
}

var tmplEntry = template.Must(tmplBase.New("entry").Parse(`{{ template "html-start" . }}
<article class="entry">
	<header>
		<h1>{{ .Entry.Date }}</h1>
		<span class="type">{{ .Entry.Type }}</span>
		<a href="/{{ .Entry.ID }}/edit">edit</a>
	</header>

	<p>{{ .Entry.Note }}</p>

	<div class="data">
		<pre>{{ .Entry.RenderJSONString }}</pre>
	</div>
</article>
{{ template "html-end" }}
`))
