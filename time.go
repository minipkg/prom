package mp_prometheus

import "time"

func timeFromStart(start time.Time) float64 {
	return float64(time.Since(start).Milliseconds())
}
