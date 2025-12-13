package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func arrivals(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		stationId := r.PathValue("station")
		if stationId == "" {
			http.Error(w, "station id required", http.StatusBadRequest)
			return
		}
		direction := r.URL.Query().Get("direction")

		var stopIDs []string
		if direction == "" {
			stopIDs = []string{
				fmt.Sprintf("%s%s", stationId, "N"),
				fmt.Sprintf("%s%s", stationId, "S"),
			}
		} else {
			stopIDs = []string{fmt.Sprintf("%s%s", stationId, direction)}
		}

		var arrivals []Arrival
		for _, stopId := range stopIDs {
			a, err := GetFutureArrivals(stopId)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			arrivals = append(arrivals, a...)
		}

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(arrivals)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func Serve(port int) error {
	addr := fmt.Sprintf(":%d", port)
	http.HandleFunc("/arrivals/{station}", arrivals)
	return http.ListenAndServe(addr, nil)
}
