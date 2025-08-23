package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"tfl"
)

// type CycleStation struct {
// 	Name            string  `json:"name"`
// 	TerminalName    string  `json:"terminal_name"`
// 	DistanceKM      float64 `json:"distance_km"`
// 	BikesAvailable  int     `json:"bikes_available"`
// 	EBikesAvailable int     `json:"ebikes_available"`
// 	DocksAvailable  int     `json:"docks_available"`
// }
//
// type CyclesNearbyResponse struct {
// 	Count                   int            `json:"count"`
// 	Stations                []CycleStation `json:"stations"`
// 	NearestStation          CycleStation   `json:"nearest_station,omitempty"`
// 	NearestStationWithDocks CycleStation   `json:"nearest_station_with_docks,omitempty"`
// 	NearestStationWithBikes CycleStation   `json:"nearest_station_with_bikes,omitempty"`
// }

func main() {
	cfg, err := tfl.Init()
	if err != nil {
		fmt.Println("Failed to initialize config:", err)
		return
	}

	http.HandleFunc("/cycles/nearby/siri", func(w http.ResponseWriter, r *http.Request) {
		coordStr := r.URL.Query().Get("coords")
		parts := strings.Split(coordStr, ",")
		if len(parts) != 2 {
			http.Error(w, "invalid coords parameter", http.StatusBadRequest)
			return
		}

		fmt.Println("Request: /cycles/nearby?coords=", coordStr)

		lat, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			http.Error(w, "invalid latitude", http.StatusBadRequest)
			return
		}

		lng, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			http.Error(w, "invalid longitude", http.StatusBadRequest)
			return
		}

		availability, err := tfl.GetClosestAvailableBikeStations(cfg, lat, lng, 1.0)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(availability.StationsSortedByClosest) == 0 {
			http.Error(w, "no cycle stations found within range", http.StatusNotFound)
			return
		}

		var message string
		// if they're the same, condense
		if availability.NearestStationWithBikes.Name == availability.NearestStationWithDocks.Name && availability.NearestStationWithBikes.Name == availability.NearestStation.Name {
			message = fmt.Sprintf("%s has both bikes & docks", availability.NearestStationWithBikes.Name)
		} else if availability.NearestStationWithBikes.Name == availability.NearestStationWithDocks.Name {
			message = fmt.Sprintf("Closest station: %s. Closest with bikes & docks: %s", availability.NearestStation.Name, availability.NearestStationWithBikes.Name)
		} else if availability.NearestStationWithBikes.Name == availability.NearestStation.Name {
			message = fmt.Sprintf("Closest station has bikes: %s. Closest with docks: %s", availability.NearestStation.Name, availability.NearestStationWithDocks.Name)
		} else if availability.NearestStationWithDocks.Name == availability.NearestStation.Name {
			message = fmt.Sprintf("Closest station has docks: %s. Closest with bikes: %s", availability.NearestStation.Name, availability.NearestStationWithBikes.Name)
		} else {
			message = fmt.Sprintf(`
			Closest station: %s,
			Closest with bikes: %s,
			Closest with docks: %s
		`,
				availability.NearestStation.Name,
				availability.NearestStationWithBikes.Name,
				availability.NearestStationWithDocks.Name,
			)
		}

		w.Write([]byte(message))
	})

	fmt.Println("Starting server on port", cfg.Port)
	http.ListenAndServe(":"+cfg.Port, nil)
}
