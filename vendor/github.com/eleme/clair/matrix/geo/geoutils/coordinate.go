package geoutils

import (
	"math"
)

func transform(x, y float64) (lat, lng float64) {
	xy := x * y
	absX := math.Sqrt(math.Abs(x))
	d := (20.0*math.Sin(6.0*x*math.Pi) + 20.0*math.Sin(2.0*x*math.Pi)) * 2.0 / 3.0

	lat = -100.0 + 2.0*x + 3.0*y + 0.2*y*y + 0.1*xy + 0.2*absX
	lng = 300.0 + x + 2.0*y + 0.1*x*x + 0.1*xy + 0.1*absX

	lat += d
	lng += d

	lat += (20.0*math.Sin(y*math.Pi) + 40.0*math.Sin(y/3.0*math.Pi)) * 2.0 / 3.0
	lng += (20.0*math.Sin(x*math.Pi) + 40.0*math.Sin(x/3.0*math.Pi)) * 2.0 / 3.0

	lat += (160.0*math.Sin(y/12.0*math.Pi) + 320*math.Sin(y/30.0*math.Pi)) * 2.0 / 3.0
	lng += (150.0*math.Sin(x/12.0*math.Pi) + 300.0*math.Sin(x/30.0*math.Pi)) * 2.0 / 3.0

	return lat, lng
}

func delta(lat, lng float64) (dLat, dLng float64) {
	const a = 6378245.0
	const ee = 0.00669342162296594323
	dLat, dLng = transform(lng-105.0, lat-35.0)
	radLat := lat / 180.0 * math.Pi
	magic := math.Sin(radLat)
	magic = 1 - ee*magic*magic
	sqrtMagic := math.Sqrt(magic)
	dLat = (dLat * 180.0) / ((a * (1 - ee)) / (magic * sqrtMagic) * math.Pi)
	dLng = (dLng * 180.0) / (a / sqrtMagic * math.Cos(radLat) * math.Pi)
	return dLat, dLng
}

// WGStoGCJ converts coordinate from WGS-84 to GCJ-02
func WGStoGCJ(wgsLat, wgsLng float64) (gcjLat, gcjLng float64) {
	dLat, dLng := delta(wgsLat, wgsLng)
	gcjLat, gcjLng = wgsLat+dLat, wgsLng+dLng
	return gcjLat, gcjLng
}

// GCJtoWGS converts from GCJ-02 to WGS-84
func GCJtoWGS(gcjLat, gcjLng float64) (wgsLat, wgsLng float64) {
	dLat, dLng := delta(gcjLat, gcjLng)
	wgsLat, wgsLng = gcjLat-dLat, gcjLng-dLng
	return wgsLat, wgsLng
}
