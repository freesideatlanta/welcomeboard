package content

import (
	"bytes"
	"crypto/x509"
	"embed"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/rand"
	"mime"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

//go:embed templates
var templatesFS embed.FS

var templates *template.Template

//go:embed greetings
var greetings embed.FS

func init() {
	var err error
	templates, err = template.ParseFS(templatesFS, "templates/*")
	if err != nil {
		panic(err)
	}
}

type ContentRouter struct {
	timelyEvent  *Event
	meetupKey    string
	meetupSecret string
	pemString    string
	memberID     string
	consumerKey  string
}

type Event struct {
	Time        time.Time
	EndTime     time.Time
	Title       string
	Description string
	ImageURL    string
	ImageB64    string
}

func NewContentRouter(meetupKey string, meetupSecret string, pemString string, memberID string, consumerKey string) *ContentRouter {
	r := ContentRouter{
		meetupKey:    meetupKey,
		meetupSecret: meetupSecret,
		pemString:    pemString,
		memberID:     memberID,
		consumerKey:  consumerKey,
	}
	r.CacheEvents()
	go r.startGettingEvents()
	return &r
}

func (c *ContentRouter) startGettingEvents() {
	ticker := time.NewTicker(10 * time.Minute)
	c.CacheEvents()
	for {
		<-ticker.C
		c.CacheEvents()
	}

}

func (c *ContentRouter) CacheEvents() {
	tok := c.GetLoginToken()
	c.timelyEvent = c.getTimelyEvent(tok)
}

func (c *ContentRouter) GetLoginToken() string {
	tok := jwt.New(jwt.SigningMethodRS256)
	tok.Header["kid"] = c.consumerKey
	claims := tok.Claims.(jwt.MapClaims)
	claims["sub"] = c.memberID
	claims["iss"] = c.meetupKey
	claims["aud"] = "api.meedup.com"
	claims["exp"] = time.Now().Add(time.Second * 120).Unix()
	block, _ := pem.Decode([]byte(c.pemString))
	key, _ := x509.ParsePKCS1PrivateKey(block.Bytes)

	tokenString, err := tok.SignedString(key)
	if err != nil {
		fmt.Println("oofskies: " + err.Error())
		return ""
	}

	req, err := http.NewRequest(http.MethodPost, "https://secure.meetup.com/oauth2/access", strings.NewReader(fmt.Sprintf(`grant_type=urn:ietf:params:oauth:grant-type:jwt-bearer&assertion=%s`, tokenString)))
	if err != nil {
		fmt.Println("sjdfklsjklfjs: ", err.Error())
	}
	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	req.Header.Add("user-agent", "curl/8.6.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("riperooni: ", err.Error())
		return ""
	}

	apiret := apiJwtReturn{}
	err = json.NewDecoder(resp.Body).Decode(&apiret)
	if err != nil {
		fmt.Println("fuuuuuuck: ", err.Error())
		return ""
	}
	return apiret.AccessToken

}

type apiJwtReturn struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type graphQlQuery struct {
	Query string `json:"query"`
}

func (c *ContentRouter) getTimelyEvent(apiToken string) *Event {
	q := `query {
  groupByUrlname(urlname: "freeside-atlanta") {
    unifiedEvents(
      sortOrder: ASC
    ) {
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
      edges {
        cursor
        node {
          id
          title
          shortDescription
          imageUrl
          dateTime
          endTime
        }
      }
    }
  }
}`
	queryBytes, err := json.Marshal(graphQlQuery{Query: q})
	if err != nil {
		fmt.Println("unable to marshal graphQlQuery: ", err.Error())
	}
	req, err := http.NewRequest(http.MethodPost, "https://api.meetup.com/gql", bytes.NewBuffer(queryBytes))
	if err != nil {
		fmt.Println("unable to generate request to meetup api: ", err.Error())
	}
	req.Header.Add("content-type", "application/json")
	req.Header.Add("authorization", "Bearer "+apiToken)
	req.Header.Add("user-agent", "curl/8.6.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("error querying graphql api: " + err.Error())
	}

	var res UnPaginatedResult
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		fmt.Println("unable to decode api respone: ", err.Error())
	}

	events, err := flatten(res)
	if err != nil {
		fmt.Println("unable to flatten api response: ", err.Error())
	}

	relevantEventIndex := 0
	for i := 0; i < len(events); i++ {
		if events[i].Time.After(time.Now()) || events[i].EndTime.After(time.Now()) {
			relevantEventIndex = i
			break
		}
	}
	return events[relevantEventIndex]
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
	event := c.timelyEvent
	// if we are more than an hour out or the event doesn't exist
	if event == nil || time.Now().Before(event.Time.Add(-time.Hour*1)) {
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
