package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPutEntry(t *testing.T) {
	t.Run("missing key", func(t *testing.T) {
		store := &Store{
			data: map[string][]entry{},
			mu:   sync.Mutex{},
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPut, "/entries/", nil)
		putEntry(store)(w, r)
		res := w.Result()

		require.Equal(t, http.StatusBadRequest, res.StatusCode)

		b, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("failed reading return body")
		}

		require.Equal(t, "key not provided\n", string(b))
	})

	t.Run("missing body", func(t *testing.T) {
		store := &Store{
			data: map[string][]entry{},
			mu:   sync.Mutex{},
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPut, "/entries/testKey", nil)
		r.SetPathValue("key", "testKey")
		putEntry(store)(w, r)
		res := w.Result()

		require.Equal(t, http.StatusBadRequest, res.StatusCode)

		b, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("failed reading return body")
		}

		require.Equal(t, "request body is empty\n", string(b))
	})

	t.Run("incorrect body", func(t *testing.T) {
		store := &Store{
			data: map[string][]entry{},
			mu:   sync.Mutex{},
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPut, "/entries/testKey", strings.NewReader("bad input"))
		r.SetPathValue("key", "testKey")
		putEntry(store)(w, r)
		res := w.Result()

		require.Equal(t, http.StatusBadRequest, res.StatusCode)

		b, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("failed reading return body")
		}

		require.Equal(t, "request body decode error\n", string(b))
	})

	t.Run("OK", func(t *testing.T) {
		store := &Store{
			data: map[string][]entry{},
			mu:   sync.Mutex{},
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPut, "/entries/testKey", strings.NewReader(`{"value":"testValue"}`))
		r.SetPathValue("key", "testKey")
		putEntry(store)(w, r)
		res := w.Result()

		require.Equal(t, http.StatusOK, res.StatusCode)

		b, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("failed reading return body")
		}

		require.Equal(t, "PUT testKey=testValue", string(b))
	})
}

func TestGetEntry(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		store := &Store{
			data: map[string][]entry{
				"testKey": {{Value: "testValue"}},
			},
			mu: sync.Mutex{},
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPut, "/entries/testKey", nil)
		r.SetPathValue("key", "testKey")
		getEntry(store)(w, r)
		res := w.Result()

		require.Equal(t, http.StatusOK, res.StatusCode)

		b, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("failed reading return body")
		}

		require.Equal(t, "GET testKey=testValue", string(b))
	})
}

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
