package partials

import (
	"html/template"
	"log"
	"net/http"
	"time"
)

type timePartialData struct {
	Time    string
}

type handler struct {
	Tmpl *template.Template
}

func TimeHandler(tmpl *template.Template) *Handler {
	return &Handler{Tmpl: tmpl}
}

func (h *Handler) GetTime(w http.ResponseWriter, r *http.Request) {
	currentTime := time.Now().Format("15:04:05 MST")
	data := timePartialData{
		Time: currentTime,
	}

	err := h.Tmpl.ExecuteTemplate(w, "time-update.html", data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}