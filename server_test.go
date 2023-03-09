package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
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

func TestGet(t *testing.T) {
	t.Parallel()
	
	makeStorage(t)
	defer cleanupStorage(t)

	kvStore := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key4": "value4",
	}

	encodedStore := map[string]string{}
	for k, v := range kvStore {
		encodedKey := base64.URLEncoding.EncodeToString([]byte(k))
		encodedValue := base64.URLEncoding.EncodeToString([]byte(v))
		encodedStore[encodedKey] = encodedValue
	}

	fileContents, _ := json.Marshal(encodedStore)
	os.WriteFile(StoragePath+"/data.json", fileContents, 0644)

	testCases := []struct {
		in  string
		out string
		err error
	}{
		{"key1", "value1", nil},
		{"key2", "value2", nil},
		{"key3", "", nil},
	}

	for _, testCase := range testCases {
		got, err := Get(context.Background(), testCase.in)
		if err != nil {
			t.Errorf("Received unexpected error %s", err)
		}
		if got != testCase.out {
			t.Errorf("Got %s, expected %s", got, testCase.out)
		}
	}
}

func makeStorage(t *testing.T) {
	err := os.Mkdir("testdata", 0755)
	if err != nil && !os.IsExist(err) {
		t.Fatalf("Couldn't create directory testdata: %s", err)
	}
	StoragePath = "testdata"
}
func cleanupStorage(t *testing.T) {
	if err := os.RemoveAll(StoragePath); err != nil {
		t.Errorf("Failed to delete storage path: %s", StoragePath)
	}
	StoragePath = "/tmp/kv"
}
