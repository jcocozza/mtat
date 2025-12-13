package main

import (
	"fmt"
	"io"
	"net/http"
	"slices"
	"sort"
	"time"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"google.golang.org/protobuf/proto"
)

type Arrival struct {
	// this is the subway line
	RouteID string `json:"route_id"`
	// station id + direction
	StopID string `json:"stop_id"`
	// N or S
	Direction   string    `json:"direction"`
	ArrivalTime time.Time `json:"arrival_time"`
}

// sort arrivals from soonest to latest (in place)
func SortArrivals(arrivals []Arrival) {
	sort.Slice(arrivals, func(i, j int) bool {
		return arrivals[i].ArrivalTime.Before(arrivals[j].ArrivalTime)
	})
}

func getFeed(url SubwayRealTimeFeedURL) (*gtfs.FeedMessage, error) {
	resp, err := http.Get(string(url))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad request: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	feed := &gtfs.FeedMessage{}
	if err := proto.Unmarshal(body, feed); err != nil {
		return nil, fmt.Errorf("failed to parse feed: %w", err)
	}
	return feed, nil
}

// extract the future arrivals for the given stopIDs
//
// note that if stopID isn't in the passed feed, nothing will be returned for that id
func futureArrivals(feed *gtfs.FeedMessage, stopIDs []string) ([]Arrival, error) {
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
			_, direction, err := ParseStopId(stopID)
			if err != nil {
				return nil, err
			}
			arrivals = append(arrivals, Arrival{
				RouteID:     routeID,
				StopID:      stopID,
				Direction:   direction,
				ArrivalTime: arrivalTime,
			})
		}
	}
	SortArrivals(arrivals)
	return arrivals, nil
}

// get the next hour of arrivals for a given stop
func GetFutureArrivals(stopID string) ([]Arrival, error) {
	_, stationIds, err := GetMappings()
	if err != nil {
		return nil, err
	}
	stationId, _, err := ParseStopId(stopID)
	if err != nil {
		return nil, err
	}
	url, ok := stationIds[stationId]
	if !ok {
		return nil, fmt.Errorf("unknown stop id: %s. no associated feed url", stopID)
	}
	feed, err := getFeed(url)
	if err != nil {
		return nil, err
	}
	return futureArrivals(feed, []string{stopID})
}
