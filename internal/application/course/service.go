package course

import (
	"context"

	"iwut-smartclass-backend/internal/domain/course"
	"iwut-smartclass-backend/internal/domain/errors"
	"iwut-smartclass-backend/internal/infrastructure/logger"
)

// Service 课程应用服务
type Service struct {
	courseRepo course.Repository
	logger     logger.Logger
}

// NewService 创建课程应用服务
func NewService(courseRepo course.Repository, logger logger.Logger) *Service {
	return &Service{
		courseRepo: courseRepo,
		logger:     logger,
	}
}

// GetCourse 获取课程
func (s *Service) GetCourse(ctx context.Context, subID int) (*course.Course, error) {
	c, err := s.courseRepo.FindBySubID(ctx, subID)
	if err != nil {
		s.logger.Error("failed to get course", logger.String("error", err.Error()))
		return nil, errors.WrapError(err, "failed to get course")
	}
	return c, nil
}

// SaveCourse 保存课程
func (s *Service) SaveCourse(ctx context.Context, c *course.Course) error {
	if err := s.courseRepo.Save(ctx, c); err != nil {
		s.logger.Error("failed to save course", logger.String("error", err.Error()))
		return errors.WrapError(err, "failed to save course")
	}
	return nil
}

// UpdateVideo 更新视频链接
func (s *Service) UpdateVideo(ctx context.Context, subID int, video string) error {
	if err := s.courseRepo.UpdateVideo(ctx, subID, video); err != nil {
		s.logger.Error("failed to update video", logger.String("error", err.Error()))
		return errors.WrapError(err, "failed to update video")
	}
	return nil
}

// UpdateAsr 更新ASR文本
func (s *Service) UpdateAsr(ctx context.Context, subID int, asr string) error {
	if err := s.courseRepo.UpdateAsr(ctx, subID, asr); err != nil {
		s.logger.Error("failed to update asr", logger.String("error", err.Error()))
		return errors.WrapError(err, "failed to update asr")
	}
	return nil
}

// UpdateSummaryStatus 更新摘要状态
func (s *Service) UpdateSummaryStatus(ctx context.Context, subID int, status string) error {
	if err := s.courseRepo.UpdateSummaryStatus(ctx, subID, status); err != nil {
		s.logger.Error("failed to update summary status", logger.String("error", err.Error()))
		return errors.WrapError(err, "failed to update summary status")
	}
	return nil
}

// UpdateSummary 更新摘要数据
func (s *Service) UpdateSummary(ctx context.Context, subID int, summary, model string, token uint32, user string) error {
	if err := s.courseRepo.UpdateSummary(ctx, subID, summary, model, token, user); err != nil {
		s.logger.Error("failed to update summary", logger.String("error", err.Error()))
		return errors.WrapError(err, "failed to update summary")
	}
	return nil
}

// UpdateAudioID 更新音频ID
func (s *Service) UpdateAudioID(ctx context.Context, subID int, audioID string) error {
	if err := s.courseRepo.UpdateAudioID(ctx, subID, audioID); err != nil {
		s.logger.Error("failed to update audio_id", logger.String("error", err.Error()))
		return errors.WrapError(err, "failed to update audio_id")
	}
	return nil
}
