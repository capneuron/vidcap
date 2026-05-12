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
