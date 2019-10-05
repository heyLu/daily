package main

import (
	"encoding/json"
	"flag"
	"fmt"
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

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/":
			renderEntries(repo, w, req)
		default:
			id := req.URL.Path[1:]
			renderEntry(repo, id, w, req)
		}
	})

	http.HandleFunc("/new", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			RenderInput(w, req, "")
		case http.MethodPost:
			createEntry(repo, w, req)
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/new/", func(w http.ResponseWriter, req *http.Request) {
		typeName := req.URL.Path[5:]
		RenderInput(w, req, typeName)
	})

	http.HandleFunc("/edit/", func(w http.ResponseWriter, req *http.Request) {
		id := req.URL.Path[6:]
		switch req.Method {
		case http.MethodGet:
			renderEditEntry(repo, w, req, id)
		case http.MethodPost:
			editEntry(repo, w, req, id)
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	})

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

	entry.ID = id

	err = entry.Render(w, req.Header.Get("Content-Type"))
	if err != nil {
		log.Printf("Could not render entry: %s", err)
		fmt.Fprintf(w, "\nCould not render entry: %s\n", err)
	}
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

