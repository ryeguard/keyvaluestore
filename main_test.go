package main

import (
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
			data: map[string][]*entry{},
			mu:   sync.Mutex{},
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPut, "/entries/", nil)
		putEntryFunc(store)(w, r)
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
			data: map[string][]*entry{},
			mu:   sync.Mutex{},
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPut, "/entries/testKey", nil)
		r.SetPathValue("key", "testKey")
		putEntryFunc(store)(w, r)
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
			data: map[string][]*entry{},
			mu:   sync.Mutex{},
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPut, "/entries/testKey", strings.NewReader("bad input"))
		r.SetPathValue("key", "testKey")
		putEntryFunc(store)(w, r)
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
			data: map[string][]*entry{},
			mu:   sync.Mutex{},
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPut, "/entries/testKey", strings.NewReader(`{"value":"testValue"}`))
		r.SetPathValue("key", "testKey")
		putEntryFunc(store)(w, r)
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
	testTime, err := time.Parse(time.RFC3339, "2024-05-07T21:08:00.0Z")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("OK", func(t *testing.T) {
		store := &Store{
			data: map[string][]*entry{
				"testKey": {{Value: "testValue"}},
			},
			mu: sync.Mutex{},
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPut, "/entries/testKey", nil)
		r.SetPathValue("key", "testKey")
		getEntryFunc(store)(w, r)
		res := w.Result()

		require.Equal(t, http.StatusOK, res.StatusCode)

		b, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("failed reading return body")
		}

		require.Equal(t, "GET testKey=testValue", string(b))
	})

	t.Run("OK - with history", func(t *testing.T) {
		store := &Store{
			data: map[string][]*entry{
				"testKey": {
					{Value: "firstValue", Ts: testTime},
					{Value: "secondValue", Ts: testTime.Add(time.Second)},
				},
			},
			mu: sync.Mutex{},
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPut, "/entries/testKey", nil)
		r.SetPathValue("key", "testKey")
		getEntryFunc(store)(w, r)
		res := w.Result()

		require.Equal(t, http.StatusOK, res.StatusCode)

		b, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("failed reading return body")
		}

		require.Equal(t, "GET testKey=secondValue", string(b))
	})

	t.Run("OK - with deleted entry", func(t *testing.T) {
		store := &Store{
			data: map[string][]*entry{
				"testKey": {
					{Value: "testValue", Ts: testTime.Add(time.Second), deletedAt: &testTime},
				},
			},
			mu: sync.Mutex{},
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPut, "/entries/testKey", nil)
		r.SetPathValue("key", "testKey")
		getEntryFunc(store)(w, r)
		res := w.Result()

		require.Equal(t, http.StatusNotFound, res.StatusCode)

		b, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("failed reading return body")
		}

		require.Equal(t, "testKey does not exist\n", string(b))
	})
}

func TestGetHistory(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2024-05-07T21:08:00.0Z")
	if err != nil {
		t.Fatal(err)
	}

	store := &Store{
		data: map[string][]*entry{
			"existentKey": {{Value: "testValue1", Ts: testTime}, {Value: "testValue2", Ts: testTime.Add(time.Second)}},
		},
		mu: sync.Mutex{},
	}

	t.Run("non-existent key", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/entries/nonExistentKey/history", nil)
		r.SetPathValue("key", "nonExistentKey")
		getHistoryFunc(store)(w, r)
		res := w.Result()

		require.Equal(t, http.StatusNotFound, res.StatusCode)

		b, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("failed reading return body")
		}

		require.Equal(t, "nonExistentKey does not exist\n", string(b))
	})

	t.Run("existent key", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/entries/existentKey/history", nil)
		r.SetPathValue("key", "existentKey")
		getHistoryFunc(store)(w, r)
		res := w.Result()

		require.Equal(t, http.StatusOK, res.StatusCode)

		b, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("failed reading return body")
		}

		require.Equal(t, "[{\"value\":\"testValue1\",\"enteredAt\":\"2024-05-07T21:08:00Z\"},{\"value\":\"testValue2\",\"enteredAt\":\"2024-05-07T21:08:01Z\"}]\n", string(b))
	})
}
