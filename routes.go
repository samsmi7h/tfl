package tfl

import (
	"fmt"
	"strings"
)

func HandleQueryStationAvailability(cfg Config, stationIDs []string) (string, error) {
	stations, err := GetCycleStations(cfg)
	if err != nil {
		return "", NewInternalError(fmt.Sprintf("error getting cycle stations: %v", err))
	}

	desiredStationIDs := map[string]CycleStation{}
	for _, id := range stationIDs {
		desiredStationIDs[id] = CycleStation{}
	}

	for _, station := range stations {
		if s, ok := desiredStationIDs[station.ID]; ok && s.Name == "" {
			desiredStationIDs[station.ID] = station
		}
	}

	var missingIDs []string
	for id := range desiredStationIDs {
		if desiredStationIDs[id].Name == "" {
			missingIDs = append(missingIDs, id)
		}
		if len(missingIDs) > 0 {
			return "", NewNotFoundError(fmt.Sprintf("could not find stations with IDs: %v", missingIDs))
		}
	}

	var message string
	for _, stationID := range stationIDs {
		station := desiredStationIDs[stationID]
		hasDocks := station.NbEmptyDocks > 0
		hasBikes := station.NbStandardBikes > 0

		if hasDocks && hasBikes {
			message += fmt.Sprintf("%s has %d bikes and %d docks.\n", shortStationName(station.Name), station.NbStandardBikes, station.NbEmptyDocks)
		} else if hasDocks {
			message += fmt.Sprintf("%s has %d docks but no bikes.\n", shortStationName(station.Name), station.NbEmptyDocks)
		} else if hasBikes {
			message += fmt.Sprintf("%s has %d bikes but no docks.\n", shortStationName(station.Name), station.NbStandardBikes)
		} else {
			message += fmt.Sprintf("%s has nothing, lol what.\n", shortStationName(station.Name))
		}
	}

	return message, nil

}

func shortStationName(n string) string {
	return strings.Split(n, ",")[0]
}

func HandleGetNearbyCycleStations(cfg Config, lat, lng float64) (string, error) {
	availability, err := GetClosestAvailableBikeStations(cfg, lat, lng, 1.0)
	if err != nil {
		return "", NewInternalError(fmt.Sprintf("error getting closest available bike stations: %v", err))
	}

	for _, s := range availability.StationsSortedByClosest {
		fmt.Printf("Station: %s\n", s.Name)
	}

	if len(availability.StationsSortedByClosest) == 0 {
		return "", NewNotFoundError("no cycle stations found within range")
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
	return message, nil
}
