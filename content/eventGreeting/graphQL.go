package eventgreeting

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type apiJwtReturn struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type graphQlQuery struct {
	Query string `json:"query"`
}



type UnPaginatedResult struct {
	Data struct {
		GroupByUrlname struct {
			UnifiedEvents struct {
				PageInfo struct {
					StartCursor     string `json:"startCursor"`
					EndCursor       string `json:"endCursor"`
					HasNextPage     bool   `json:"hasNextPage"`
					HasPreviousPage bool   `json:"hasPreviousPage"`
				} `json:"pageInfo"`
				Edges []struct {
					Cursor string `json:"cursor"`
					Node   struct {
						ID               string      `json:"id"`
						Title            string      `json:"title"`
						ShortDescription interface{} `json:"shortDescription"`
						ImageURL         string      `json:"imageUrl"`
						DateTime         string      `json:"dateTime"`
						EndTime          string      `json:"endTime"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"unifiedEvents"`
		} `json:"groupByUrlname"`
	} `json:"data"`
}

func flatten(raw UnPaginatedResult) ([]*Event, error) {
	var ret []*Event
	for _, edge := range raw.Data.GroupByUrlname.UnifiedEvents.Edges {
		rawEvent := edge.Node
		startTime, err := NewParsetime(rawEvent.DateTime)
		if err != nil {
			return nil, fmt.Errorf("unable to parse time: %s", err.Error())
		}

		endTime, err := NewParsetime(rawEvent.DateTime)
		if err != nil {
			return nil, fmt.Errorf("unable to parse time: %s", err.Error())
		}

		ret = append(ret, &Event{
			Title:    rawEvent.Title,
			ImageURL: rawEvent.ImageURL,
			ImageB64: "",
			Time:     startTime,
			EndTime:  endTime,
		})
	}
	return ret, nil
}

func NewParsetime(in string) (time.Time, error) {
	//reg := regexp.MustCompile(`([0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}):([+-][0-9]{2}:[0-9]{2})`)
	reg := regexp.MustCompile(`(^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2})([\+-][0-9]{2}:[0-9]{2})`)
	submatches := reg.FindSubmatch([]byte(in))
	if len(submatches) != 3 {
		return time.Time{}, errors.New("time parsing got f*cked up")
	}
	t, err := time.Parse("2006-01-02T15:04", string(submatches[1]))
	if err != nil {
		return time.Time{}, fmt.Errorf("oopsiedoopsie: %s", err.Error())
	}
	offsetString := string(submatches[2])
	multiple := 0
	switch offsetString[0] {
	case '-':
		multiple = 1
	case '+':
		multiple = -1
	}
	splitOffset := strings.Split(offsetString[1:], ":")
	hours, err := strconv.Atoi(splitOffset[0])
	if err != nil {
		return time.Time{}, fmt.Errorf("unable to get hours of time offset: %s", err.Error())
	}
	minutes, err := strconv.Atoi(splitOffset[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("unable to get minutes of time offset: %s", err.Error())
	}

	offset := time.Duration(multiple) * (time.Hour*time.Duration(hours) + time.Minute*time.Duration(minutes))
	ret := t.Add(offset)

	return ret, nil
}

