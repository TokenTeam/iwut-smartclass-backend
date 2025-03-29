package util

import (
	"fmt"
	"iwut-smart-timetable-backend/internal/middleware"
	"os/exec"
)

// ConvertVideoToAudio 将视频文件转换为音频文件
func ConvertVideoToAudio(inputFile, outputFile string) error {
	middleware.Logger.Log("INFO", fmt.Sprintf("[FFMPEG] Starting conversion: %s to %s", inputFile, outputFile))

	cmd := exec.Command("ffmpeg", "-i", inputFile, "-vn", "-acodec", "copy", outputFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("[FFMPEG] Conversion failed: %v, output: %s", err, string(output)))
		return err
	}

	middleware.Logger.Log("INFO", fmt.Sprintf("[FFMPEG] Conversion successful: %s to %s", inputFile, outputFile))
	return nil
}
