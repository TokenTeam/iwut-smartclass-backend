package course

import "context"

// Repository 课程仓储接口
type Repository interface {
	// FindBySubID 根据SubID查找课程
	FindBySubID(ctx context.Context, subID int) (*Course, error)
	// Save 保存课程
	Save(ctx context.Context, course *Course) error
	// UpdateVideo 更新视频链接
	UpdateVideo(ctx context.Context, subID int, video string) error
	// UpdateAsr 更新ASR文本
	UpdateAsr(ctx context.Context, subID int, asr string) error
	// UpdateSummaryStatus 更新摘要状态
	UpdateSummaryStatus(ctx context.Context, subID int, status string) error
	// UpdateSummary 更新摘要数据
	UpdateSummary(ctx context.Context, subID int, summary, model string, token uint32, user string) error
	// UpdateAudioID 更新音频ID
	UpdateAudioID(ctx context.Context, subID int, audioID string) error
}
