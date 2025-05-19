package main

import (
	pages "hmdogan00/email-client/pages"
	partials "hmdogan00/email-client/partials"
	"html/template"
	"log"
	"net/http"
)

var tmpl *template.Template

func main() {
	// parse partial templates
	tmpl = template.Must(template.ParseGlob("templates/partials/*.html"))
	// Initialize handlers
	partialHandler := partials.NewHandler(tmpl)

	// Define HTTP handlers
	http.HandleFunc("/", pages.Index)
	http.HandleFunc("/mails", pages.Mails)
	http.HandleFunc("/get-time", partialHandler.GetTime)

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
