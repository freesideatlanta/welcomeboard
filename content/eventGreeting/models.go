package eventgreeting

import "time"

type Event struct {
	Time        time.Time
	EndTime     time.Time
	Title       string
	Description string
	ImageURL    string
	ImageB64    string
}

type EventGreetingData struct {
	Title       string
	Description string
	Mime        string
	Data        string
	DeltaTime   string
	IsHappening bool
}

type EventGreeter struct {
	TimelyEvent  *Event
	meetupKey    string
	meetupSecret string
	pemString    string
	memberID     string
	consumerKey  string
}





