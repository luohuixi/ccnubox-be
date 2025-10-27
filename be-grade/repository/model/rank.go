package model

import "time"

type Rank struct {
	Id        int64  `json:"id;primary_key;auto_increment"`
	StudentId string `gorm:"column:student_id;type:varchar(100);not null;"`
	XnmBegin  int64  `gorm:"index"`
	XqmBegin  int64  `gorm:"index"`
	XnmEnd    int64  `gorm:"index"`
	XqmEnd    int64  `gorm:"index"`
	Rank      string
	Score     string
	Include   string
	Update    bool //该数据是否需要更新
	ViewAt    time.Time
}
