package util

import (
	"context"
	"fmt"
	"iwut-smartclass-backend/internal/middleware"
	"os/exec"
)

func ConvertVideoToAudio(ctx context.Context, inputFile, outputFile string) error {
	middleware.Logger.Log("INFO", fmt.Sprintf("[FFMPEG] Starting conversion: %s to %s", inputFile, outputFile))

	cmd := exec.CommandContext(ctx, "ffmpeg", "-i", inputFile, "-vn", "-acodec", "copy", outputFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("[FFMPEG] Conversion failed: %v, output: %s", err, string(output)))
		return err
	}

	middleware.Logger.Log("INFO", fmt.Sprintf("[FFMPEG] Conversion finished: %s to %s", inputFile, outputFile))
	return nil
}
