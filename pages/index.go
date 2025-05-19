package pages

import (
	"html/template"
	"log"
	"net/http"
)

type PageData struct {
	Title   string
	Message string
}

type Handler struct {
	Tmpl *template.Template
}

func NewHandler(tmpl *template.Template) *Handler {
	return &Handler{Tmpl: tmpl}
}

// Index serves the main index page.
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:   "Boilerplate",
		Message: "Welcome to Boilerplate! Its a highly customizable email client.",
	}

	err := h.Tmpl.ExecuteTemplate(w, "pages/index.html", data)
	if err != nil {
		log.Printf("Error executing index template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}