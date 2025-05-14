package course

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"iwut-smartclass-backend/internal/middleware"
)

type CourseDbService struct {
	Database *sql.DB
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

// NewCourseDbService 创建实例
func NewCourseDbService(db *sql.DB) *CourseDbService {
	return &CourseDbService{Database: db}
}

// GetCourseDataFromDb 从数据库中获取课程数据
func (s *CourseDbService) GetCourseDataFromDb(subId int) (Course, error) {
	var course Course
	query := `SELECT sub_id, course_id, name, teacher, location, date, time, video, asr, summary_status, summary_data, summary_user FROM course WHERE sub_id = ?`
	row := s.Database.QueryRow(query, subId)
	var video, summaryStatus, summaryData, summaryUser sql.NullString
	err := row.Scan(&course.SubId, &course.CourseId, &course.Name, &course.Teacher, &course.Location, &course.Date, &course.Time, &video, &course.Asr, &summaryStatus, &summaryData, &summaryUser)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			middleware.Logger.Log("DEBUG", fmt.Sprintf("[DB] Could not find course data in database, subId: %d: %v", subId, err))
			return Course{}, fmt.Errorf("sql: no rows in result set")
		}
		return Course{}, err
	}
	course.Video = video.String
	course.Summary = map[string]string{
		"status": summaryStatus.String,
		"data":   summaryData.String,
	}
	middleware.Logger.Log("DEBUG", fmt.Sprintf("Course data found in database, subId: %d", subId))
	return course, nil
}

// SaveCourseDataToDb 将课程数据写入数据库
func (s *CourseDbService) SaveCourseDataToDb(course Course) error {
	query := `INSERT INTO course (sub_id, course_id, name, teacher, location, date, time, video, summary_status, summary_data, summary_user) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.Database.Exec(query, course.SubId, course.CourseId, course.Name, course.Teacher, course.Location, course.Date, course.Time, course.Video, course.Summary["status"], course.Summary["data"], "")
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("[DB] %v", err))
		middleware.Logger.Log("ERROR", fmt.Sprintf("[DB] Failed to write course data to database, CourseId: %d, subId: %d: %v", course.CourseId, course.SubId, err))
		return err
	}
	middleware.Logger.Log("DEBUG", fmt.Sprintf("Write course data to database, CourseId: %d, subId: %d", course.CourseId, course.SubId))
	return nil
}

// SaveVideo 将 Video 内容写入数据库
func (s *CourseDbService) SaveVideo(subId int, video string) error {
	query := `UPDATE course SET video = ? WHERE sub_id = ?`
	_, err := s.Database.Exec(query, video, subId)
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to update database, subId: %d: %v", subId, err))
		return err
	}
	return nil
}
