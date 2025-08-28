package service

import (
	"context"

	pb "github.com/asynccnu/ccnubox-be/be-api/gen/proto/library/v1"
	"github.com/asynccnu/ccnubox-be/be-library/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
)

type LibraryService struct {
	pb.UnimplementedLibraryServer
	biz       biz.LibraryBiz
	comment   biz.CommentRepo
	converter Converter
	log       *log.Helper
}

func NewLibraryService(biz biz.LibraryBiz, logger log.Logger, comment biz.CommentRepo) *LibraryService {
	return &LibraryService{
		biz:     biz,
		comment: comment,
		log:     log.NewHelper(logger),
	}
}

func (ls *LibraryService) GetSeat(ctx context.Context, req *pb.GetSeatRequest) (*pb.GetSeatResponse, error) {
	return ls.biz.GetSeat(ctx, req.StuId)
}

func (ls *LibraryService) ReserveSeat(ctx context.Context, req *pb.ReserveSeatRequest) (*pb.ReserveSeatResponse, error) {
	return ls.biz.ReserveSeat(ctx, req.StuId, req.DevId, req.Start, req.End)
}

func (ls *LibraryService) GetSeatRecord(ctx context.Context, req *pb.GetSeatRecordRequest) (*pb.GetSeatRecordResponse, error) {
	return ls.biz.GetSeatRecord(ctx, req.StuId)
}

func (ls *LibraryService) GetHistory(ctx context.Context, req *pb.GetHistoryRequest) (*pb.GetHistoryResponse, error) {
	return ls.biz.GetHistory(ctx, req.StuId)
}

func (ls *LibraryService) GetCreditPoint(ctx context.Context, req *pb.GetCreditPointRequest) (*pb.GetCreditPointResponse, error) {
	return ls.biz.GetCreditPoint(ctx, req.StuId)
}

func (ls *LibraryService) GetDiscussion(ctx context.Context, req *pb.GetDiscussionRequest) (*pb.GetDiscussionResponse, error) {
	return ls.biz.GetDiscussion(ctx, req.StuId, req.ClassId, req.Date)
}

func (ls *LibraryService) SearchUser(ctx context.Context, req *pb.SearchUserRequest) (*pb.SearchUserResponse, error) {
	return ls.biz.SearchUser(ctx, req.StuId, req.StudentId)
}

func (ls *LibraryService) ReserveDiscussion(ctx context.Context, req *pb.ReserveDiscussionRequest) (*pb.ReserveDiscussionResponse, error) {
	return ls.biz.ReserveDiscussion(ctx, req.StuId, req.DevId, req.LabId, req.KindId, req.Title, req.Start, req.End, req.List)
}

func (ls *LibraryService) CancelReserve(ctx context.Context, req *pb.CancelReserveRequest) (*pb.CancelReserveResponse, error) {
	return ls.biz.CancelReserve(ctx, req.StuId, req.Id)
}

func (ls *LibraryService) ReserveSeatRandomly(ctx context.Context, req *pb.ReserveSeatRamdonlyRequest) (*pb.ReserveSeatRamdonlyResponse, error) {
	return ls.biz.ReserveSeatRandomly(ctx, req.StuId, req.RoomId, req.Start, req.End)
}

func (ls *LibraryService) CreateComment(ctx context.Context, req *pb.CreateCommentReq) (*pb.Resp, error) {
	msg, err := ls.comment.CreateComment(&biz.CreateCommentReq{
		SeatID:   req.SeatId,
		Content:  req.Content,
		Rating:   int(req.Rating),
		Username: req.Username,
	})

	if err != nil {
		return nil, err
	}

	return &pb.Resp{
		Message: msg,
	}, nil
}

func (ls *LibraryService) GetComments(ctx context.Context, req *pb.ID) (*pb.GetCommentResp, error) {
	comments, err := ls.comment.GetCommentsBySeatID(int(req.Id))

	result := ls.converter.ConvertMessages(comments)

	return result, err
}

func (ls *LibraryService) DeleteComment(ctx context.Context, req *pb.ID) (*pb.Resp, error) {
	msg, err := ls.comment.DeleteComment(int(req.Id))

	return &pb.Resp{
		Message: msg,
	}, err
}
