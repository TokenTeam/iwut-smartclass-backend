package _struct

import "time"

type Summary struct {
	User     string    `gorm:"column:user"`
	SubId    int       `gorm:"column:sub_id"`
	CreateAt time.Time `gorm:"column:create_at"`
	Summary  string    `gorm:"column:summary"`
	Model    string    `gorm:"column:model"`
	Token    uint32    `gorm:"column:token"`
}

func (Summary) TableName() string {
	return "summary"
}
