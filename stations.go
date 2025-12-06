package main

import (
	"fmt"
	"io"
	"os"
	"errors"
	"strconv"
	"encoding/csv"
)

type Station struct {
	Id string
	Name string
	// direction -> stop
	Stops map[string]Stop	
}

type Stop struct {
	Id string
	Direction string
	Latitude float64
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

// return map [station id]-> station
func ReadStops(path string) (map[string]Station, error) {
	f, err := os.Open(path)
	if err != nil { return nil, err }
	rdr := csv.NewReader(f)
	header, err := rdr.Read()
	if err != nil {
		return nil, err
	}

	cols := make(map[string]int)
	for i, colName := range header {
		cols[colName]	= i
	}

	stations := make(map[string]Station)
	for {
		row, err := rdr.Read()
		if errors.Is(io.EOF, err) {
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
		if err != nil { return nil, err }
		lon, err := strconv.ParseFloat(lonStr, 64)
		if err != nil { return nil, err }

		stationId, direction, err := ParseStopId(stopId)
		if err != nil { return nil, err }

		isParent := locationType == "1"

		if _, exists := stations[stationId]; !exists {
			station := Station{
				Id: stationId,
				Name: name,
				Stops: make(map[string]Stop),
			}
			stations[stationId] = station
			if isParent { continue }
		}
		stop := Stop{
			Id: stopId,
			Direction: direction,
			Latitude: lat,
			Longitude: lon,
		}
		stations[stationId].Stops[direction] = stop
	}
	return stations, nil
}
