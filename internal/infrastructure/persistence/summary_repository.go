package persistence

import (
	"context"
	"fmt"
	"strings"
	"time"

	"iwut-smartclass-backend/internal/domain/summary"
	"iwut-smartclass-backend/internal/infrastructure/logger"

	"gorm.io/gorm"
)

const summaryTimeLayout = "2006-01-02 15:04:05"

// SummaryRepository 摘要仓储实现
type SummaryRepository struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewSummaryRepository 创建摘要仓储
func NewSummaryRepository(db *gorm.DB, logger logger.Logger) *SummaryRepository {
	return &SummaryRepository{
		db:     db,
		logger: logger,
	}
}

// FindBySubIDAndUser 根据SubID和用户查找摘要列表
func (r *SummaryRepository) FindBySubIDAndUser(ctx context.Context, subID int, user string) ([]*summary.Summary, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var results []struct {
		Summary     string
		Model       string
		Token       uint32
		CreateAtRaw string `gorm:"column:create_at"`
	}

	err := r.db.WithContext(ctx).Table("summary").
		Where("sub_id = ? AND user = ?", subID, user).
		Order("create_at DESC").
		Find(&results).Error

	if err != nil {
		r.logger.Error("failed to find summaries", logger.String("error", err.Error()))
		return nil, err
	}

	summaries := make([]*summary.Summary, 0, len(results))
	for _, result := range results {
		createAt, parseErr := time.ParseInLocation(summaryTimeLayout, strings.TrimSpace(result.CreateAtRaw), time.Local)
		if parseErr != nil {
			r.logger.Warn("failed to parse create_at, using zero time", logger.String("error", parseErr.Error()), logger.String("value", result.CreateAtRaw))
		}
		summaries = append(summaries, &summary.Summary{
			User:     user,
			SubID:    subID,
			CreateAt: createAt,
			Summary:  result.Summary,
			Model:    result.Model,
			Token:    result.Token,
		})
	}

	return summaries, nil
}

// Save 保存摘要
func (r *SummaryRepository) Save(ctx context.Context, s *summary.Summary) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	createAt := s.CreateAt
	if createAt.IsZero() {
		createAt = time.Now()
	}
	createAtStr := createAt.Format(summaryTimeLayout)

	err := r.db.WithContext(ctx).Table("summary").Create(map[string]interface{}{
		"user":      s.User,
		"sub_id":    s.SubID,
		"create_at": createAtStr,
		"summary":   s.Summary,
		"model":     s.Model,
		"token":     s.Token,
	}).Error

	if err != nil {
		r.logger.Error("failed to save summary", logger.String("error", err.Error()))
		return err
	}

	return nil
}

// Update 更新摘要
func (r *SummaryRepository) Update(ctx context.Context, s *summary.Summary) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	createAt := s.CreateAt
	if createAt.IsZero() {
		createAt = time.Now()
	}
	createAtStr := createAt.Format(summaryTimeLayout)

	err := r.db.WithContext(ctx).Table("summary").
		Where("sub_id = ? AND user = ? AND create_at = ?", s.SubID, s.User, createAtStr).
		Updates(map[string]interface{}{
			"summary": s.Summary,
			"model":   s.Model,
			"token":   s.Token,
		}).Error

	if err != nil {
		r.logger.Error("failed to update summary", logger.String("error", err.Error()))
		return err
	}

	return nil
}

// InitNewSummary 初始化新摘要
func (r *SummaryRepository) InitNewSummary(ctx context.Context, subID int, user string) (*summary.Summary, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	now := time.Now()
	createAtStr := now.Format(summaryTimeLayout)
	s := &summary.Summary{
		User:     user,
		SubID:    subID,
		CreateAt: now,
		Summary:  "",
		Model:    "",
		Token:    0,
	}

	err := r.db.WithContext(ctx).Table("summary").Create(map[string]interface{}{
		"user":      s.User,
		"sub_id":    s.SubID,
		"create_at": createAtStr,
		"summary":   s.Summary,
		"model":     s.Model,
		"token":     s.Token,
	}).Error

	if err != nil {
		r.logger.Error("failed to init new summary", logger.String("error", err.Error()))
		return nil, fmt.Errorf("failed to init new summary: %w", err)
	}

	return s, nil
}
