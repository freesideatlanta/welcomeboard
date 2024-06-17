package content

import (
	"bytes"
	"embed"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"mime"
	"os"
	"strings"
	"text/template"
	"time"
)

//go:embed templates
var templatesFS embed.FS

var templates *template.Template

//go:embed greetings
var greetings embed.FS

func init() {
	rand.Seed(time.Now().Unix())
	var err error
	templates, err = template.ParseFS(templatesFS, "templates/*")
	if err != nil {
		panic(err)
	}
}

type ContentRouter struct {
	meetupAPIKey string
}

type Event struct {
	Title       string
	Description string
	ImageB64    string
}

func NewContentRouter(meetupAPIKey string) *ContentRouter {
	r := ContentRouter{
		meetupAPIKey: meetupAPIKey,
	}
	go r.startGettingEvents()
	return &r
}

func (c *ContentRouter) startGettingEvents() {
	ticker := time.NewTicker(10 * time.Minute)
	c.UpdateEvents()
	for {
		<-ticker.C
		c.UpdateEvents()
	}

}

func (c *ContentRouter) UpdateEvents() {

}

func (c *ContentRouter) getTimelyEvent() *Event {
	return nil
	/*
		req, err := http.NewRequest(http.MethodGet, "https://filesamples.com/samples/image/png/sample_640%C3%97426.png", nil)
		if err != nil {
			fmt.Println("oops: ", err.Error())
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("oops2: ", err.Error())
		}
		imageData, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("oops3: ", err.Error())
		}
		imgenc := base64.StdEncoding.EncodeToString(imageData)
		return &Event{
			Title:       "Doing stuff",
			Description: "We here at freeside are doing stuff and things! wow! holy crap!",
			ImageB64:    imgenc,
		}
	*/
}

type ImageInfo struct {
	Mime string
	Data string
}

func (c *ContentRouter) GetContent() []byte {
	event := c.getTimelyEvent()
	if event == nil {
		return genericGreeting()
	}
	return c.eventGreeting(event)
}

func (c *ContentRouter) eventGreeting(e *Event) []byte {
	buf := bytes.NewBuffer([]byte{})
	templates.ExecuteTemplate(buf, "eventnow.template.html", e)
	return buf.Bytes()
}

func genericGreeting() []byte {
	items, err := greetings.ReadDir("greetings")
	if err != nil {
		fmt.Println("oops4: ", err.Error())
	}
	n := rand.Int() % len(items)
	image := items[n]

	splitName := strings.Split(image.Name(), ".")
	extension := splitName[len(splitName)-1]
	mimeType := mime.TypeByExtension(extension)
	f, err := greetings.Open("greetings" + string(os.PathSeparator) + image.Name())
	if err != nil {
		fmt.Println("oops5: ", err.Error())
	}
	defer f.Close()
	buf := bytes.NewBuffer([]byte{})
	_, err = io.Copy(buf, f)
	if err != nil {
		fmt.Println("oops6: ", err.Error())
	}
	imgenc := base64.StdEncoding.EncodeToString(buf.Bytes())
	buf2 := bytes.NewBuffer([]byte{})
	templates.ExecuteTemplate(buf2, "photo.template.html", ImageInfo{
		Mime: mimeType,
		Data: imgenc,
	})
	return buf2.Bytes()
}
