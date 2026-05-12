package timestamps

import "fmt"

// CalcEvery returns timestamps spaced by interval seconds, starting at interval,
// stopping before duration. Both values are in seconds.
func CalcEvery(duration, interval float64) ([]float64, error) {
	if interval <= 0 {
		return nil, fmt.Errorf("interval must be positive, got %v", interval)
	}
	var ts []float64
	for t := interval; t < duration; t += interval {
		ts = append(ts, t)
	}
	return ts, nil
}

// CalcCount returns count timestamps evenly distributed across duration,
// avoiding the very start and end of the video.
func CalcCount(duration float64, count int) ([]float64, error) {
	if count <= 0 {
		return nil, fmt.Errorf("count must be positive, got %v", count)
	}
	ts := make([]float64, count)
	for i := 0; i < count; i++ {
		ts[i] = duration * float64(i+1) / float64(count+1)
	}
	return ts, nil
}
