package adjust

import (
	"math"

	mgeo "github.com/eleme/clair/matrix/geo"
)

func getDistance(loc, loc2 Location) float64 {
	radians := func(val float64) float64 {
		return math.Pi * val / 180.0
	}

	lonDiff := 0.5 * (loc.Lng - loc2.Lng)
	latDiff := 0.5 * (loc.Lat - loc2.Lat)

	lngLength := math.Sin(radians(lonDiff))
	latLength := math.Sin(radians(latDiff))
	x := math.Pow(latLength, 2) + math.Cos(radians(loc.Lat))*math.Cos(radians(loc2.Lat))*math.Pow(lngLength, 2)
	return 2.0 * EarthRadius * math.Asin(math.Sqrt(x))
}

func AdjustedRoute(rawRoute []Location) (route []Location, err error) {
	buildGPSInfo := func(routeList []Location) []map[string]float64 {
		var GPSInfoList []map[string]float64
		for i := 0; i < len(routeList)-1; i++ {
			GPSInfo := make(map[string]float64)
			GPSInfo["distance"] = getDistance(routeList[i], routeList[i+1])
			GPSInfo["time"] = float64(routeList[i+1].UTC - routeList[i].UTC)
			GPSInfoList = append(GPSInfoList, GPSInfo)
		}
		return GPSInfoList
	}

	getHashString := func(point Location) (string, error) {
		lat := point.Lat
		lng := point.Lng
		hashString, err := mgeo.HashEncodeWithPrecision(lat, lng, 8)
		return hashString, err
	}

	reconstructRoute := func(routeList []Location, suspiciousValues map[string]int) ([]Location, error) {
		maxValue := 1
		for _, value := range suspiciousValues {
			if value > maxValue {
				maxValue = value
			}
		}

		var newRouteList []Location
		for _, point := range routeList {
			hashString, err := getHashString(point)
			if err != nil {
				return newRouteList, err
			}
			// if key hashString does not exists, suspiciousValues[hashString] return 0
			if suspiciousValues[hashString] < maxValue {
				newRouteList = append(newRouteList, point)
			}
		}

		if n := len(routeList); n > 2 {
			var h1, h2 string
			h1, _ = getHashString(routeList[0])
			h2, _ = getHashString(routeList[1])
			if suspiciousValues[h1] == suspiciousValues[h2] && suspiciousValues[h1] == maxValue {
				tmpRouteList := append([]Location{}, routeList[1])
				newRouteList = append(tmpRouteList, newRouteList...)
			}

			if 1 != n-2 {
				h1, _ = getHashString(routeList[n-2])
				h2, _ = getHashString(routeList[n-1])
				if suspiciousValues[h1] == suspiciousValues[h2] && suspiciousValues[h1] == maxValue {
					newRouteList = append(newRouteList, routeList[n-2])
				}
			}
		}
		return newRouteList, nil
	}

	updateSuspiciousValues := func(point Location, suspiciousValues map[string]int) error {
		hashString, err := getHashString(point)
		if err != nil {
			return err
		}
		suspiciousValues[hashString]++
		return nil
	}

	upperSpeed := 20.0
	thresholdN := 10
	for i := 0; i < thresholdN; i++ {
		GPSInfoList := buildGPSInfo(rawRoute)
		suspiciousValues := make(map[string]int)
		for j, GPSInfo := range GPSInfoList {
			if 2.0*upperSpeed*GPSInfo["time"] < GPSInfo["distance"] {
				updateSuspiciousValues(rawRoute[j], suspiciousValues)
				updateSuspiciousValues(rawRoute[j+1], suspiciousValues)
			}
		}
		rawRoute, err = reconstructRoute(rawRoute, suspiciousValues)
		if err != nil {
			return []Location{}, err
		}
	}
	return rawRoute, nil
}
