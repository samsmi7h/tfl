package main

import (
	"fmt"
	"net/http"
	"strings"
	"tfl"
	"tfl/pkg"
)

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

		lat, lng, err := pkg.ParseCoords(coordStr)
		if err != nil {
			http.Error(w, "invalid coords parameter", http.StatusBadRequest)
			return
		}

		message, err := tfl.HandleGetNearbyCycleStations(cfg, lat, lng)
		if err != nil {
			tfl.HttpError(w, err)
			return
		}
		w.Write([]byte(message))
	})

	http.HandleFunc("/cycles/stations/siri", func(w http.ResponseWriter, r *http.Request) {
		idsStr := r.URL.Query().Get("ids")
		ids := strings.Split(idsStr, ",")
		if len(ids) == 0 {
			http.Error(w, "invalid ids parameter", http.StatusBadRequest)
			return
		}
		fmt.Println("Request: /cycles/stations?ids=", idsStr)
		message, err := tfl.HandleQueryStationAvailability(cfg, ids)
		if err != nil {
			tfl.HttpError(w, err)
			return
		}
		w.Write([]byte(message))
	})

	fmt.Println("Starting server on port", cfg.Port)
	http.ListenAndServe(":"+cfg.Port, nil)
}
