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
	testTime, err := time.Parse(time.RFC3339, "2024-05-07T21:08:00.0Z")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		request        *http.Request
		wantStatusCode int
	}{
		{
			request:        httptest.NewRequest(http.MethodGet, "/kv/hist?nonExistingValue", nil),
			wantStatusCode: http.StatusNotFound,
		},
		{
			request:        httptest.NewRequest(http.MethodGet, "/kv/hist?existingKey", nil),
			wantStatusCode: http.StatusOK,
		},
	}

	store := &Store{
		data: map[string][]entry{
			"existingKey": {{"testValue", testTime}},
		},
		mu: sync.Mutex{},
	}

	for _, tc := range tests {
		w := httptest.NewRecorder()
		historyHandler(store)(w, tc.request)
		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != tc.wantStatusCode {
			t.Fatalf("status got != want: %v != %v", res.StatusCode, tc.wantStatusCode)
		}

		if !is2XX(tc.wantStatusCode) {
			continue
		}

		var got []entry
		err := json.NewDecoder(res.Body).Decode(&got)
		if err != nil {
			t.Fatal(err)
		}

		if len(got) != 1 {
			t.Fatal("want len 1")
		}

		want := entry{Value: "testValue", Ts: testTime}
		if want != got[0] {
			t.Fatal("want and got not equal")
		}
	}
}

func is2XX(code int) bool {
	return code >= 200 && code < 300
}
