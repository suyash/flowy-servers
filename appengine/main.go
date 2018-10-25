package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

var apiKey = os.Getenv("API_KEY")

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/set", func(res http.ResponseWriter, req *http.Request) {
		ctx := appengine.NewContext(req)
		set(ctx, res, req)
	})

	r.HandleFunc("/{id}", func(res http.ResponseWriter, req *http.Request) {
		ctx := appengine.NewContext(req)

		vars := mux.Vars(req)
		id := vars["id"]

		getOrDelete(ctx, id, res, req)
	})

	r.HandleFunc("/", index)

	http.Handle("/", r)
	appengine.Main()
}

func index(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(res, "ok")
}

// Task is a single item.
type Task struct {
	ID       string   `json:"id"`
	Text     string   `json:"text"`
	Checked  bool     `json:"checked"`
	Children []string `json:"children"`
}

func set(ctx context.Context, res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodOptions {
		addCORSHeaders(res)
		res.WriteHeader(http.StatusOK)
		return
	}

	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	if req.Header.Get("X-API-Key") != apiKey {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	data, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()

	if err != nil {
		log.Errorf(ctx, "Could not read body")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		log.Errorf(ctx, "Could not unmarshal, error: %v", err)
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	key := datastore.NewKey(ctx, "Task", task.ID, 0, nil)
	if _, err := datastore.Put(ctx, key, &task); err != nil {
		log.Errorf(ctx, "Could not add to datastore, error: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	addCORSHeaders(res)
	res.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(res, "{ \"ok\": true }")
}

func getOrDelete(ctx context.Context, id string, res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodOptions {
		addCORSHeaders(res)
		res.WriteHeader(http.StatusOK)
		return
	}

	if req.Method != http.MethodGet && req.Method != http.MethodDelete {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	if req.Header.Get("X-API-Key") != apiKey {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	addCORSHeaders(res)

	key := datastore.NewKey(ctx, "Task", id, 0, nil)

	res.Header().Set("Content-Type", "application/json")

	if req.Method == http.MethodDelete {
		if err := datastore.Delete(ctx, key); err != nil {
			log.Errorf(ctx, "Could not remove from datastore, error: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(res, "\"ok\": true")
	} else {
		var task Task
		if err := datastore.Get(ctx, key, &task); err != nil {
			log.Errorf(ctx, "Could not get from datastore, error: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		data, err := json.Marshal(task)
		if err != nil {
			log.Errorf(ctx, "Could not encode to json, error: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Write(data)
	}
}

func addCORSHeaders(res http.ResponseWriter) {
	res.Header().Set("Access-Control-Allow-Origin", "https://suy.io")
	res.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE")
	res.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, X-API-Key")
	res.Header().Set("Access-Control-Allow-Credentials", "true")
}
