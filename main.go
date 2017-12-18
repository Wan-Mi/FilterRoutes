package main

import (
	"fmt"

	"github.com/Wan-Mi/FilterRoutes/adjust"
)

func main() {

	oriLocations := []adjust.Location{}
	loc := adjust.Location{
		22.1,
		112.2,
		1513590840,
	}

	oriLocations = append(oriLocations, loc)

	if resRouts, err := adjust.AdjustedRoute(oriLocations); err != nil {
		fmt.Println("err:", err)
	} else {
		fmt.Println(resRouts)
	}
}
