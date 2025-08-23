package pkg

import "math"

const earthRadiusKm = 6371.0088

// Haversine calculates the great-circle distance between two points in kilometers
func Haversine(lat1, lon1, lat2, lon2 float64) float64 {
	toRad := func(d float64) float64 { return d * math.Pi / 180 }
	φ1, λ1 := toRad(lat1), toRad(lon1)
	φ2, λ2 := toRad(lat2), toRad(lon2)

	dφ := φ2 - φ1
	dλ := λ2 - λ1

	sinDφ := math.Sin(dφ / 2)
	sinDλ := math.Sin(dλ / 2)

	a := sinDφ*sinDφ + math.Cos(φ1)*math.Cos(φ2)*sinDλ*sinDλ
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKm * c
}
