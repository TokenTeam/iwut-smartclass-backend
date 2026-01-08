package persistence

import (
	"context"
	"fmt"
	"time"

	"iwut-smartclass-backend/internal/domain/course"
	"iwut-smartclass-backend/internal/infrastructure/logger"

	"gorm.io/gorm"
)

// CourseRepository 课程仓储实现
type CourseRepository struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewCourseRepository 创建课程仓储
func NewCourseRepository(db *gorm.DB, logger logger.Logger) *CourseRepository {
	return &CourseRepository{
		db:     db,
		logger: logger,
	}
}

// FindBySubID 根据SubID查找课程
func (r *CourseRepository) FindBySubID(ctx context.Context, subID int) (*course.Course, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var result struct {
		SubID         int
		CourseID      int
		Name          string
		Teacher       string
		Location      string
		Date          string
		Time          string
		Video         *string
		AudioID      *string
		Asr           *string
		SummaryStatus *string
		SummaryData   *string
		Model         *string
		Token         *uint32
		SummaryUser   *string
	}

	err := r.db.WithContext(ctx).Table("course").
		Where("sub_id = ?", subID).
		First(&result).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("course not found: sub_id=%d", subID)
		}
		r.logger.Error("failed to find course", logger.String("error", err.Error()))
		return nil, err
	}

	c := &course.Course{
		SubID:       result.SubID,
		CourseID:    result.CourseID,
		Name:        result.Name,
		Teacher:     result.Teacher,
		Location:    result.Location,
		Date:        result.Date,
		Time:        result.Time,
		SummaryUser: "",
	}

	if result.Video != nil {
		c.Video = *result.Video
	}
	if result.AudioID != nil {
		c.AudioID = *result.AudioID
	}
	if result.Asr != nil {
		c.Asr = *result.Asr
	}
	if result.SummaryStatus != nil {
		c.SummaryStatus = *result.SummaryStatus
	}
	if result.SummaryData != nil {
		c.SummaryData = *result.SummaryData
	}
	if result.Model != nil {
		c.Model = *result.Model
	}
	if result.Token != nil {
		c.Token = *result.Token
	}
	if result.SummaryUser != nil {
		c.SummaryUser = *result.SummaryUser
	}

	return c, nil
}

// Save 保存课程
func (r *CourseRepository) Save(ctx context.Context, c *course.Course) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := r.db.WithContext(ctx).Table("course").Create(map[string]interface{}{
		"sub_id":          c.SubID,
		"course_id":       c.CourseID,
		"name":            c.Name,
		"teacher":         c.Teacher,
		"location":        c.Location,
		"date":            c.Date,
		"time":            c.Time,
		"video":           c.Video,
		"summary_status":  c.SummaryStatus,
		"summary_data":    c.SummaryData,
		"model":           c.Model,
		"token":           c.Token,
		"summary_user":    c.SummaryUser,
	}).Error

	if err != nil {
		r.logger.Error("failed to save course", logger.String("error", err.Error()))
		return err
	}

	return nil
}

// UpdateVideo 更新视频链接
func (r *CourseRepository) UpdateVideo(ctx context.Context, subID int, video string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := r.db.WithContext(ctx).Table("course").
		Where("sub_id = ?", subID).
		Update("video", video).Error

	if err != nil {
		r.logger.Error("failed to update video", logger.String("error", err.Error()))
		return err
	}

	return nil
}

// UpdateAsr 更新ASR文本
func (r *CourseRepository) UpdateAsr(ctx context.Context, subID int, asr string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := r.db.WithContext(ctx).Table("course").
		Where("sub_id = ?", subID).
		Update("asr", asr).Error

	if err != nil {
		r.logger.Error("failed to update asr", logger.String("error", err.Error()))
		return err
	}

	return nil
}

// UpdateSummaryStatus 更新摘要状态
func (r *CourseRepository) UpdateSummaryStatus(ctx context.Context, subID int, status string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := r.db.WithContext(ctx).Table("course").
		Where("sub_id = ?", subID).
		Update("summary_status", status).Error

	if err != nil {
		r.logger.Error("failed to update summary status", logger.String("error", err.Error()))
		return err
	}

	return nil
}

// UpdateSummary 更新摘要数据
func (r *CourseRepository) UpdateSummary(ctx context.Context, subID int, summary, model string, token uint32, user string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := r.db.WithContext(ctx).Table("course").
		Where("sub_id = ?", subID).
		Updates(map[string]interface{}{
			"summary_data":   summary,
			"model":          model,
			"token":          token,
			"summary_status": "finished",
			"summary_user":   user,
		}).Error

	if err != nil {
		r.logger.Error("failed to update summary", logger.String("error", err.Error()))
		return err
	}

	return nil
}

// UpdateAudioID 更新音频ID
func (r *CourseRepository) UpdateAudioID(ctx context.Context, subID int, audioID string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := r.db.WithContext(ctx).Table("course").
		Where("sub_id = ?", subID).
		Update("audio_id", audioID).Error

	if err != nil {
		r.logger.Error("failed to update audio_id", logger.String("error", err.Error()))
		return err
	}

	return nil
}
