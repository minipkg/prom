package prometheus_utils

import "time"

func MillisecondsFromStart(start time.Time) float64 {
	return float64(time.Since(start).Milliseconds())
}

func SecondsFromStart(start time.Time) float64 {
	return float64(time.Since(start).Seconds())
}
