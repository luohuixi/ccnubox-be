package service

import (
	"context"

	pb "github.com/asynccnu/ccnubox-be/be-api/gen/proto/library/v1"
	"github.com/asynccnu/ccnubox-be/be-library/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
)

type LibraryService struct {
	pb.UnimplementedLibraryServer
	biz biz.LibraryBiz
	log *log.Helper
	conv *Assembler
}

func NewLibraryService(biz biz.LibraryBiz, logger log.Logger) *LibraryService {
	return &LibraryService{
		biz:  biz,
		log:  log.NewHelper(logger),
		conv: NewAssembler(),
	}
}

func (ls *LibraryService) GetSeat(ctx context.Context, req *pb.GetSeatRequest) (*pb.GetSeatResponse, error) {
	data, err := ls.biz.GetSeat(ctx, req.StuId)
	if err != nil {
		return nil, err
	}
	return ls.conv.ConvertGetSeatResponse(data), nil
}

func (ls *LibraryService) ReserveSeat(ctx context.Context, req *pb.ReserveSeatRequest) (*pb.ReserveSeatResponse, error) {
	msg, err := ls.biz.ReserveSeat(ctx, req.StuId, req.DevId, req.Start, req.End)
	if err != nil {
		return nil, err
	}
	return &pb.ReserveSeatResponse{Message: msg}, nil
}

func (ls *LibraryService) GetSeatRecord(ctx context.Context, req *pb.GetSeatRecordRequest) (*pb.GetSeatRecordResponse, error) {
	records, err := ls.biz.GetSeatRecord(ctx, req.StuId)
	if err != nil {
		return nil, err
	}
	return &pb.GetSeatRecordResponse{
		Record: ls.conv.ConvertRecords(records),
	}, nil
}

func (ls *LibraryService) GetHistory(ctx context.Context, req *pb.GetHistoryRequest) (*pb.GetHistoryResponse, error) {
	history, err := ls.biz.GetHistory(ctx, req.StuId)
	if err != nil {
		return nil, err
	}
	return &pb.GetHistoryResponse{
		History: ls.conv.ConvertHistory(history),
	}, nil
}

func (ls *LibraryService) GetCreditPoint(ctx context.Context, req *pb.GetCreditPointRequest) (*pb.GetCreditPointResponse, error) {
	cp, err := ls.biz.GetCreditPoint(ctx, req.StuId)
	if err != nil {
		return nil, err
	}
	return &pb.GetCreditPointResponse{
		CreditSummary: &pb.CreditSummary{
			System: cp.Summary.System,
			Remain: cp.Summary.Remain,
			Total:  cp.Summary.Total,
		},
		CreditRecord: ls.conv.ConvertCreditRecords(cp.Records),
	}, nil
}

func (ls *LibraryService) GetDiscussion(ctx context.Context, req *pb.GetDiscussionRequest) (*pb.GetDiscussionResponse, error) {
	ds, err := ls.biz.GetDiscussion(ctx, req.StuId, req.ClassId, req.Date)
	if err != nil {
		return nil, err
	}
	return &pb.GetDiscussionResponse{
		Discussions: ls.conv.ConvertDiscussions(ds),
	}, nil
}

func (ls *LibraryService) SearchUser(ctx context.Context, req *pb.SearchUserRequest) (*pb.SearchUserResponse, error) {
	u, err := ls.biz.SearchUser(ctx, req.StuId, req.StudentId)
	if err != nil {
		return nil, err
	}
	return &pb.SearchUserResponse{
		Id:    u.ID,
		Pid:   u.Pid,
		Name:  u.Name,
		Label: u.Label,
	}, nil
}

func (ls *LibraryService) ReserveDiscussion(ctx context.Context, req *pb.ReserveDiscussionRequest) (*pb.ReserveDiscussionResponse, error) {
	msg, err := ls.biz.ReserveDiscussion(ctx, req.StuId, req.DevId, req.LabId, req.KindId, req.Title, req.Start, req.End, req.List)
	if err != nil {
		return nil, err
	}
	return &pb.ReserveDiscussionResponse{Message: msg}, nil
}

func (ls *LibraryService) CancelReserve(ctx context.Context, req *pb.CancelReserveRequest) (*pb.CancelReserveResponse, error) {
	msg, err := ls.biz.CancelReserve(ctx, req.StuId, req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.CancelReserveResponse{Message: msg}, nil
}

func (ls *LibraryService) ReserveSeatRamdomly(ctx context.Context, req *pb.ReserveSeatRamdonlyRequest) (*pb.ReserveSeatRamdonlyResponse, error) {
	msg, err := ls.biz.ReserveSeatRamdomly(ctx, req.StuId, req.RoomId, req.Start, req.End)
	if err != nil {
		return nil, err
	}
	return &pb.ReserveSeatRamdonlyResponse{Message: msg}, nil
}
