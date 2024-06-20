package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/freesideatlanta/welcomeboard/content"
)

//go:embed static
var staticFS embed.FS

//go:embed templates
var templatesFS embed.FS

var templates *template.Template

var meetupKey string
var meetupSecret string
var pemString string
var memberID string
var consumerKey string

func init() {
	var err error
	templates, err = template.ParseFS(templatesFS, "templates/*")
	if err != nil {
		panic(err)
	}

	meetupKey = os.Getenv("MEETUP_KEY")
	if meetupKey == "" {
		panic("no meetup key provided (env var MEETUP_KEY)")
	}
	meetupSecret = os.Getenv("MEETUP_SECRET")
	if meetupKey == "" {
		panic("no meetup secret provided (env var MEETUP_SECRET)")
	}
	pemString = os.Getenv("MEETUP_PEM_STRING")
	if pemString == "" {
		panic("no meetup pem string provided (env var MEETUP_PEM_STRING)")
	}
	memberID = os.Getenv("MEETUP_MEMBER_ID")
	if memberID == "" {
		panic("no meetup memberID string provided (env var MEETUP_MEMBER_ID)")
	}
	consumerKey = os.Getenv("MEETUP_CONSUMER_KEY")
	if consumerKey == "" {
		panic("no meetup pem string provided (env var MEETUP_CONSUMER_KEY)")
	}

}

func main() {
	router := content.NewContentRouter(meetupKey, meetupSecret, pemString, memberID, consumerKey)

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
