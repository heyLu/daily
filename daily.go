package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Entry struct {
	ID    string                 `json:"id"`
	Date  time.Time              `json:"date"`
	Type  string                 `json:"type"`
	Note  string                 `json:"note,omitempty"`
	Value float64                `json:"value"`
	Data  map[string]interface{} `json:"data,omitempty"`
}

func main() {
	dbName := "./test.db"
	repo, err := NewRepository(dbName)
	if err != nil {
		log.Fatalf("Failed to open database %q: %s", dbName, err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			renderMoodInput(w, req)
		case http.MethodPost:
			saveEntry(repo, w, req)
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	})

	addr := "localhost:11111"
	log.Printf("Listening on http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func saveEntry(repo Repository, w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Println(err)
		fmt.Fprintln(w, err)
		return
	}

	entry := Entry{
		Date: time.Now().Round(time.Millisecond),
		Type: req.PostForm.Get("type"),
		Note: req.PostForm.Get("note"),
	}

	if val, ok := req.PostForm["value"]; ok && len(val) >= 1 {
		v, err := strconv.ParseFloat(val[0], 64)
		if err != nil {
			fmt.Fprintf(w, "value %q of 'value' is not a number: %s\n", val[0], err)
			return
		}
		entry.Value = v
	}

	additionalData := map[string]interface{}{}
	for key, vals := range req.PostForm {
		// ignore "standard" fields
		switch key {
		case "type", "note", "value":
			continue
		}

		parsedVals := []interface{}{}
		for _, val := range vals {
			var parsedVal interface{}
			err := json.Unmarshal([]byte(val), &parsedVal)
			if err != nil {
				parsedVals = append(parsedVals, val)
			} else {
				parsedVals = append(parsedVals, parsedVal)
			}
		}
		if len(parsedVals) == 1 {
			additionalData[key] = parsedVals[0]
		} else {
			additionalData[key] = parsedVals
		}
	}
	if len(additionalData) >= 1 {
		entry.Data = additionalData
	}

	id, err := repo.Create(req.Context(), entry)
	if err != nil {
		log.Printf("Could not create entry: %s", err)
		fmt.Fprintf(w, "Could not create new entry: %s", err)
		return
	}

	entry.ID = id

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	w.Write(data)
}

func renderMoodInput(w http.ResponseWriter, req *http.Request) {
	tmpl := template.Must(template.New("").Parse(`<!doctype html>
<html>
<head>
	<title>mood (daily)</title>
	<meta name="viewport" content="width=device-width, initial-scale=1" />

<style>
body {
	margin: 0;

	font-family: monospace;
}

.hidden {
	visibility: hidden;
}

#mood-gradient {
	width: 100%;
	height: 100vh;
	background-image: linear-gradient(to right, red, orange, yellow, rgb(200, 255, 0) 40%, green, blue);
	opacity: 0.8;
}

#content {
	margin: 1em;
}

input[type=submit] {
	margin-top: 1em;
}
</style>

</head>

<body>
	<div id="mood-gradient" class="hidden"></div>

	<section id="content">
		<h1>More stuff</h1>
		<p>Let's try things... ^^</p>

		<form method="POST" action="/">
			<input name="type" value="mood" hidden />

			<div class="field">
				<label for="value">Mood</label>
				<input id="mood" name="value" type="range" min="0" max="1" step="0.01" />
			</div>

			<div class="field">
				<label for="Note">Note</label>
				<input name="note" type="text" />
			</div>

			<input type="submit" value="Save" />
		</form>
	</section>

	<script>
		let moodInput = document.querySelector("#mood");
		let moodGradient = document.querySelector("#mood-gradient");

		moodGradient.classList.remove("hidden");

		moodGradient.addEventListener("click", function(ev) {
			moodInput.value = ev.clientX / document.body.clientWidth;
		});
	</script>
</body>
</html>
`))
	err := tmpl.Execute(w, map[string]interface{}{})
	if err != nil {
		log.Println(err)
		fmt.Fprint(w, err)
	}
}
