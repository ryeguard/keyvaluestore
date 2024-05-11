package main

import (
	"fmt"
	"sync"
	"time"
)

type Store struct {
	data map[string][]entry
	mu   sync.Mutex
}

type entry struct {
	Value string    `json:"value"`
	Ts    time.Time `json:"enteredAt"`
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

func (kv *Store) Get(key string) (string, error) {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	v, ok := kv.data[key]
	if !ok {
		return "", fmt.Errorf("no value")
	}

	if len(v) == 0 {
		return "", fmt.Errorf("no value")
	}

	return v[len(v)-1].Value, nil
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
