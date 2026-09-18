package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(home))
	mux.Handle("/snippet/view", http.HandlerFunc(snippetView))
	mux.Handle("/snippet/create", http.HandlerFunc(snippetCreate))

	log.Println("Server starting on :8080")	
	err := http.ListenAndServe(":8080", mux)

	if err != nil {
		log.Fatal(err)
	}
}
