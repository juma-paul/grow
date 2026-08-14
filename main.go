package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/juma-paul/grow/internal/server"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	http.HandleFunc("/execute", server.HandleExecute)

	log.Println("listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
