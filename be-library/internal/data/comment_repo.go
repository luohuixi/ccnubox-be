package data

import (
	"time"

	"github.com/asynccnu/ccnubox-be/be-library/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
)

type CommentRepo struct {
	data *Data
	log  *log.Helper
}

func NewCommentRepo(data *Data, logger log.Logger) biz.CommentRepo {
	return &CommentRepo{
		log:  log.NewHelper(logger),
		data: data,
	}
}

func (r CommentRepo) CreateComment(req *biz.CreateCommentReq) (string, error) {
	comment := biz.Comment{
		SeatID:    req.SeatID,
		Content:   req.Content,
		Rating:    req.Rating,
		Username:  req.Username,
		CreatedAt: time.Now(),
	}

	err := r.data.db.Create(&comment).Error
	if err != nil {
		return "", err
	}

	return "success", err
}

func (r CommentRepo) GetCommentsBySeatID(seatID int) ([]biz.Comment, error) {
	var comments []biz.Comment
	err := r.data.db.Where("seat_id = ?", seatID).Order("created_at desc").Find(&comments).Error
	return comments, err
}

func (r CommentRepo) DeleteComment(id int) (string, error) {
	err := r.data.db.Delete(&biz.Comment{}, id).Error
	if err != nil {
		return "", nil
	}

	return "success", err
}
