package course

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"iwut-smart-timetable-backend/internal/middleware"
)

type Service struct {
	Database *sql.DB
}

type Course struct {
	CourseID int               `json:"course_id"`
	SubID    int               `json:"sub_id"`
	Name     string            `json:"name"`
	Teacher  string            `json:"teacher"`
	Location string            `json:"location"`
	Date     string            `json:"date"`
	Time     string            `json:"time"`
	Video    string            `json:"video"`
	Summary  map[string]string `json:"summary"`
}

// NewGetCourseDBService 创建实例
func NewGetCourseDBService(db *sql.DB) *Service {
	return &Service{Database: db}
}

// GetCourseDataFromDB 从数据库中获取课程数据
func (s *Service) GetCourseDataFromDB(SubID int) (Course, error) {
	var course Course
	query := `SELECT sub_id, course_id, name, teacher, location, date, time, video, summary_status, summary_data FROM course WHERE sub_id = ?`
	row := s.Database.QueryRow(query, SubID)
	var video sql.NullString
	var summaryStatus, summaryData sql.NullString
	err := row.Scan(&course.SubID, &course.CourseID, &course.Name, &course.Teacher, &course.Location, &course.Date, &course.Time, &video, &summaryStatus, &summaryData)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			middleware.Logger.Log("DEBUG", fmt.Sprintf("[DB] Could not find course data in database, SubId: %d: %v", SubID, err))
			return Course{}, fmt.Errorf("sql: no rows in result set")
		}
		return Course{}, err
	}
	course.Video = video.String
	course.Summary = map[string]string{
		"status": summaryStatus.String,
		"data":   summaryData.String,
	}
	middleware.Logger.Log("DEBUG", fmt.Sprintf("Course data found in database, SubId: %d", SubID))
	return course, nil
}

// SaveCourseDataToDB 将课程数据写入数据库
func (s *Service) SaveCourseDataToDB(course Course) error {
	query := `INSERT INTO course (sub_id, course_id, name, teacher, location, date, time, video, summary_status, summary_data) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.Database.Exec(query, course.SubID, course.CourseID, course.Name, course.Teacher, course.Location, course.Date, course.Time, course.Video, course.Summary["status"], course.Summary["data"])
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("[DB] %v", err))
		middleware.Logger.Log("ERROR", fmt.Sprintf("[DB] Failed to write course data to database, CourseId: %d, SubId: %d: %v", course.CourseID, course.SubID, err))
		return err
	}
	middleware.Logger.Log("DEBUG", fmt.Sprintf("Write course data to database, CourseId: %d, SubId: %d", course.CourseID, course.SubID))
	return nil
}
