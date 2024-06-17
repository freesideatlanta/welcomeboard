package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"

	"github.com/freesideatlanta/welcomeboard/content"
)

//go:embed static
var staticFS embed.FS

//go:embed templates
var templatesFS embed.FS

var templates *template.Template

func init() {
	var err error
	templates, err = template.ParseFS(templatesFS, "templates/*")
	if err != nil {
		panic(err)
	}
}

func main() {
	router := content.NewContentRouter("meetupAPIKey")

	http.HandleFunc("/", indexPage)
	http.HandleFunc("/update", updateFunc(router))
	http.Handle("/static/", http.FileServerFS(staticFS))
	http.ListenAndServe(":8080", nil)
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "index.templ.html", nil)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func updateFunc(cr *content.ContentRouter) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write(cr.GetContent())
	}
}
