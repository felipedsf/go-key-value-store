package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestJSON(t *testing.T) {
	t.Parallel()

	header := http.Header{}
	headerKey := "Content-Type"
	headerValue := "application/json; charset=utf-8"
	header.Add(headerKey, headerValue)

	testCases := []struct {
		in     interface{}
		header http.Header
		out    string
	}{
		{map[string]string{"hello": "world"}, header, `{"hello":"world"}`},
		{map[string]string{"hello": "tables"}, header, `{"hello":"tables"}`},
		{make(chan bool), header, `{"error":"json: unsupported type: chan bool"}`},
	}
	for _, testCase := range testCases {
		recorder := httptest.NewRecorder()
		JSON(recorder, testCase.in)

		response := recorder.Result()
		defer response.Body.Close()

		got, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatalf("Error reading response body: %s", err)
		}
		if string(got) != testCase.out {
			t.Errorf("Got %s, expected %s", string(got), testCase.out)
		}
		if contentType := response.Header.Get(headerKey); contentType != headerValue {
			t.Errorf("Got %s, expected %s", contentType, headerValue)
		}
	}
}

func BenchmarkGet(b *testing.B) {

	makeStorage(b)
	defer cleanupStorage(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Get(context.Background(), "key1")
	}
}

func makeStorage(tb testing.TB) {
	err := os.Mkdir("testdata", 0755)
	if err != nil && !os.IsExist(err) {
		tb.Fatalf("Couldn't create directory testdata: %s", err)
	}
	StoragePath = "testdata"
}
func cleanupStorage(tb testing.TB) {
	if err := os.RemoveAll(StoragePath); err != nil {
		tb.Errorf("Failed to delete storage path: %s", StoragePath)
	}
	StoragePath = "/tmp/kv"
}
