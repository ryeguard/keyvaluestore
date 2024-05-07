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

	http.HandleFunc("/kv", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PUT":
			var p struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			}

			err := json.NewDecoder(r.Body).Decode(&p)
			if err != nil {
				http.Error(w, fmt.Sprintf("decode: %v", err), http.StatusBadRequest)
				return
			}

			if p.Key == "" || p.Value == "" {
				http.Error(w, "both key and value must be set", http.StatusBadRequest)
				return
			}
			store.Put(p.Key, p.Value)

			fmt.Fprintf(w, "PUT %v=%v", p.Key, p.Value)
		case "GET":
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
					http.Error(w, fmt.Sprintf("%v does not exist", k), http.StatusBadRequest)
					return
				}

				fmt.Fprintf(w, "GET %v=%v", k, entry[len(entry)-1].Value)
			}
		default:
			fmt.Println(r.Method)
		}
	})

	http.HandleFunc("GET /kv/hist/", historyHandler(store))

	http.HandleFunc("PUT /kv/del_hist", func(w http.ResponseWriter, r *http.Request) {
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
	})

	fmt.Println("Serving...")
	mux := http.NewServeMux()
	http.ListenAndServe(":8080", mux)
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
