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

	"github.com/gorilla/mux"
)

type Entry struct {
	ID    string                 `json:"id"`
	Date  time.Time              `json:"date"`
	Type  string                 `json:"type"`
	Note  string                 `json:"note,omitempty"`
	Value float64                `json:"value"`
	Data  map[string]interface{} `json:"data,omitempty"`
}

type Entries []Entry

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

	router := mux.NewRouter()

	router.Methods("GET").Path("/").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		renderEntries(repo, w, req)
	})

	router.Methods("GET").Path("/new").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		RenderInput(w, req, "")
	})

	router.Methods("GET").Path("/new/{type}").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		RenderInput(w, req, mux.Vars(req)["type"])
	})

	router.Methods("GET").Path("/query").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		renderQuery(repo, w, req)
	})

	router.Methods("GET").Path("/{id}").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		renderEntry(repo, mux.Vars(req)["id"], w, req)
	})

	router.Methods("POST").Path("/new").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		createEntry(repo, w, req)
	})

	router.Methods("POST").Path("/{id}").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		editEntry(repo, w, req, mux.Vars(req)["id"])
	})

	router.Methods("GET").Path("/{id}/edit").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		renderEditEntry(repo, w, req, mux.Vars(req)["id"])
	})

	http.Handle("/", router)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	log.Printf("Listening on http://%s", config.addr)
	log.Fatal(http.ListenAndServe(config.addr, nil))
}

func renderEntries(repo Repository, w http.ResponseWriter, req *http.Request) {
	now := time.Now().UTC()
	entries, err := repo.FindBetween(req.Context(), now.AddDate(0, 0, -30), now, Descending)
	if err != nil {
		log.Printf("Could not list entries: %s", err)
		http.Error(w, fmt.Sprintf("could not list entries: %s", err), http.StatusInternalServerError)
		return
	}

	err = entries.Render(w, req.Header.Get("Content-Type"))
	if err != nil {
		log.Printf("Could not render entries: %s", err)
		fmt.Fprintf(w, "\nCould not render entries: %s\n", err)
	}
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

func renderQuery(repo Repository, w http.ResponseWriter, req *http.Request) {
	var entries Entries = nil
	var err error

	query := req.URL.Query().Get("query")
	if query != "" {
		entries, err = repo.Query(req.Context(), query)
		if err != nil {
			log.Printf("Could not execute query: %s", err)
		}
	}

	err = tmplQuery.Execute(w, map[string]interface{}{
		"Query": query,
		"Entries": entries,
		"Error": err,
	})
	if err != nil {
		log.Printf("Could not execute template: %s", err)
		fmt.Fprintf(w, "\n%s\n", err)
	}
}

var tmplQuery = template.Must(template.New("query").Parse(`<!doctype html>
<html>
<head>
	<meta charset="utf-8" />
	<title>Query!</title>
</head>

<body>
	<form method="GET" action="/query">
		<div>
			<textarea name="query" cols="80" rows="10">{{ .Query }}</textarea>
		</div>

		<input type="submit" value="Search!" />
	</form>

	{{ if .Error }}
	<div class="error">
		<pre>{{ .Error }}</pre>
	</div>
	{{ end }}

	<div class="result">
		<pre>{{ .Entries.RenderJSONString }}</pre>
	</div>
</body>
</html>
`))

func createEntry(repo Repository, w http.ResponseWriter, req *http.Request) {
	entry, err := FromPostForm(req)
	if err != nil {
		log.Printf("Could not parse entry: %s", err)
		http.Error(w, fmt.Sprintf("Could not parse entry: %s", err), http.StatusBadRequest)
		return
	}

	id, err := repo.Create(req.Context(), entry)
	if err != nil {
		log.Printf("Could not create entry: %s", err)
		http.Error(w, fmt.Sprintf("Could not create new entry: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", "/" + id)
	w.WriteHeader(http.StatusFound)
}

func renderEditEntry(repo Repository, w http.ResponseWriter, req *http.Request, id string) {
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

	RenderEdit(w, req, entry)
}

func editEntry(repo Repository, w http.ResponseWriter, req *http.Request, id string) {
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

	editedEntry, err := FromPostForm(req)
	if err != nil {
		log.Printf("Could not parse entry: %s", err)
		http.Error(w, fmt.Sprintf("Could not parse entry: %s", err), http.StatusBadRequest)
		return
	}

	editedEntry.ID = entry.ID
	editedEntry.Date = entry.Date
	editedEntry.Type = entry.Type

	err = repo.Update(req.Context(), editedEntry)
	if err != nil {
		log.Printf("Could not update entry: %s", err)
		http.Error(w, fmt.Sprintf("Could not update entry: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", "/" + entry.ID)
	w.WriteHeader(http.StatusFound)
}

func FromPostForm(req *http.Request) (*Entry, error) {
	err := req.ParseForm()
	if err != nil {
		return nil, fmt.Errorf("invalid form: %s", err)
	}

	date := time.Now().UTC().Round(time.Millisecond)
	if len(req.PostForm.Get("date")) > 0 {
		date, err = time.Parse(time.RFC3339, req.PostForm.Get("date"))
		if err != nil {
			return nil, fmt.Errorf("value of 'date' (%q) is not a valid date: %s",
				req.PostForm.Get("date"), err)
		}
	}

	entry := &Entry{
		Date: date,
		Type: req.PostForm.Get("type"),
		Note: req.PostForm.Get("note"),
	}

	if val, ok := req.PostForm["value"]; ok && len(val) >= 1 {
		v, err := strconv.ParseFloat(val[0], 64)
		if err != nil {
			return nil, fmt.Errorf("value %q of 'value' is not a number: %s", val[0], err)
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

	return entry, nil
}

