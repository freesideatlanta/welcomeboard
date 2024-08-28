package content

import (
	"time"

	eventgreeting "github.com/freesideatlanta/welcomeboard/content/eventGreeting"
	photowelcome "github.com/freesideatlanta/welcomeboard/content/photoWelcome"
)

const eventHorizon time.Duration = time.Minute * 30

type ContentRouter struct {
	eventGreeter *eventgreeting.EventGreeter
}

func NewContentRouter(meetupKey string, meetupSecret string, pemString string, memberID string, consumerKey string) *ContentRouter {
	r := ContentRouter{
		eventGreeter: eventgreeting.NewEventGreeter(meetupKey, meetupSecret, pemString, memberID, consumerKey),
	}
	r.eventGreeter.CacheEvents()
	go r.eventGreeter.StartGettingEvents()
	return &r
}

func (c *ContentRouter) GetContent() []byte {
	event := c.eventGreeter.TimelyEvent
	// if we are more than an hour out or the event doesn't exist
	if event == nil || time.Now().Before(event.Time.Add(-eventHorizon)) {
		return photowelcome.PhotoGreeting()
	}
	return eventgreeting.EventGreeting(event)
}
