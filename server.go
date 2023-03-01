package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	log.Println("Starting server!!!")
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprintf(writer, "Hello world\n")
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
