package summary

import "time"

// Summary 摘要实体
type Summary struct {
	User    string
	SubID   int
	CreateAt time.Time
	Summary string
	Model   string
	Token   uint32
}

// IsEmpty 检查摘要是否为空
func (s *Summary) IsEmpty() bool {
	return s.Summary == ""
}
