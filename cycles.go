package tfl

import (
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"tfl/pkg"
	"time"
)

type EpochMillis struct {
	time.Time
}

func (e *EpochMillis) UnmarshalText(text []byte) error {
	s := string(text)
	if s == "" { // handles <removalDate/> etc.
		e.Time = time.Time{}
		return nil
	}
	ms, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	e.Time = time.Unix(0, ms*int64(time.Millisecond))
	return nil
}

type CycleStations struct {
	XMLName    xml.Name       `xml:"stations"`
	LastUpdate EpochMillis    `xml:"lastUpdate,attr"`
	Version    string         `xml:"version,attr"`
	Stations   []CycleStation `xml:"station"`
}

type CycleStation struct {
	ID              int         `xml:"id"`
	Name            string      `xml:"name"`
	TerminalName    string      `xml:"terminalName"`
	Lat             float64     `xml:"lat"`
	Long            float64     `xml:"long"`
	Installed       bool        `xml:"installed"`
	Locked          bool        `xml:"locked"`
	InstallDate     EpochMillis `xml:"installDate"`
	RemovalDate     EpochMillis `xml:"removalDate"`
	Temporary       bool        `xml:"temporary"`
	NbBikes         int         `xml:"nbBikes"`
	NbStandardBikes int         `xml:"nbStandardBikes"`
	NbEBikes        int         `xml:"nbEBikes"`
	NbEmptyDocks    int         `xml:"nbEmptyDocks"`
	NbDocks         int         `xml:"nbDocks"`
}

func DecodeCycleAvailability(r io.Reader) (CycleStations, error) {
	dec := xml.NewDecoder(r)
	var data CycleStations
	if err := dec.Decode(&data); err != nil {
		return CycleStations{}, err
	}

	return data, nil
}

func cycleURL(cfg Config) string {
	return fmt.Sprintf("https://tfl.gov.uk/tfl/syndication/feeds/cycle-hire/livecyclehireupdates.xml?app_key=%s", cfg.TFLAppKey)
}

func GetCycleAvailability(cfg Config) (CycleStations, error) {
	url := cycleURL(cfg)
	fmt.Println("Fetching cycle availability from", url)
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
	req.Header.Set("Accept", "application/xml, text/xml;q=0.9, */*;q=0.8")
	req.Header.Set("Connection", "close") // optional: avoid long-lived connections

	res, err := client.Do(req)
	if err != nil {
		return CycleStations{}, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		fmt.Println("Response body:", string(b))
		return CycleStations{}, fmt.Errorf("error fetching cycle availability: status %d", res.StatusCode)
	}

	return DecodeCycleAvailability(res.Body)
}

type CycleStationWithDistance struct {
	CycleStation
	DistanceKM float64
}

func GetCycleAvailabilityWithinRange(cfg Config, lat, long float64, radiusLimit float64) ([]CycleStationWithDistance, error) {
	data, err := GetCycleAvailability(cfg)
	if err != nil {
		return nil, err
	}

	var nearby []CycleStationWithDistance
	for _, station := range data.Stations {
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
