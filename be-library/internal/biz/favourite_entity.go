package biz

import "time"

// 收藏座位核心结构体 - 只保留必要字段
type FavoriteSeat struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	StudentID  string    `gorm:"index:idx_student_dev,unique;not null;size:20" json:"student_id"` // 学号
	SeatID     string    `gorm:"index:idx_student_dev,unique;not null;size:50" json:"seat_id"`    // 设备ID，唯一标识座位
	LayerName  string    `gorm:"size:100;not null" json:"layer_name"`
	RoomName   string    `gorm:"size:50;not null" json:"room_name"`
	SeatName   string    `gorm:"size:100;not null" json:"dev_name"`
	CreateTime time.Time `gorm:"index:idx_create_time;not null" json:"create_time"`

	// 可选：座位状态快照（便于展示时不用再查询原始座位信息）
	IsAvailable bool `json:"is_available"` // 当前是否可用（缓存字段）
}

// 用于API响应的DTO
type FavoriteSeatResponse struct {
	Seats []FavoriteSeat `json:"favourite_seats"`
}

type FavoriteRepo struct {
}
