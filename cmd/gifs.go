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

var gifEvery time.Duration
var gifCount int
var gifDuration time.Duration

var gifsCmd = &cobra.Command{
	Use:   "gifs [flags] <video>",
	Short: "Extract GIF clips from a video",
	Args:  cobra.ExactArgs(1),
	RunE:  runGifs,
}

func init() {
	gifsCmd.Flags().DurationVar(&gifDuration, "duration", 0, "duration of each GIF clip (e.g. 3s)")
	gifsCmd.Flags().DurationVar(&gifEvery, "every", 0, "interval between GIF clips (e.g. 30s)")
	gifsCmd.Flags().IntVar(&gifCount, "count", 0, "total number of GIF clips evenly spaced")
	if err := gifsCmd.MarkFlagRequired("duration"); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(gifsCmd)
}

func runGifs(cmd *cobra.Command, args []string) error {
	if (gifEvery == 0) == (gifCount == 0) {
		return fmt.Errorf("specify exactly one of --every or --count")
	}

	inputPath := args[0]
	videoDuration, err := ffmpeg.Duration(inputPath)
	if err != nil {
		return err
	}

	var ts []float64
	if gifEvery != 0 {
		ts, err = timestamps.CalcEvery(videoDuration, gifEvery.Seconds())
	} else {
		ts, err = timestamps.CalcCount(videoDuration, gifCount)
	}
	if err != nil {
		return err
	}
	if len(ts) == 0 {
		return fmt.Errorf("no timestamps generated — video may be shorter than the interval")
	}

	base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	outDir := filepath.Join(filepath.Dir(inputPath), base+"-gifs")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	clipSecs := gifDuration.Seconds()
	for i, t := range ts {
		outPath := filepath.Join(outDir, fmt.Sprintf("%03d.gif", i+1))
		fmt.Printf("[%d/%d] gif at %.1fs (%.1fs long) → %s\n", i+1, len(ts), t, clipSecs, outPath)
		if err := ffmpeg.GIF(inputPath, t, clipSecs, outPath); err != nil {
			return err
		}
	}
	fmt.Printf("Done. %d GIFs saved to %s\n", len(ts), outDir)
	return nil
}
