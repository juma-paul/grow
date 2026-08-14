package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	log.Println("listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
