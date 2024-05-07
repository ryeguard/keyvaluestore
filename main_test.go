package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestGetHistory(t *testing.T) {
	ts, err := time.Parse(time.RFC3339, "2024-05-07T21:08:00.0Z")
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/kv/hist?testKey", nil)
	w := httptest.NewRecorder()

	store := &Store{
		data: map[string][]entry{
			"testKey": {{"testValue", ts}},
		},
		mu: sync.Mutex{},
	}

	historyHandler(store)(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status code not %v: %v", http.StatusOK, res.StatusCode)
	}

	var got []entry

	err = json.NewDecoder(res.Body).Decode(&got)
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 1 {
		t.Fatal("want len 1")
	}

	want := entry{Value: "testValue", Ts: ts}
	if want != got[0] {
		t.Fatal("want and got not equal")
	}
}
