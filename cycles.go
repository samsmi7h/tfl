package tfl

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"tfl/pkg"
	"time"
)

type BikePointsResponse []BikePoint

type BikePoint struct {
	ID                   string               `json:"id"`
	CommonName           string               `json:"commonName"`
	AdditionalProperties []AdditionalProperty `json:"additionalProperties"`
	Lat                  float64              `json:"lat"`
	Long                 float64              `json:"lon"`
}

type AdditionalProperty struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type CycleStation struct {
	ID              string
	Name            string
	Lat             float64
	Long            float64
	Installed       bool
	Locked          bool
	NbBikes         int
	NbStandardBikes int
	NbEBikes        int
	NbEmptyDocks    int
	NbDocks         int
}

func DecodeCycleStations(r io.Reader) ([]CycleStation, error) {
	dec := json.NewDecoder(r)
	var data BikePointsResponse
	if err := dec.Decode(&data); err != nil {
		return []CycleStation{}, err
	}

	var stations []CycleStation
	for _, bp := range data {
		var cs CycleStation

		cs.ID = bp.ID
		cs.Name = bp.CommonName
		cs.Lat = bp.Lat
		cs.Long = bp.Long

		for _, prop := range bp.AdditionalProperties {
			switch prop.Key {
			case "Installed":
				if prop.Value == "true" {
					cs.Installed = true
				} else {
					cs.Installed = false
				}
			case "Locked":
				if prop.Value == "true" {
					cs.Locked = true
				} else {
					cs.Locked = false
				}

			case "NbBikes":
				fmt.Sscanf(prop.Value, "%d", &cs.NbBikes)
			case "NbStandardBikes":
				fmt.Sscanf(prop.Value, "%d", &cs.NbStandardBikes)
			case "NbEBikes":
				fmt.Sscanf(prop.Value, "%d", &cs.NbEBikes)
			case "NbEmptyDocks":
				fmt.Sscanf(prop.Value, "%d", &cs.NbEmptyDocks)
			case "NbDocks":
				fmt.Sscanf(prop.Value, "%d", &cs.NbDocks)
			}

		}
		stations = append(stations, cs)
	}

	return stations, nil
}

func cycleURL(cfg Config) string {
	return fmt.Sprintf("https://api.tfl.gov.uk/BikePoint?app_key=%s", cfg.TFLAppKey)
}

func GetCycleStations(cfg Config) ([]CycleStation, error) {
	url := cycleURL(cfg)
	tr := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
		// Disable HTTP/2 (optional, but helps with some Cloudflare rulesets)
		TLSNextProto: map[string]func(authority string, c *tls.Conn) http.RoundTripper{},
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   15 * time.Second,
	}

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	// Make the request look like curl/a normal browser
	req.Header.Set("User-Agent", "curl/8.5.0")
	req.Header.Set("Accept", "application/json, text/xml;q=0.9, */*;q=0.8")
	req.Header.Set("Connection", "close") // optional: avoid long-lived connections

	res, err := client.Do(req)
	if err != nil {
		return []CycleStation{}, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return []CycleStation{}, fmt.Errorf("error fetching cycle availability: status %d", res.StatusCode)
	}

	return DecodeCycleStations(res.Body)
}

type CycleStationWithDistance struct {
	CycleStation
	DistanceKM float64
}

func GetCycleAvailabilityWithinRange(cfg Config, lat, long float64, radiusLimit float64) ([]CycleStationWithDistance, error) {
	data, err := GetCycleStations(cfg)
	if err != nil {
		return nil, err
	}

	var nearby []CycleStationWithDistance
	for _, station := range data {
		if !station.Installed || station.Locked {
			continue
		}

		distanceKM := pkg.Haversine(lat, long, station.Lat, station.Long)
		if distanceKM <= radiusLimit {
			nearby = append(nearby, CycleStationWithDistance{CycleStation: station, DistanceKM: distanceKM})
		}
	}

	return nearby, nil
}

type StationAvaialbility struct {
	StationsSortedByClosest []CycleStationWithDistance
	NearestStation          CycleStationWithDistance
	NearestStationWithBikes CycleStationWithDistance
	NearestStationWithDocks CycleStationWithDistance
}

func GetClosestAvailableBikeStations(cfg Config, lat, long float64, radiusLimit float64) (StationAvaialbility, error) {
	var av StationAvaialbility
	stations, err := GetCycleAvailabilityWithinRange(cfg, lat, long, radiusLimit)
	if err != nil {
		return StationAvaialbility{}, err
	}

	if len(stations) == 0 {
		return av, nil
	}

	sort.Slice(stations, func(i, j int) bool {
		return stations[i].DistanceKM < stations[j].DistanceKM
	})

	av.StationsSortedByClosest = stations
	av.NearestStation = stations[0]

	for _, s := range stations {
		if av.NearestStationWithDocks.Name == "" {
			if s.NbEmptyDocks > 0 {
				av.NearestStationWithDocks = s
			}
		}

		if av.NearestStationWithBikes.Name == "" {
			if s.NbStandardBikes > 0 {
				av.NearestStationWithBikes = s
			}
		}
	}

	return av, nil
}
