package biz

import "time"

type Comment struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"` // 评论ID
	SeatID    string    `gorm:"index;not null" json:"seat_id"`      // 关联座位
	Username  string    `gorm:"index;not null" json:"user_id"`      // 发表评论的用户
	Content   string    `gorm:"type:text;not null" json:"content"`  // 评论内容
	Rating    int       `gorm:"not null" json:"rating"`             // 评分（1-5）
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`   // 创建时间
}

type CommentRepo interface {
	CreateComment(req *CreateCommentReq) (string, error)
	GetCommentsBySeatID(seatID int) ([]Comment, error)
	DeleteComment(id int) (string, error)
}

type CreateCommentReq struct {
	SeatID   string
	Content  string
	Rating   int
	Username string
}
