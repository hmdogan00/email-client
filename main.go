package main

import (
	pages "hmdogan00/email-client/pages"
	partials "hmdogan00/email-client/partials"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var tmpl *template.Template

func main() {
	var err error

	tmpl, err = findAndParseTemplates("templates", template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("DD/MM/YYYY HH:mm:ss")
		},
	})
	print("Templates parsed successfully\n")
	for _, t := range tmpl.Templates() {
		print(t.Name() + "\n")
	}
	if err != nil {
		log.Fatalf("Error parsing templates: %v", err)
	}

	// Initialize handlers
	pageHandler := pages.NewHandler(tmpl)
	partialHandler := partials.NewHandler(tmpl)

	// Define HTTP handlers
	http.HandleFunc("/", pageHandler.Index)
	http.HandleFunc("/get-time", partialHandler.GetTime)

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func findAndParseTemplates(rootDir string, funcMap template.FuncMap) (*template.Template, error) {
    cleanRoot := filepath.Clean(rootDir)
    pfx := len(cleanRoot)+1
    root := template.New("")

    err := filepath.Walk(cleanRoot, func(path string, info os.FileInfo, e1 error) error {
        if !info.IsDir() && strings.HasSuffix(path, ".html") {
            if e1 != nil {
                return e1
            }

            b, e2 := os.ReadFile(path)
            if e2 != nil {
                return e2
            }

            name := path[pfx:]
            t := root.New(name).Funcs(funcMap)
            _, e2 = t.Parse(string(b))
            if e2 != nil {
                return e2
            }
        }

        return nil
    })

    return root, err
}