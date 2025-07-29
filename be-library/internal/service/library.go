package service

import (
	"context"

	pb "github.com/asynccnu/ccnubox-be/be-api/gen/proto/library/v1"
	"github.com/asynccnu/ccnubox-be/be-library/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
)

type LibraryService struct {
	pb.UnimplementedLibraryServer
	use biz.LibraryUsecase
	log *log.Helper
}

func NewLibraryService(use biz.LibraryUsecase, logger log.Logger) *LibraryService {
	return &LibraryService{
		use: use,
		log: log.NewHelper(logger),
	}
}

func (ls *LibraryService) GetSeat(ctx context.Context, req *pb.GetSeatRequest) (*pb.GetSeatResponse, error) {
	// 调用 usecase 层方法，爬取(或获取数据)
	data, err := ls.use.GetSeatFromCrawler(ctx, req.StuId)
	if err != nil {
		return nil, err
	}

	var seatResp pb.GetSeatResponse
	for roomID, seats := range data {
		room := &pb.RoomSeat{
			RoomId: roomID,
		}
		for _, seat := range seats {
			s := &pb.Seat{
				LabName:  seat.LabName,
				KindName: seat.KindName,
				DevId:    seat.DevID,
				DevName:  seat.DevName,
			}
			for _, ts := range seat.Ts {
				s.Ts = append(s.Ts, &pb.TimeSlot{
					Start:  ts.Start,
					End:    ts.End,
					State:  ts.State,
					Owner:  ts.Owner,
					Occupy: ts.Occupy,
				})
			}

			room.Seats = append(room.Seats, s)
		}

		seatResp.RoomSeats = append(seatResp.RoomSeats, room)
	}

	return &seatResp, nil
}

func (ls *LibraryService) ReserveSeat(ctx context.Context, req *pb.ReserveSeatRequest) (*pb.ReserveSeatResponse, error) {
	message, err := ls.use.ReserveFromCrawler(ctx, req.StuId, req.DevId, req.Start, req.End)
	if err != nil {
		return nil, err
	}

	return &pb.ReserveSeatResponse{
		Message: message,
	}, nil
}

func (ls *LibraryService) GetSeatRecord(ctx context.Context, req *pb.GetSeatRecordRequest) (*pb.GetSeatRecordResponse, error) {
	records, err := ls.use.GetRecordFromCrawler(ctx, req.StuId)
	if err != nil {
		return nil, err
	}

	var pbRecords []*pb.Record
	for _, r := range records {
		pbRecords = append(pbRecords, &pb.Record{
			Id:       r.ID,
			Owner:    r.Owner,
			Start:    r.Start,
			End:      r.End,
			TimeDesc: r.TimeDesc,
			Occur:    r.Occur,
			States:   r.States,
			DevName:  r.DevName,
			RoomId:   r.RoomID,
			RoomName: r.RoomName,
			LabName:  r.LabName,
		})
	}

	return &pb.GetSeatRecordResponse{
		Record: pbRecords,
	}, nil
}

func (ls *LibraryService) CancelSeat(ctx context.Context, req *pb.CancelSeatRequest) (*pb.CancelSeatResponse, error) {
	message, err := ls.use.CancelFromCrawler(ctx, req.StuId, req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.CancelSeatResponse{
		Message: message,
	}, nil
}

func (ls *LibraryService) GetCreditPoint(ctx context.Context, req *pb.GetCreditPointRequest) (*pb.GetCreditPointResponse, error) {
	creditPoints, err := ls.use.GetCreditPointFromCrawler(ctx, req.StuId)
	if err != nil {
		return nil, err
	}

	summary := &pb.CreditSummary{
		System: creditPoints.Summary.System,
		Remain: creditPoints.Summary.Remain,
		Total:  creditPoints.Summary.Total,
	}

	var records []*pb.CreditRecord
	for _, r := range creditPoints.Records {
		records = append(records, &pb.CreditRecord{
			Title:    r.Title,
			Subtitle: r.Subtitle,
			Location: r.Location,
		})
	}

	return &pb.GetCreditPointResponse{
		CreditSummary: summary,
		CreditRecord:  records,
	}, nil
}

func (ls *LibraryService) GetDiscussion(ctx context.Context, req *pb.GetDiscussionRequest) (*pb.GetDiscussionResponse, error) {
	discussions, err := ls.use.GetDiscussionFromCrawler(ctx, req.StuId, req.ClassId, req.Date)
	if err != nil {
		return nil, err
	}

	var pbDiscussions []*pb.Discussion
	for _, d := range discussions {
		var ts []*pb.DiscussionTS
		for _, t := range d.TS {
			ts = append(ts, &pb.DiscussionTS{
				Start:  t.Start,
				End:    t.End,
				State:  t.State,
				Title:  t.Title,
				Owner:  t.Owner,
				Occupy: t.Occupy,
			})
		}
		pbDiscussions = append(pbDiscussions, &pb.Discussion{
			LabName:  d.LabName,
			KindName: d.KindName,
			DevId:    d.DevID,
			DevName:  d.DevName,
			TS:       ts,
		})
	}

	return &pb.GetDiscussionResponse{
		Discussions: pbDiscussions,
	}, nil
}

func (ls *LibraryService) SearchUser(ctx context.Context, req *pb.SearchUserRequest) (*pb.SearchUserResponse, error) {
	user, err := ls.use.SearchUserFromCrawler(ctx, req.StuId, req.StudentId)
	if err != nil {
		return nil, err
	}

	return &pb.SearchUserResponse{
		Id:    user.ID,
		Pid:   user.Pid,
		Name:  user.Name,
		Label: user.Label,
	}, nil
}

func (ls *LibraryService) ReserveDiscussion(ctx context.Context, req *pb.ReserveDiscussionRequest) (*pb.ReserveDiscussionResponse, error) {
	message, err := ls.use.ReserveDFromCrawler(ctx, req.StuId, req.DevId, req.LabId, req.KindId, req.Title, req.Start, req.End, req.List)
	if err != nil {
		return nil, err
	}

	return &pb.ReserveDiscussionResponse{
		Message: message,
	}, nil
}

func (ls *LibraryService) CancelDiscussion(ctx context.Context, req *pb.CancelDiscussionRequest) (*pb.CancelDiscussionResponse, error) {
	message, err := ls.use.CancelDFromCrawler(ctx, req.StuId, req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.CancelDiscussionResponse{
		Message: message,
	}, nil
}
