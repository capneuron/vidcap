package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"vidcap/internal/ffmpeg"
	"vidcap/internal/timestamps"
)

var ssEvery time.Duration
var ssCount int

var screenshotsCmd = &cobra.Command{
	Use:   "screenshots [flags] <video>",
	Short: "Extract screenshots from a video",
	Args:  cobra.ExactArgs(1),
	RunE:  runScreenshots,
}

func init() {
	screenshotsCmd.Flags().DurationVar(&ssEvery, "every", 0, "interval between screenshots (e.g. 5s, 1m)")
	screenshotsCmd.Flags().IntVar(&ssCount, "count", 0, "total number of screenshots evenly spaced")
	rootCmd.AddCommand(screenshotsCmd)
}

func runScreenshots(cmd *cobra.Command, args []string) error {
	if (ssEvery == 0) == (ssCount == 0) {
		return fmt.Errorf("specify exactly one of --every or --count")
	}

	inputPath := args[0]
	duration, err := ffmpeg.Duration(inputPath)
	if err != nil {
		return err
	}

	var ts []float64
	if ssEvery != 0 {
		ts, err = timestamps.CalcEvery(duration, ssEvery.Seconds())
	} else {
		ts, err = timestamps.CalcCount(duration, ssCount)
	}
	if err != nil {
		return err
	}
	if len(ts) == 0 {
		return fmt.Errorf("no timestamps generated — video may be shorter than the interval")
	}

	base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	outDir := filepath.Join(filepath.Dir(inputPath), base+"-screenshots")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	for i, t := range ts {
		outPath := filepath.Join(outDir, fmt.Sprintf("%03d.png", i+1))
		fmt.Printf("[%d/%d] screenshot at %.1fs → %s\n", i+1, len(ts), t, outPath)
		if err := ffmpeg.Screenshot(inputPath, t, outPath); err != nil {
			return err
		}
	}
	fmt.Printf("Done. %d screenshots saved to %s\n", len(ts), outDir)
	return nil
}
