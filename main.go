package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func home (w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	log.Println("Home page requested")
	w.Write([]byte("Welcome to the Home Page! Hello from snippet"))
}
	
func snippetView(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		log.Println("Url Param 'id' is missing")	
		http.NotFound(w, r)
		return
	}

	log.Println("Snippet view requested")
	fmt.Fprintf(w, "Display a specific snippet with ID %d...", id)
}

func snippetCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {

		// w.WriteHeader(405)
		// w.Write([]byte("Method Not Allowed"))
		// return

		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method Not Allowed"))
		return
	}

	// if r.Method != http.MethodPost {
	// 	w.WriteHeader(http.StatusMethodNotAllowed)
	// 	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	// 	return
	// }

	log.Println("Snippet create requested")
	w.Write([]byte("Create new snippet....!"))
}

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
