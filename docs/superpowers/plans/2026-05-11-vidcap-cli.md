# vidcap CLI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a `vidcap` CLI that extracts screenshots and GIF clips from MP4 videos using ffmpeg.

**Architecture:** Cobra-based CLI with two subcommands (`screenshots`, `gifs`). Pure timestamp calculation logic lives in `internal/timestamps`. FFmpeg shell-out wrappers live in `internal/ffmpeg`. Commands wire flags → timestamps → ffmpeg calls.

**Tech Stack:** Go, github.com/spf13/cobra, ffmpeg/ffprobe (system dependency)

---

## File Map

| File | Responsibility |
|------|----------------|
| `main.go` | Entry point — calls `cmd.Execute()` |
| `cmd/root.go` | Root cobra command, ffmpeg/ffprobe presence check |
| `cmd/screenshots.go` | `screenshots` subcommand — flags, validation, orchestration |
| `cmd/gifs.go` | `gifs` subcommand — flags, validation, orchestration |
| `internal/timestamps/calc.go` | Pure timestamp calculation (`CalcEvery`, `CalcCount`) |
| `internal/timestamps/calc_test.go` | Unit tests for timestamp logic |
| `internal/ffmpeg/probe.go` | `ffprobe` wrapper — get video duration |
| `internal/ffmpeg/extract.go` | `ffmpeg` wrappers — screenshot and GIF extraction |

---

## Task 1: Initialize Go module and install cobra

**Files:**
- Create: `go.mod`
- Create: `main.go`

- [ ] **Step 1: Initialize module**

```bash
cd E:/Projects/screenshot-generator
go mod init vidcap
```

Expected: `go.mod` created with `module vidcap`

- [ ] **Step 2: Add cobra dependency**

```bash
go get github.com/spf13/cobra@latest
```

Expected: `go.mod` and `go.sum` updated

- [ ] **Step 3: Create main.go**

```go
package main

import "vidcap/cmd"

func main() {
	cmd.Execute()
}
```

- [ ] **Step 4: Commit**

```bash
git init
git add go.mod go.sum main.go
git commit -m "feat: initialize Go module with cobra"
```

---

## Task 2: Timestamp calculation logic (with tests)

**Files:**
- Create: `internal/timestamps/calc.go`
- Create: `internal/timestamps/calc_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/timestamps/calc_test.go`:

```go
package timestamps_test

import (
	"testing"

	"vidcap/internal/timestamps"
)

func TestCalcEvery(t *testing.T) {
	ts, err := timestamps.CalcEvery(100.0, 30.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// expect 30, 60, 90
	want := []float64{30.0, 60.0, 90.0}
	if len(ts) != len(want) {
		t.Fatalf("got %d timestamps, want %d: %v", len(ts), len(want), ts)
	}
	for i, v := range ts {
		if v != want[i] {
			t.Errorf("ts[%d] = %.3f, want %.3f", i, v, want[i])
		}
	}
}

func TestCalcEvery_ExactBoundary(t *testing.T) {
	// interval divides duration exactly — last point should be excluded
	ts, err := timestamps.CalcEvery(90.0, 30.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 30, 60 — 90 equals duration, excluded
	if len(ts) != 2 {
		t.Fatalf("got %d timestamps, want 2: %v", len(ts), ts)
	}
}

func TestCalcEvery_InvalidInterval(t *testing.T) {
	_, err := timestamps.CalcEvery(100.0, 0)
	if err == nil {
		t.Fatal("expected error for zero interval")
	}
}

func TestCalcCount(t *testing.T) {
	ts, err := timestamps.CalcCount(100.0, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// duration=100, count=3: 100*1/4=25, 100*2/4=50, 100*3/4=75
	want := []float64{25.0, 50.0, 75.0}
	if len(ts) != len(want) {
		t.Fatalf("got %d timestamps, want %d: %v", len(ts), len(want), ts)
	}
	for i, v := range ts {
		if v != want[i] {
			t.Errorf("ts[%d] = %.3f, want %.3f", i, v, want[i])
		}
	}
}

func TestCalcCount_InvalidCount(t *testing.T) {
	_, err := timestamps.CalcCount(100.0, 0)
	if err == nil {
		t.Fatal("expected error for zero count")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd E:/Projects/screenshot-generator
go test ./internal/timestamps/...
```

Expected: compile error — package `timestamps` does not exist yet

- [ ] **Step 3: Implement calc.go**

Create `internal/timestamps/calc.go`:

```go
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
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/timestamps/... -v
```

Expected: all 5 tests PASS

- [ ] **Step 5: Commit**

```bash
git add internal/timestamps/
git commit -m "feat: add timestamp calculation logic with tests"
```

---

## Task 3: ffprobe duration probe

**Files:**
- Create: `internal/ffmpeg/probe.go`

- [ ] **Step 1: Create probe.go**

```go
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
	).Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed on %q: %w", inputPath, err)
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
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./internal/ffmpeg/...
```

Expected: no output (success)

- [ ] **Step 3: Commit**

```bash
git add internal/ffmpeg/probe.go
git commit -m "feat: add ffprobe duration wrapper"
```

---

## Task 4: ffmpeg extract wrappers (screenshot + GIF)

**Files:**
- Create: `internal/ffmpeg/extract.go`

- [ ] **Step 1: Create extract.go**

```go
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
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./internal/ffmpeg/...
```

Expected: no output (success)

- [ ] **Step 3: Commit**

```bash
git add internal/ffmpeg/extract.go
git commit -m "feat: add ffmpeg screenshot and gif extraction wrappers"
```

---

## Task 5: Root cobra command with ffmpeg check

**Files:**
- Create: `cmd/root.go`

- [ ] **Step 1: Create cmd/root.go**

```go
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "vidcap",
	Short: "Extract screenshots and GIFs from video files",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		for _, tool := range []string{"ffmpeg", "ffprobe"} {
			if _, err := exec.LookPath(tool); err != nil {
				return fmt.Errorf("%s not found in PATH — please install ffmpeg", tool)
			}
		}
		return nil
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./cmd/...
```

Expected: no output (success)

- [ ] **Step 3: Commit**

```bash
git add cmd/root.go
git commit -m "feat: add root cobra command with ffmpeg presence check"
```

---

## Task 6: screenshots subcommand

**Files:**
- Create: `cmd/screenshots.go`

- [ ] **Step 1: Create cmd/screenshots.go**

```go
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
```

- [ ] **Step 2: Build the full binary**

```bash
go build -o vidcap .
```

Expected: `vidcap` binary created (or `vidcap.exe` on Windows)

- [ ] **Step 3: Smoke test**

```bash
./vidcap screenshots --help
```

Expected: usage printed with `--every` and `--count` flags shown

- [ ] **Step 4: Commit**

```bash
git add cmd/screenshots.go
git commit -m "feat: add screenshots subcommand"
```

---

## Task 7: gifs subcommand

**Files:**
- Create: `cmd/gifs.go`

- [ ] **Step 1: Create cmd/gifs.go**

```go
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
```

- [ ] **Step 2: Build the full binary**

```bash
go build -o vidcap .
```

Expected: binary created successfully

- [ ] **Step 3: Smoke test**

```bash
./vidcap gifs --help
```

Expected: usage printed with `--duration`, `--every`, and `--count` flags shown

- [ ] **Step 4: Run all tests**

```bash
go test ./...
```

Expected: all tests PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/gifs.go
git commit -m "feat: add gifs subcommand"
```

---

## Task 8: End-to-end test with a real video

This task requires ffmpeg to be installed and a real video file.

- [ ] **Step 1: Download a short test video**

```bash
# Use any short mp4 you have, or download one:
curl -L "https://www.w3schools.com/html/mov_bbb.mp4" -o test.mp4
```

- [ ] **Step 2: Test screenshots --every**

```bash
./vidcap screenshots --every 1s test.mp4
```

Expected: `test-screenshots/` created, PNG files numbered `001.png`, `002.png`, etc.

- [ ] **Step 3: Test screenshots --count**

```bash
./vidcap screenshots --count 5 test.mp4
```

Expected: `test-screenshots/` with exactly 5 PNGs (overwrites previous)

- [ ] **Step 4: Test gifs --every**

```bash
./vidcap gifs --duration 2s --every 5s test.mp4
```

Expected: `test-gifs/` created with GIF files

- [ ] **Step 5: Test gifs --count**

```bash
./vidcap gifs --duration 2s --count 3 test.mp4
```

Expected: `test-gifs/` with exactly 3 GIFs

- [ ] **Step 6: Test error cases**

```bash
# Missing flag
./vidcap screenshots test.mp4
# Expected: error "specify exactly one of --every or --count"

# Both flags
./vidcap screenshots --every 5s --count 3 test.mp4
# Expected: error "specify exactly one of --every or --count"

# Missing --duration on gifs
./vidcap gifs --count 3 test.mp4
# Expected: error about required flag --duration
```

- [ ] **Step 7: Final commit**

```bash
git add .
git commit -m "feat: vidcap CLI complete — screenshots and gifs from video"
```

---

## Usage Summary

```bash
# Screenshots every 5 seconds
vidcap screenshots --every 5s input.mp4

# 10 screenshots evenly spaced
vidcap screenshots --count 10 input.mp4

# GIF clips: 3s each, every 30s
vidcap gifs --duration 3s --every 30s input.mp4

# GIF clips: 3s each, 5 total evenly spaced
vidcap gifs --duration 3s --count 5 input.mp4
```
