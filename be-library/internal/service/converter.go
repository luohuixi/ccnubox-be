package service

import (
	pb "github.com/asynccnu/ccnubox-be/be-api/gen/proto/library/v1"
	"github.com/asynccnu/ccnubox-be/be-library/internal/biz"
)

type Converter struct{}

func NewConverter() *Converter {
	return &Converter{}
}

func (c *Converter) ConvertMessages(data []biz.Comment) *pb.GetCommentResp {
	if len(data) == 0 {
		return &pb.GetCommentResp{}
	}

	result := make([]*pb.Comment, 0, len(data))
	for _, r := range data {
		result = append(result, &pb.Comment{
			Id:        int64(r.ID),
			SeatId:    r.SeatID,
			Username:  r.Username,
			Content:   r.Content,
			Rating:    int64(r.Rating),
			CreatedAt: r.CreatedAt.String(),
		})
	}

	return &pb.GetCommentResp{
		Comment: result,
	}
}
