package main

import (
	"html/template"
	"log"
	"net/http"
)

func serveRoot(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)

	if r.URL.Path != "/" {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tmplt, err := template.ParseFiles("home.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmplt.Execute(w, struct {
		CurrentWatts int
	} {
		CurrentWatts: 5,
	})
}

func main() {
	http.HandleFunc("/", serveRoot)

	addr := "localhost:9090"

	log.Printf("Serving on %s\n", addr)
	err := http.ListenAndServe("localhost:9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
