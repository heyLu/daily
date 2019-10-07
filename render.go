package main

import (
	"bytes"
	"encoding/json"
	"io"
)

func (e Entry) Render(w io.Writer, contentType string) error {
	return e.RenderJSON(w)
}

func (e Entry) RenderJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(e)
}

func (es Entries) Render(w io.Writer, contentType string) error {
	return es.RenderJSON(w)
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
