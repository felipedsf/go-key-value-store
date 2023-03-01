package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

var data = map[string]string{}

func main() {
	port := "8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}
	log.Printf("Starting up on http://localhost:%s", port)

	r := chi.NewRouter()
	r.Get("/", func(writer http.ResponseWriter, request *http.Request) {
		JSON(writer, map[string]string{"Hello": "World!!"})
	})

	r.Get("/key/{key}", func(writer http.ResponseWriter, request *http.Request) {
		key := chi.URLParam(request, "key")
		data, err := Get(request.Context(), key)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			JSON(writer, map[string]string{"error: ": err.Error()})
			return
		}
		writer.Write([]byte(data))
	})

	r.Delete("/key/{key}", func(writer http.ResponseWriter, request *http.Request) {
		key := chi.URLParam(request, "key")
		err := Delete(request.Context(), key)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			JSON(writer, map[string]string{"error": err.Error()})
			return
		}
		JSON(writer, map[string]string{"status": "success"})
	})

	r.Post("/key/{key}", func(writer http.ResponseWriter, request *http.Request) {
		key := chi.URLParam(request, "key")
		body, err := io.ReadAll(request.Body)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			JSON(writer, map[string]string{"error": err.Error()})
			return
		}

		err = Set(request.Context(), key, string(body))
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			JSON(writer, map[string]string{"error": err.Error()})
			return
		}
		JSON(writer, map[string]string{"status": "success"})
	})

	log.Fatal(http.ListenAndServe(":"+port, r))
}

func Get(ctx context.Context, key string) (string, error) {
	return data[key], nil
}

func Delete(ctx context.Context, key string) error {
	delete(data, key)
	return nil
}

func Set(context context.Context, key string, value string) error {
	data[key] = value
	return nil
}

// JSON encodes data to json and writes it to the http response.
func JSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	b, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JSON(w, map[string]string{"error": err.Error()})
		return
	}

	w.Write(b)
}
