package course

import (
	"context"
	"fmt"
	"iwut-smartclass-backend/internal/middleware"
	"time"

	"gorm.io/gorm"
)

type CourseDBService struct {
	Database *gorm.DB
}

type Course struct {
	CourseId int               `json:"course_id"`
	SubId    int               `json:"sub_id"`
	Name     string            `json:"name"`
	Teacher  string            `json:"teacher"`
	Location string            `json:"location"`
	Date     string            `json:"date"`
	Time     string            `json:"time"`
	Video    string            `json:"video"`
	Asr      string            `json:"asr"`
	Summary  map[string]string `json:"summary"`
}

type UserSummary struct {
	Summary string
	Model   string
	Token   uint32
}

func NewCourseDbService(db *gorm.DB) *CourseDBService {
	return &CourseDBService{Database: db}
}

// GetCourseDataFromDb 从数据库中获取课程数据
func (s *CourseDBService) GetCourseDataFromDb(subId int) (Course, error) {
	var course Course

	// 添加超时控制
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 使用 GORM Raw SQL 查询
	var result struct {
		SubId         int
		CourseId      int
		Name          string
		Teacher       string
		Location      string
		Date          string
		Time          string
		Video         *string
		Asr           *string
		SummaryStatus *string
		SummaryData   *string
		Model         *string
		Token         *uint32
		SummaryUser   *string
	}

	dbResult := s.Database.WithContext(ctx).Raw(
		`SELECT sub_id, course_id, name, teacher, location, date, time, video, asr, summary_status, summary_data, model, token, summary_user FROM course WHERE sub_id = ?`,
		subId,
	).Scan(&result)

	if dbResult.Error != nil {
		return Course{}, dbResult.Error
	}

	// 检查是否找到记录（GORM Raw Scan 在找不到记录时不会返回错误，但 RowsAffected 会是 0）
	if dbResult.RowsAffected == 0 {
		middleware.Logger.Log("DEBUG", fmt.Sprintf("[DB] Could not find course data in database, subId: %d", subId))
		return Course{}, fmt.Errorf("sql: no rows in result set")
	}

	course.SubId = result.SubId
	course.CourseId = result.CourseId
	course.Name = result.Name
	course.Teacher = result.Teacher
	course.Location = result.Location
	course.Date = result.Date
	course.Time = result.Time

	if result.Video != nil {
		course.Video = *result.Video
	}
	if result.Asr != nil {
		course.Asr = *result.Asr
	}

	var status, data, model, token string
	if result.SummaryStatus != nil {
		status = *result.SummaryStatus
	}
	if result.SummaryData != nil {
		data = *result.SummaryData
	}
	if result.Model != nil {
		model = *result.Model
	}
	if result.Token != nil {
		token = fmt.Sprintf("%d", *result.Token)
	}

	course.Summary = map[string]string{
		"status": status,
		"data":   data,
		"model":  model,
		"token":  token,
	}
	middleware.Logger.Log("DEBUG", fmt.Sprintf("Course data found in database, subId: %d", subId))
	return course, nil
}

// SaveCourseDataToDb 将课程数据写入数据库
func (s *CourseDBService) SaveCourseDataToDb(course Course) error {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := s.Database.WithContext(ctx).Exec(
		`INSERT INTO course (sub_id, course_id, name, teacher, location, date, time, video, summary_status, summary_data, model, token, summary_user) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		course.SubId, course.CourseId, course.Name, course.Teacher, course.Location, course.Date, course.Time, course.Video, course.Summary["status"], course.Summary["data"], course.Summary["model"], course.Summary["token"], "",
	).Error

	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("[DB] %v", err))
		middleware.Logger.Log("ERROR", fmt.Sprintf("[DB] Failed to write course data to database, CourseId: %d, subId: %d: %v", course.CourseId, course.SubId, err))
		return err
	}
	middleware.Logger.Log("DEBUG", fmt.Sprintf("Write course data to database, CourseId: %d, subId: %d", course.CourseId, course.SubId))
	return nil
}

// SaveVideo 将 Video 内容写入数据库
func (s *CourseDBService) SaveVideo(subId int, video string) error {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := s.Database.WithContext(ctx).Exec(
		`UPDATE course SET video = ? WHERE sub_id = ?`,
		video, subId,
	).Error

	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to update database, subId: %d: %v", subId, err))
		return err
	}
	return nil
}

func (s *CourseDBService) GetUserSummaryFromDb(subId int, user string) ([]UserSummary, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var results []struct {
		Summary string
		Model   string
		Token   uint32
	}

	err := s.Database.WithContext(ctx).Raw(
		`SELECT summary, model, token FROM summary WHERE sub_id = ? AND user = ? ORDER BY create_at DESC`,
		subId, user,
	).Scan(&results).Error

	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to query user summary from database, subId: %d, user: %s: %v", subId, user, err))
		return nil, err
	}

	var summaries []UserSummary
	for _, result := range results {
		summaries = append(summaries, UserSummary{
			Summary: result.Summary,
			Model:   result.Model,
			Token:   result.Token,
		})
	}

	return summaries, nil
}
