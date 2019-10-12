package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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

type VisualizeInfo struct {
	Color  string
	Amount float64
}

func (vi VisualizeInfo) ToHTML(width, height int) template.HTML {
	buf := new(bytes.Buffer)
	numFullSpans := int(vi.Amount)
	for i := 0; i < numFullSpans; i++ {
		fmt.Fprintf(buf, `<span style="display: inline-block; width: %dpx; height: %dpx; background-color: %s"></span>`,
			width, height, vi.Color)
	}
	partialSpanWidth := (vi.Amount - float64(numFullSpans)) * float64(width)
	fmt.Fprintf(buf, `<span style="display: inline-block; width: %dpx; height: %dpx; background-color: %s"></span>`,
		int(partialSpanWidth), height, vi.Color)
	return template.HTML(buf.String())
}

func Visualize(typ string, numEntries int, sumValue float64) VisualizeInfo {
	avgValue := sumValue / float64(numEntries)
	switch typ {
	case "coffee":
		return VisualizeInfo{Color: "brown", Amount: sumValue}
	case "water":
		return VisualizeInfo{Color: "lightblue", Amount: sumValue}
	case "throat":
		// shade of green
		//   0.0 == fine == rgb(255, 255, 255) == white
		//   1.0 == bad  == rgb( 40, 203,   0) == awfulgreen
		red := 255 - ((255 - 40) * avgValue)
		green := 255 - ((255 - 203) * avgValue)
		blue := 255 - ((255 - 0) * avgValue)
		return VisualizeInfo{Color: fmt.Sprintf("rgb(%.1f, %.1f, %.1f)", red, green, blue), Amount: 1}
	case "mood":
		return VisualizeInfo{Color: "yellow", Amount: avgValue}
	case "shower":
		return VisualizeInfo{Color: "blue", Amount: sumValue / 10}
	case "expense":
		return VisualizeInfo{Color: "red", Amount: sumValue / 10}
	default:
		return VisualizeInfo{Color: "grey", Amount: float64(numEntries)}
	}
}

var entryFuncs = template.FuncMap{
	"visualize": Visualize,
}

var tmplEntry = template.Must(tmplBase.Funcs(entryFuncs).New("entry").Parse(`{{ template "html-start" . }}
<a href="/new">/new</a>

<article class="entry">
	<header>
		<h1>{{ .Entry.Date }}</h1>
		<span class="type">{{ .Entry.Type }}</span>
		{{ (visualize .Entry.Type 1 .Entry.Value).ToHTML 16 16 }}
		<a href="/{{ .Entry.ID }}/edit">/edit</a>
	</header>

	<p>{{ .Entry.Note }}</p>

	<div class="data">
		<pre>{{ .Entry.RenderJSONString }}</pre>
	</div>
</article>
{{ template "html-end" }}
`))
