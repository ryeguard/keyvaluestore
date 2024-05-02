package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type PutRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type KeyValueStore struct {
	data map[string]string
	mu   sync.Mutex
}

func main() {

	store := KeyValueStore{
		data: map[string]string{},
		mu:   sync.Mutex{},
	}

	http.HandleFunc("/kv", func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case "PUT":
			var p PutRequest
			err := json.NewDecoder(r.Body).Decode(&p)
			if err != nil {
				http.Error(w, fmt.Sprintf("decode: %v", err), http.StatusBadRequest)
				return
			}

			if p.Key == "" || p.Value == "" {
				http.Error(w, "both key and value must be set", http.StatusBadRequest)
				return
			}
			store.mu.Lock()
			store.data[p.Key] = p.Value
			store.mu.Unlock()

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

				fmt.Fprintf(w, "GET %v=%v", k, entry)
			}
		default:
			fmt.Println(r.Method)
		}
	})

	fmt.Println("Serving...")
	http.ListenAndServe(":8080", nil)
}
