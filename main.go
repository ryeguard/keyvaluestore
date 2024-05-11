package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type entry struct {
	Value string    `json:"value"`
	Ts    time.Time `json:"enteredAt"`
}

type Store struct {
	data map[string][]entry
	mu   sync.Mutex
}

func (kv *Store) Put(key, value string) {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	_, ok := kv.data[key]
	if !ok {
		kv.data[key] = []entry{{Value: value, Ts: time.Now()}}
		return
	}

	kv.data[key] = append(kv.data[key], entry{Value: value, Ts: time.Now()})
}

func (kv *Store) DeleteAll(key string) {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	_, ok := kv.data[key]
	if !ok {
		return
	}

	kv.data[key] = nil
}

func main() {

	store := &Store{
		data: map[string][]entry{},
		mu:   sync.Mutex{},
	}

	http.HandleFunc("PUT /entries/{key}", putEntry(store))
	http.HandleFunc("GET /entries/{key}", getEntry(store))
	http.HandleFunc("GET /kv/hist/", historyHandler(store))
	http.HandleFunc("PUT /kv/del_hist", deleteHistoryHandler(store))

	mux := http.NewServeMux()
	http.ListenAndServe(":8080", mux)
}

func putEntry(store *Store) func(w http.ResponseWriter, r *http.Request) {
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

func getEntry(store *Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		key := r.PathValue("key")
		if key == "" {
			http.Error(w, "key not provided", http.StatusBadRequest)
			return
		}

		entry, ok := store.data[key]
		if !ok {
			http.Error(w, fmt.Sprintf("%v does not exist", key), http.StatusNotFound)
			return
		}

		fmt.Fprintf(w, "GET %v=%v", key, entry[len(entry)-1].Value)
	}
}

func historyHandler(store *Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		err := r.ParseForm()
		if err != nil {
			http.Error(w, fmt.Sprintf("parse form: %v", err), http.StatusBadRequest)
			return
		}

		if len(r.Form) != 1 {
			http.Error(w, "Only a single query parameter is supported", http.StatusBadRequest)
			return
		}

		for k, v := range r.Form {
			if len(v) != 1 {
				http.Error(w, "Only a single value per param is supported", http.StatusBadRequest)
				return
			}

			entry, ok := store.data[k]
			if !ok {
				http.Error(w, fmt.Sprintf("%v does not exist", k), http.StatusNotFound)
				return
			}

			err := json.NewEncoder(w).Encode(entry)
			if err != nil {
				http.Error(w, fmt.Sprintf("encode: %v", err), http.StatusInternalServerError)
				return
			}
		}
	}
}

func deleteHistoryHandler(store *Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			http.Error(w, "must be PUT method", http.StatusBadRequest)
		}

		var d struct {
			Key string `json:"key"`
		}

		err := json.NewDecoder(r.Body).Decode(&d)
		if err != nil {
			http.Error(w, fmt.Sprintf("decode: %v", err), http.StatusBadRequest)
			return
		}

		if d.Key == "" {
			http.Error(w, "key must be set", http.StatusBadRequest)
			return
		}

		store.DeleteAll(d.Key)

		fmt.Fprintf(w, "PUT %v=''", d.Key)
	}
}
