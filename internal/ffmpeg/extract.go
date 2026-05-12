package ffmpeg

import (
	"fmt"
	"os/exec"
)

// Screenshot extracts a single frame at timestamp ts (seconds) and saves as a PNG to outputPath.
func Screenshot(inputPath string, ts float64, outputPath string) error {
	cmd := exec.Command("ffmpeg",
		"-ss", fmt.Sprintf("%.3f", ts),
		"-i", inputPath,
		"-frames:v", "1",
		"-q:v", "2",
		"-y",
		outputPath,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg screenshot at %.3fs: %w\n%s", ts, err, out)
	}
	return nil
}

// GIF extracts a GIF clip starting at ts (seconds) for gifDuration seconds, saved to outputPath.
// Output is scaled to 640px wide, 10fps, using lanczos filter.
func GIF(inputPath string, ts, gifDuration float64, outputPath string) error {
	cmd := exec.Command("ffmpeg",
		"-ss", fmt.Sprintf("%.3f", ts),
		"-i", inputPath,
		"-t", fmt.Sprintf("%.3f", gifDuration),
		"-vf", "fps=10,scale=640:-1:flags=lanczos",
		"-y",
		outputPath,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg gif at %.3fs: %w\n%s", ts, err, out)
	}
	return nil
}
