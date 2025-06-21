package pages

import (
	"fmt"
	"html/template"
	"net/http"
	"time"
)

type ConstantPageData struct {
	Title   string
	Year   	string
    Path    string
}

var year string = fmt.Sprint(time.Now().Year())

func render(w http.ResponseWriter, tmpl string) {
	data := ConstantPageData{
		Title: "Boilerplate",
		Year:  year,
        Path: tmpl,
	}
    // Parse templates
    templates, err := template.ParseFiles(
        "templates/layouts/base.html",
        //"templates/partials/_navbar.html",
		fmt.Sprintf("templates/pages/%s.html", tmpl),
    )
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    // Execute the base layout (which includes nested templates)
    err = templates.ExecuteTemplate(w, "base", data)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func Index(w http.ResponseWriter, r *http.Request) {
    render(w, "index")
}

func Mails(w http.ResponseWriter, r *http.Request) {
	render(w, "mails")
}