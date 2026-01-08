package dto

// GetCourseRequest 获取课程请求
type GetCourseRequest struct {
	CourseName string `json:"course_name" binding:"required"`
	Date       string `json:"date" binding:"required"`
	Token      string `json:"token" binding:"required"`
}

// GenerateSummaryRequest 生成摘要请求
type GenerateSummaryRequest struct {
	SubID int    `json:"sub_id" binding:"required"`
	Token string `json:"token" binding:"required"`
	Task  string `json:"task" binding:"required,oneof=new regenerate"`
}
