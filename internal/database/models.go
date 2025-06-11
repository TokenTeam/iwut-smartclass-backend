package database

type Course struct {
	SubID         int    `gorm:"primaryKey;column:sub_id"`
	CourseID      int    `gorm:"column:course_id"`
	Name          string `gorm:"column:name"`
	Teacher       string `gorm:"column:teacher"`
	Location      string `gorm:"column:location"`
	Date          string `gorm:"column:date"`
	Time          string `gorm:"column:time"`
	Video         string `gorm:"column:video"`
	AudioID       string `gorm:"column:audio_id"`
	Asr           string `gorm:"column:asr;type:longtext"`
	SummaryStatus string `gorm:"column:summary_status"`
	SummaryData   string `gorm:"column:summary_data;type:longtext"`
	SummaryUser   string `gorm:"column:summary_user"`
}

func (Course) TableName() string {
	return "course"
}
