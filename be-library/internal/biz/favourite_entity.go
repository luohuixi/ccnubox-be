package biz

import "time"

// 收藏座位核心结构体 - 只保留必要字段
type FavoriteSeat struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	StudentID  string    `gorm:"index:idx_student_dev,unique;not null;size:20" json:"student_id"` // 学号
	DevID      string    `gorm:"index:idx_student_dev,unique;not null;size:50" json:"dev_id"`     // 设备ID，唯一标识座位
	LabName    string    `gorm:"size:100;not null" json:"lab_name"`
	KindName   string    `gorm:"size:50;not null" json:"kind_name"`
	DevName    string    `gorm:"size:100;not null" json:"dev_name"`
	CreateTime time.Time `gorm:"index:idx_create_time;not null" json:"create_time"`
	UpdateTime time.Time `gorm:"autoUpdateTime" json:"update_time"`

	// 可选：座位状态快照（便于展示时不用再查询原始座位信息）
	LastState   string `gorm:"size:20" json:"last_state,omitempty"` // 最后已知状态
	IsAvailable bool   `json:"is_available"`                        // 当前是否可用（缓存字段）
}

// 用于API响应的DTO
type FavoriteSeatResponse struct {
	ID          uint64    `json:"id"`
	DevID       string    `json:"dev_id"`
	LabName     string    `json:"lab_name"`
	KindName    string    `json:"kind_name"`
	DevName     string    `json:"dev_name"`
	CreateTime  time.Time `json:"create_time"`
	IsAvailable bool      `json:"is_available"`
	LastState   string    `json:"last_state,omitempty"`
}
