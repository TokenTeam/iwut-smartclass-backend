package summary

import "context"

// Repository 摘要仓储接口
type Repository interface {
	// FindBySubIDAndUser 根据SubID和用户查找摘要列表
	FindBySubIDAndUser(ctx context.Context, subID int, user string) ([]*Summary, error)
	// Save 保存摘要
	Save(ctx context.Context, summary *Summary) error
	// Update 更新摘要
	Update(ctx context.Context, summary *Summary) error
	// InitNewSummary 初始化新摘要
	InitNewSummary(ctx context.Context, subID int, user string) (*Summary, error)
}
