package eventgreeting

import (
	"bytes"
	"crypto/x509"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"html/template"
	"io"
	"math"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

//go:embed templates/eventnow.template.html
var eventTemplateRaw string

var eventTemplate *template.Template

func init() {
	eventTemplate = template.Must(template.New("event greeting").Parse(eventTemplateRaw))
}

func NewEventGreeter(meetupKey string, meetupSecret string, pemString string, memberID string, consumerKey string) *EventGreeter {
	r := EventGreeter{
		meetupKey:    meetupKey,
		meetupSecret: meetupSecret,
		pemString:    pemString,
		memberID:     memberID,
		consumerKey:  consumerKey,
	}
	return &r
}

func (c *EventGreeter) StartGettingEvents() {
	ticker := time.NewTicker(10 * time.Minute)
	c.CacheEvents()
	for {
		<-ticker.C
		c.CacheEvents()
	}

}

func (c *EventGreeter) CacheEvents() {
	tok := c.GetLoginToken()
	c.TimelyEvent = c.getTimelyEvent(tok)
}
func (c *EventGreeter) GetLoginToken() string {
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
func (c *EventGreeter) getTimelyEvent(apiToken string) *Event {
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
	ret := events[relevantEventIndex]
	// get image base64 data
	req, err = http.NewRequest(http.MethodGet, ret.ImageURL, nil)
	if err != nil {
		fmt.Println("oops: ", err.Error())
	}
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("oops2: ", err.Error())
	}
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("oops3: ", err.Error())
	}
	ret.ImageB64 = base64.StdEncoding.EncodeToString(imageData)
	return ret
}

func EventGreeting(e *Event) []byte {
	buf := bytes.NewBuffer([]byte{})
	splitName := strings.Split(e.ImageURL, ".")
	extension := splitName[len(splitName)-1]
	mimeType := mime.TypeByExtension(extension)
	eventContent := EventGreetingData{
		Title:       e.Title,
		Description: e.Description,
		Data:        e.ImageB64,
		Mime:        mimeType,
		IsHappening: !time.Now().Before(e.Time),
		DeltaTime:   strconv.Itoa(int(math.Abs(float64(time.Since(e.Time).Minutes())))),
	}
	err := eventTemplate.Execute(buf,eventContent)
	if err != nil{
		fmt.Printf("unable to execute template: %s\n", err.Error())
	}
	return buf.Bytes()
}
