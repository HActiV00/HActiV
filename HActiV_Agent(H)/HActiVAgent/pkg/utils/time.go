// Copyright Authors of HActiV

// utils package for helping other package
package utils

import (
	"time"
)

func TimeExchange(timeInput string, hostRegion string) time.Time {
	times, _ := time.Parse(time.RFC3339, timeInput)
	hostLoc, _ := time.LoadLocation(hostRegion)
	return times.In(hostLoc)

}
