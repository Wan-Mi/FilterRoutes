package geoutils

import (
	"encoding/json"
	"math"
)

// Location Represents a Physical Location in geographic notation [lat, lng].
type Location struct {
	lat float64
	lng float64
}

const (
	// EarthRadius According to Wikipedia, the Earth's radius is about 6,371km
	EarthRadius = 6378137.0
)

// NewLocation Returns a new Location populated by the passed in latitude (lat) and longitude (lng) values.
func NewLocation(lat float64, lng float64) *Location {
	return &Location{lat: lat, lng: lng}
}

// GetLatitude Returns Location p's latitude.
func (loc *Location) GetLatitude() float64 {
	return loc.lat
}

// GetLongitude Returns Location p's longitude.
func (loc *Location) GetLongitude() float64 {
	return loc.lng
}

// EuclideanDistance Calculates the Haversine distance between two points in kilometers.
// Original Implementation from: http://www.movable-type.co.uk/scripts/latlong.html
func (loc *Location) EuclideanDistance(loc2 *Location) float64 {

	radians := func(val float64) float64 {
		return math.Pi * val / 180.0
	}

	lonDiff := 0.5 * (loc.lng - loc2.lng)
	latDiff := 0.5 * (loc.lat - loc2.lat)

	lngLength := math.Sin(radians(lonDiff))
	latLength := math.Sin(radians(latDiff))
	x := math.Pow(latLength, 2) + math.Cos(radians(loc.lat))*math.Cos(radians(loc2.lat))*math.Pow(lngLength, 2)
	return 2.0 * EarthRadius * math.Asin(math.Sqrt(x))
}

//GetShiftLocation returns the location with shift
func (loc *Location) GetShiftLocation(shiftAngle float64, shiftDistance float64) *Location {
	// implements math radians
	radians := func(val float64) float64 {
		return math.Pi * val / 180.0
	}

	// implements math degrees
	degrees := func(val float64) float64 {
		return val * 180.0 / math.Pi
	}

	latShiftDis := math.Sin(radians(shiftAngle)) * float64(shiftDistance)
	lngShiftDis := math.Cos(radians(shiftAngle)) * float64(shiftDistance)
	latDiff := degrees(latShiftDis / EarthRadius)
	lngDiff := degrees(lngShiftDis / (EarthRadius * math.Cos(radians(0.5*latDiff+loc.GetLatitude()))))
	shiftLatitude := loc.lat + latDiff
	shiftLongitude := loc.lng + lngDiff
	return &Location{lat: shiftLatitude, lng: shiftLongitude}
}

//MarshalJSON returns the loc marshall string
func (loc *Location) MarshalJSON() ([]byte, error) {
	locDict := make(map[string]float64)
	locDict["Latitude"] = loc.GetLatitude()
	locDict["Longitude"] = loc.GetLongitude()
	return json.Marshal(locDict)
}
