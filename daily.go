package main

import (
	"encoding/json"
	"flag"
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

var config struct {
	addr   string
	dbName string
}

func main() {
	flag.StringVar(&config.addr, "addr", "localhost:11111", "Address to listen on")
	flag.StringVar(&config.dbName, "db", "./test.db", "Path to the database to use")
	flag.Parse()

	log.Printf("Opening database %q", config.dbName)
	repo, err := NewRepository(config.dbName, "./schema-init.sql")
	if err != nil {
		log.Fatalf("Failed to open database %q: %s", config.dbName, err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/":
			http.Error(w, "not implemented", http.StatusNotImplemented)
		default:
			id := req.URL.Path[1:]
			renderEntry(repo, id, w, req)
		}
	})

	http.HandleFunc("/new", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			renderMoodInput(w, req)
		case http.MethodPost:
			saveEntry(repo, w, req)
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	})

	log.Printf("Listening on http://%s", config.addr)
	log.Fatal(http.ListenAndServe(config.addr, nil))
}

func renderEntry(repo Repository, id string, w http.ResponseWriter, req *http.Request) {
	entry, err := repo.Get(req.Context(), id)
	if err != nil {
		log.Printf("Could not get entry: %s", err)
		http.Error(w, fmt.Sprintf("Could not get entry: %s", err), http.StatusInternalServerError)
		return
	}

	if entry == nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	entryJSON, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		log.Printf("Could not serialize entry: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Write(entryJSON)
}

func saveEntry(repo Repository, w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Println(err)
		fmt.Fprintln(w, err)
		return
	}

	date := time.Now().UTC().Round(time.Millisecond)
	if len(req.PostForm.Get("date")) > 0 {
		date, err = time.Parse(time.RFC3339, req.PostForm.Get("date"))
		if err != nil {
			http.Error(w, fmt.Sprintf("value of 'date' (%q) is not a valid date: %s",
				req.PostForm.Get("date"), err),
				http.StatusBadRequest)
			return
		}
	}

	entry := Entry{
		Date: date,
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
		case "date", "type", "note", "value":
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

		<form method="POST" action="/new">
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
