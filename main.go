package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

func main() {

	store := &Store{
		data: map[string][]*entry{},
		mu:   sync.Mutex{},
	}

	mux := http.NewServeMux()

	mux.HandleFunc("PUT /entries/{key}", putEntryFunc(store))
	mux.HandleFunc("GET /entries/{key}", getEntryFunc(store))
	mux.HandleFunc("DELETE /entries/{key}", deleteEntryFunc(store))
	mux.HandleFunc("GET /entries/{key}/history", getHistoryFunc(store))
	mux.HandleFunc("DELETE /entries/{key}/history", deleteHistoryFunc(store))

	http.ListenAndServe(":8080", mux)
}

func putEntryFunc(store *Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		key := r.PathValue("key")
		if key == "" {
			http.Error(w, "key not provided", http.StatusBadRequest)
			return
		}

		var p struct {
			Value string `json:"value"`
		}

		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			if err.Error() == "EOF" {
				http.Error(w, "request body is empty", http.StatusBadRequest)
				return
			}
			http.Error(w, "request body decode error", http.StatusBadRequest)
			return
		}

		if p.Value == "" {
			http.Error(w, "value must be set", http.StatusBadRequest)
			return
		}
		store.Put(key, p.Value)

		fmt.Fprintf(w, "PUT %v=%v", key, p.Value)
	}
}

func getEntryFunc(store *Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		key := r.PathValue("key")
		if key == "" {
			http.Error(w, "key not provided", http.StatusBadRequest)
			return
		}

		value, err := store.Get(key)
		if err != nil {
			http.Error(w, fmt.Sprintf("%v does not exist", key), http.StatusNotFound)
			return
		}

		fmt.Fprintf(w, "GET %v=%v", key, value)
	}
}

func deleteEntryFunc(store *Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")
		if key == "" {
			http.Error(w, "key not provided", http.StatusBadRequest)
			return
		}

		store.Delete(key)
	}
}

func getHistoryFunc(store *Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		key := r.PathValue("key")
		if key == "" {
			http.Error(w, "key not provided", http.StatusBadRequest)
			return
		}

		entries, err := store.GetAll(key)
		if err != nil {
			http.Error(w, fmt.Sprintf("%v does not exist", key), http.StatusNotFound)
			return
		}

		err = json.NewEncoder(w).Encode(entries)
		if err != nil {
			http.Error(w, fmt.Sprintf("encode: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func deleteHistoryFunc(store *Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		key := r.PathValue("key")
		if key == "" {
			http.Error(w, "key not provided", http.StatusBadRequest)
			return
		}

		store.DeleteAll(key)
	}
}
