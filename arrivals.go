package main

import (
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"google.golang.org/protobuf/proto"
)

type Arrival struct {
	RouteID     string
	StopID      string
	ArrivalTime time.Time
}

// pass an empty list of stop ids to match on all
//
// think of this as a "low" level method.
// you should construct your own stopIDs elsewhere.
//
// an empty list of stopIDs will return all of them for the passed feed
func GetFutureArrivals(feedURL string, stopIDs []string) ([]Arrival, error) {
	resp, err := http.Get(feedURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Bad request: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	feed := &gtfs.FeedMessage{}
	if err := proto.Unmarshal(body, feed); err != nil {
		return nil, fmt.Errorf("failed to parse feed: %w", err)
	}

	var arrivals []Arrival
	for _, entity := range feed.Entity {
		if entity.TripUpdate == nil {
			continue
		}
		tripUpdate := entity.TripUpdate
		routeID := ""
		if tripUpdate.Trip != nil && tripUpdate.Trip.RouteId != nil {
			routeID = *tripUpdate.Trip.RouteId
		}

		// Check all stop time updates for this trip
		for _, stopTime := range tripUpdate.StopTimeUpdate {
			stopID := *stopTime.StopId
			// we match on everything if we have 0 stopids
			if stopTime.StopId == nil || len(stopIDs) != 0 && !slices.Contains(stopIDs, stopID) {
				continue
			}

			// Get arrival time
			var arrivalTime time.Time
			if stopTime.Arrival != nil && stopTime.Arrival.Time != nil {
				arrivalTime = time.Unix(*stopTime.Arrival.Time, 0)
			} else if stopTime.Departure != nil && stopTime.Departure.Time != nil {
				arrivalTime = time.Unix(*stopTime.Departure.Time, 0)
			} else {
				continue
			}

			// only get future arrivals
			now := time.Now()
			if arrivalTime.Before(now) {
				continue
			}

			arrivals = append(arrivals, Arrival{
				RouteID:     routeID,
				StopID:      stopID,
				ArrivalTime: arrivalTime,
			})
		}

	}
	return arrivals, nil
}
