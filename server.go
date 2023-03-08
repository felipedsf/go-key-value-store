package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi"
)

var data = map[string]string{}

func main() {
	port := "9080"
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

func Get(ctx context.Context, key string) (string, error) {
	data, err := loadData(ctx)
	if err != nil {
		return "", err
	}
	return data[key], nil
}

func Delete(ctx context.Context, key string) error {
	data, err := loadData(ctx)
	if err != nil {
		return err
	}
	delete(data, key)
	return saveData(ctx, data)
}

func Set(ctx context.Context, key string, value string) error {
	data, err := loadData(ctx)
	if err != nil {
		return err
	}
	data[key] = value
	return saveData(ctx, data)
}

var StoragePath = "/tmp"

func dataPath() string {
	return filepath.Join(StoragePath, "data.json")
}

func loadData(ctx context.Context) (map[string]string, error) {
	empty := map[string]string{}
	emptyData, err := encode(map[string]string{})
	if err != nil {
		return empty, err
	}

	if _, err := os.Stat(StoragePath); os.IsNotExist(err) {
		err = os.MkdirAll(StoragePath, 0755)
		if err != nil {
			return empty, err
		}
	}

	if _, err := os.Stat(dataPath()); os.IsNotExist(err) {
		err := os.WriteFile(dataPath(), emptyData, 0644)
		if err != nil {
			return empty, err
		}
	}

	content, err := os.ReadFile(dataPath())
	if err != nil {
		return empty, err
	}
	return decode(content)
}

func saveData(ctx context.Context, data map[string]string) error {
	if _, err := os.Stat(StoragePath); os.IsNotExist(err) {
		err = os.MkdirAll(StoragePath, 0755)
		if err != nil {
			return err
		}
	}
	encodedData, err := encode(data)
	if err != nil {
		return err
	}
	return os.WriteFile(dataPath(), encodedData, 0644)
}

func encode(data map[string]string) ([]byte, error) {
	encodedData := map[string]string{}
	for k, v := range data {
		ek := base64.URLEncoding.EncodeToString([]byte(k))
		ev := base64.URLEncoding.EncodeToString([]byte(v))
		encodedData[ek] = ev
	}
	return json.Marshal(encodedData)
}

func decode(data []byte) (map[string]string, error) {
	var jsonData map[string]string

	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, err
	}
	returnData := map[string]string{}
	for k, v := range jsonData {
		dk, err := base64.URLEncoding.DecodeString(k)
		if err != nil {
			return nil, err
		}
		dv, err := base64.URLEncoding.DecodeString(v)
		if err != nil {
			return nil, err
		}
		returnData[string(dk)] = string(dv)

	}
	return returnData, nil

}
