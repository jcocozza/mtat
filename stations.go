package main

import (
	"embed"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
)

type SubwayRealTimeFeedURL string

const (
	SRTF_ACESr    SubwayRealTimeFeedURL = "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-ace"
	SRTF_G        SubwayRealTimeFeedURL = "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-g"
	SRTF_NQRW     SubwayRealTimeFeedURL = "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-nqrw"
	SRTF_1234567S SubwayRealTimeFeedURL = "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs"
	SRTF_BDFMSf   SubwayRealTimeFeedURL = "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-bdfm"
	SRTF_JZ       SubwayRealTimeFeedURL = "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-jz"
	SRTF_L        SubwayRealTimeFeedURL = "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-l"
	SRTF_SIR      SubwayRealTimeFeedURL = "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-si"
)

type Subway string

const (
	SUBWAY_A  Subway = "A"
	SUBWAY_C  Subway = "C"
	SUBWAY_E  Subway = "E"
	SUBWAY_Sr Subway = "Sr"

	SUBWAY_N Subway = "N"
	SUBWAY_Q Subway = "Q"
	SUBWAY_R Subway = "R"
	SUBWAY_W Subway = "W"

	SUBWAY_1 Subway = "1"
	SUBWAY_2 Subway = "2"
	SUBWAY_3 Subway = "3"
	SUBWAY_4 Subway = "4"
	SUBWAY_5 Subway = "5"
	SUBWAY_6 Subway = "6"
	SUBWAY_7 Subway = "7"
	SUBWAY_S Subway = "S"

	SUBWAY_B  Subway = "B"
	SUBWAY_D  Subway = "D"
	SUBWAY_F  Subway = "F"
	SUBWAY_M  Subway = "M"
	SUBWAY_Sf Subway = "Sf"

	SUBWAY_J Subway = "J"
	SUBWAY_Z Subway = "Z"

	SUBWAY_L Subway = "L"

	SUBWAY_SIR Subway = "SI"
)

var SubwayFeedMap = map[Subway]SubwayRealTimeFeedURL{
	// ACE / Sr
	SUBWAY_A:  SRTF_ACESr,
	SUBWAY_C:  SRTF_ACESr,
	SUBWAY_E:  SRTF_ACESr,
	SUBWAY_Sr: SRTF_ACESr,

	// NQRW
	SUBWAY_N: SRTF_NQRW,
	SUBWAY_Q: SRTF_NQRW,
	SUBWAY_R: SRTF_NQRW,
	SUBWAY_W: SRTF_NQRW,

	// 1234567S
	SUBWAY_1: SRTF_1234567S,
	SUBWAY_2: SRTF_1234567S,
	SUBWAY_3: SRTF_1234567S,
	SUBWAY_4: SRTF_1234567S,
	SUBWAY_5: SRTF_1234567S,
	SUBWAY_6: SRTF_1234567S,
	SUBWAY_7: SRTF_1234567S,
	SUBWAY_S: SRTF_1234567S,

	// BDFM / Sf
	SUBWAY_B:  SRTF_BDFMSf,
	SUBWAY_D:  SRTF_BDFMSf,
	SUBWAY_F:  SRTF_BDFMSf,
	SUBWAY_M:  SRTF_BDFMSf,
	SUBWAY_Sf: SRTF_BDFMSf,

	// JZ
	SUBWAY_J: SRTF_JZ,
	SUBWAY_Z: SRTF_JZ,

	// L
	SUBWAY_L: SRTF_L,

	// SIR
	SUBWAY_SIR: SRTF_SIR,
}

type Station struct {
	Id   string
	Name string
	// direction -> stop
	Stops map[string]Stop
}

type Stop struct {
	Id        string
	Direction string
	Latitude  float64
	Longitude float64
}

// return station id, direction, and error
//
// direction can be empty
func ParseStopId(id string) (string, string, error) {
	if len(id) == 0 {
		return "", "", fmt.Errorf("unable to parse empty string")
	}
	lastChar := id[len(id)-1:]
	stationId := id[:len(id)-1]

	// this is a switch because i will want some more elaborate parsing later
	switch lastChar {
	case "N":
		return stationId, lastChar, nil
	case "S":
		return stationId, lastChar, nil
	default:
		return stationId, "", nil
	}
}

//go:embed gtfs_subway/stops.txt
//go:embed gtfs_subway/stop_times.txt
//go:embed gtfs_subway/trips.txt
var gtfsSubway embed.FS

// return map [station id]-> station
//
// read and parse stops.txt provided by the mta
func ReadStations() (map[string]Station, error) {
	f, err := gtfsSubway.Open("gtfs_subway/stops.txt")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	rdr := csv.NewReader(f)
	header, err := rdr.Read()
	if err != nil {
		return nil, err
	}

	cols := make(map[string]int)
	for i, colName := range header {
		cols[colName] = i
	}

	stations := make(map[string]Station)
	for {
		row, err := rdr.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}

		stopId := row[cols["stop_id"]]
		name := row[cols["stop_name"]]
		latStr := row[cols["stop_lat"]]
		lonStr := row[cols["stop_lon"]]
		locationType := row[cols["location_type"]]

		lat, err := strconv.ParseFloat(latStr, 64)
		if err != nil {
			return nil, err
		}
		lon, err := strconv.ParseFloat(lonStr, 64)
		if err != nil {
			return nil, err
		}

		stationId, direction, err := ParseStopId(stopId)
		if err != nil {
			return nil, err
		}

		isParent := locationType == "1"

		if _, exists := stations[stationId]; !exists {
			station := Station{
				Id:    stationId,
				Name:  name,
				Stops: make(map[string]Stop),
			}
			stations[stationId] = station
			if isParent {
				continue
			}
		}
		stop := Stop{
			Id:        stopId,
			Direction: direction,
			Latitude:  lat,
			Longitude: lon,
		}
		stations[stationId].Stops[direction] = stop
	}
	return stations, nil
}

type StopTime struct {
	TripId        string
	StopId        string
	ArrivalTime   string
	DepartureTime string
	StopSequence  string
}

// read the stop_times.txt file
// TODO: this is pathetically slow
func ReadStopTimes() ([]StopTime, error) {
	f, err := gtfsSubway.Open("gtfs_subway/stop_times.txt")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	rdr := csv.NewReader(f)
	header, err := rdr.Read()
	if err != nil {
		return nil, err
	}

	cols := make(map[string]int)
	for i, colName := range header {
		cols[colName] = i
	}

	var stopTimes []StopTime
	for {
		row, err := rdr.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}

		stopTime := StopTime{
			TripId:        row[cols["trip_id"]],
			StopId:        row[cols["stop_id"]],
			ArrivalTime:   row[cols["arrival_time"]],
			DepartureTime: row[cols["departure_time"]],
			StopSequence:  row[cols["stop_sequence"]],
		}

		stopTimes = append(stopTimes, stopTime)
	}
	return stopTimes, nil
}

type Trip struct {
	RouteId      string
	TripId       string
	ServiceId    string
	TripHeadSign string
	DirectionId  string
	ShapeId      string
}

// read the trips.txt file
//
// TODO: this is pathetically slow
func ReadTrips() ([]Trip, error) {
	f, err := gtfsSubway.Open("gtfs_subway/trips.txt")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	rdr := csv.NewReader(f)
	header, err := rdr.Read()
	if err != nil {
		return nil, err
	}

	cols := make(map[string]int)
	for i, colName := range header {
		cols[colName] = i
	}

	var trips []Trip
	for {
		row, err := rdr.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}

		t := Trip{
			RouteId:      row[cols["route_id"]],
			TripId:       row[cols["trip_id"]],
			ServiceId:    row[cols["service_id"]],
			TripHeadSign: row[cols["trip_headsign"]],
			DirectionId:  row[cols["direction_id"]],
			ShapeId:      row[cols["shape_id"]],
		}

		trips = append(trips, t)
	}
	return trips, nil
}

// map subway line (e.g. "A") to its url
//
// map station id (e.g. 122) to its url
//
// returns subway line -> feed, station id -> feed, station id -> line, err
func GetMappings() (map[Subway]SubwayRealTimeFeedURL, map[string]SubwayRealTimeFeedURL, error) {
	trips, err := ReadTrips()
	if err != nil {
		return nil, nil, err
	}

	stopTimes, err := ReadStopTimes()
	if err != nil {
		return nil, nil, err
	}

	// stop id -> trip ids
	stopToTrip := make(map[string][]string)
	for _, stopTime := range stopTimes {
		stationId, _, err := ParseStopId(stopTime.StopId)
		if err != nil {
			return nil, nil, err
		}
		stopToTrip[stationId] = append(stopToTrip[stationId], stopTime.TripId)
	}

	// trip id -> route id
	tripToRoute := make(map[string]string)
	for _, trip := range trips {
		tripToRoute[trip.TripId] = trip.RouteId
	}

	lineToFeed := make(map[Subway]SubwayRealTimeFeedURL)
	stationToFeed := make(map[string]SubwayRealTimeFeedURL)
	for stopId, tripIds := range stopToTrip {
		for _, tripId := range tripIds {
			if routeId, ok := tripToRoute[tripId]; ok {
				if url, ok := SubwayFeedMap[Subway(routeId)]; ok {
					lineToFeed[Subway(routeId)] = SubwayRealTimeFeedURL(url)
					stationToFeed[stopId] = SubwayRealTimeFeedURL(url)
				}
			}
		}
	}
	return lineToFeed, stationToFeed, nil
}
