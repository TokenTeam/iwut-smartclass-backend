package external

import (
	"context"
	"os/exec"

	"iwut-smartclass-backend/internal/domain/errors"
	"iwut-smartclass-backend/internal/infrastructure/logger"
)

// FFmpegService FFmpeg服务
type FFmpegService struct {
	logger logger.Logger
}

// NewFFmpegService 创建FFmpeg服务
func NewFFmpegService(logger logger.Logger) *FFmpegService {
	return &FFmpegService{
		logger: logger,
	}
}

// ConvertVideoToAudio 将视频转换为音频
func (s *FFmpegService) ConvertVideoToAudio(ctx context.Context, inputFile, outputFile string) error {
	s.logger.Info("starting video to audio conversion",
		logger.String("input", inputFile),
		logger.String("output", outputFile),
	)

	cmd := exec.CommandContext(ctx, "ffmpeg", "-i", inputFile, "-vn", "-acodec", "copy", outputFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		s.logger.Error("conversion failed",
			logger.String("error", err.Error()),
			logger.String("output", string(output)),
		)
		return errors.NewInternalError("failed to convert video to audio", err)
	}

	s.logger.Info("conversion finished",
		logger.String("input", inputFile),
		logger.String("output", outputFile),
	)
	return nil
}
