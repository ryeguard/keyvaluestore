package main

import (
	"fmt"
	"sync"
	"time"
)

type Store struct {
	data map[string][]*entry
	mu   sync.Mutex
}

type entry struct {
	Value     string    `json:"value"`
	Ts        time.Time `json:"enteredAt"`
	deletedAt *time.Time
}

func (kv *Store) Put(key, value string) {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	_, ok := kv.data[key]
	if !ok {
		kv.data[key] = []*entry{{Value: value, Ts: time.Now()}}
		return
	}

	// "delete" latest item
	now := time.Now()
	kv.data[key][len(kv.data[key])-1].deletedAt = &now

	kv.data[key] = append(kv.data[key], &entry{Value: value, Ts: time.Now()})
}

func (kv *Store) Get(key string) (string, error) {
	entries, err := kv.GetAll(key)
	if err != nil {
		return "", err
	}

	latest := entries[len(entries)-1]
	if latest.deletedAt != nil {
		return "", fmt.Errorf("no value")
	}

	return latest.Value, nil
}

func (kv *Store) Delete(key string) {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	entries, ok := kv.data[key]
	if !ok {
		return
	}

	latest := entries[len(entries)-1]
	now := time.Now()
	latest.deletedAt = &now

	kv.data[key][len(entries)-1] = latest
}

func (kv *Store) GetAll(key string) ([]*entry, error) {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	v, ok := kv.data[key]
	if !ok {
		return nil, fmt.Errorf("no value")
	}

	if len(v) == 0 {
		return nil, fmt.Errorf("no value")
	}

	return v, nil
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
