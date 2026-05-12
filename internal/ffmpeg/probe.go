package ffmpeg

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
)

type probeFormat struct {
	Duration string `json:"duration"`
}

type probeResult struct {
	Format probeFormat `json:"format"`
}

// Duration returns the duration of the video file in seconds using ffprobe.
func Duration(inputPath string) (float64, error) {
	out, err := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		inputPath,
	).CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed on %q: %w\n%s", inputPath, err, out)
	}
	var result probeResult
	if err := json.Unmarshal(out, &result); err != nil {
		return 0, fmt.Errorf("parsing ffprobe output: %w", err)
	}
	d, err := strconv.ParseFloat(result.Format.Duration, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing duration %q: %w", result.Format.Duration, err)
	}
	return d, nil
}
