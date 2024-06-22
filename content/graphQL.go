package content

import (
	"fmt"
	"time"
)

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

		startTime, err := time.Parse(time.RFC3339, rawEvent.DateTime)
		if err != nil {
			fmt.Println("unable to parse time: ", err.Error())
			return nil, err
		}
		endTime, err := time.Parse(time.RFC3339, rawEvent.DateTime)
		if err != nil {
			fmt.Println("unable to parse time: ", err.Error())
			return nil, err
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
