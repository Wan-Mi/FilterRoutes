package geo

import "github.com/eleme/clair/matrix/geo/geoutils"

// WGStoGCJ converts coordinate from WGS-84 to GCJ-02
func WGStoGCJ(wgsLat, wgsLng float64) (gcjLat, gcjLng float64) {
	return geoutils.WGStoGCJ(wgsLat, wgsLng)
}

// GCJtoWGS converts from GCJ-02 to WGS-84
func GCJtoWGS(gcjLat, gcjLng float64) (wgsLat, wgsLng float64) {
	return geoutils.GCJtoWGS(gcjLat, gcjLng)
}

//EuclideanDistance returns earth distance with two given location
func EuclideanDistance(fromLat, fromLng, toLat, toLng float64) float64 {
	fromLocation := geoutils.NewLocation(fromLat, fromLng)
	toLocation := geoutils.NewLocation(toLat, toLng)
	return fromLocation.EuclideanDistance(toLocation)
}

//CircleArea get circle sqaure with given radius range
func CircleArea(radius float64) float64 {
	return geoutils.CircleArea(radius)
}

//PolygonArea get area of a polygon
func PolygonArea(locationList [][2]float64) (float64, error) {
	var locList []*geoutils.Location
	for _, location := range locationList {
		locList = append(locList, geoutils.NewLocation(location[0], location[1]))
	}
	return geoutils.PolygonArea(locList)
}

//HashEncode encode lat, lng location to string
func HashEncode(lat, lng float64) (string, error) {
	return geoutils.Encode(lat, lng, 7)
}

//HashDecode decode geohash string
//returns lat, lng, error
func HashDecode(geohashStr string) (float64, float64, error) {
	return geoutils.Decode(geohashStr)
}

//GeoHashNearby finds nearby geohash string list.
//geohash length cannot larger than 7
func GeoHashNearby(geohash string, nearbyRange float64) ([]string, error) {
	return geoutils.Nearby(geohash, int32(nearbyRange))
}

//HashEncodeWithPrecision encode lat, lng to string with given precision length
func HashEncodeWithPrecision(lat, lng float64, precison int) (string, error) {
	return geoutils.Encode(lat, lng, precison)
}
