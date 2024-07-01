package main

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"

	"github.com/freesideatlanta/welcomeboard/content"
	toml "github.com/pelletier/go-toml/v2"
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

type Config struct {
	MeetupKey    string
	MeetupSecret string
	PemString    string
	MemberID     string
	ConsumerKey  string
}

func init() {
	var cfg Config
	cfgFile, err := os.OpenFile("./config.toml", os.O_RDONLY, 0777)
	if err != nil {
		panic("unable to read config.toml " + err.Error())
	}
	cfgFileContents, err := io.ReadAll(cfgFile)
	if err != nil {
		panic("unable to read config.toml file contents " + err.Error())
	}
	err = toml.Unmarshal(cfgFileContents, &cfg)
	if err != nil {
		panic("unable to decode config.toml file contents " + err.Error())
	}

	templates, err = template.ParseFS(templatesFS, "templates/*")
	if err != nil {
		panic(err)
	}

	meetupKey = cfg.MeetupKey
	if meetupKey == "" {
		panic("no meetup key provided (MeetupKey in toml)")
	}
	meetupSecret = cfg.MeetupSecret
	if meetupKey == "" {
		panic("no meetup secret provided (MeetupSecret in toml")
	}
	pemString = cfg.PemString
	if pemString == "" {
		panic("no meetup pem string provided (PemString in toml)")
	}
	memberID = cfg.MemberID
	if memberID == "" {
		panic("no meetup memberID string provided (MemberID in toml)")
	}
	consumerKey = cfg.ConsumerKey
	if consumerKey == "" {
		panic("no meetup pem string provided (ConsumerKey in toml)")
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
