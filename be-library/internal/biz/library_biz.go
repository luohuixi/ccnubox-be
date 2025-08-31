package biz

import (
	"context"

	pb "github.com/asynccnu/ccnubox-be/be-api/gen/proto/library/v1"
	"github.com/go-kratos/kratos/v2/log"
)

type libraryBiz struct {
	crawler   LibraryCrawler
	SeatRepo  SeatRepo
	converter *Converter
	log       *log.Helper
}

func NewLibraryBiz(crawler LibraryCrawler, logger log.Logger, seatRepo SeatRepo) LibraryBiz {
	return &libraryBiz{
		crawler:   crawler,
		converter: NewConverter(),
		log:       log.NewHelper(logger),
		SeatRepo:  seatRepo,
	}
}

func (b *libraryBiz) GetSeat(ctx context.Context, stuID string) (*pb.GetSeatResponse, error) {
	data, err := b.crawler.GetSeatInfos(ctx, stuID)
	if err != nil {
		b.log.Errorf("craw seats(stu_id:%v) failed: %v", stuID, err)
		return nil, err
	}
	return b.converter.ConvertGetSeatResponse(data), nil
}

func (b *libraryBiz) ReserveSeat(ctx context.Context, stuID, devID, start, end string) (*pb.ReserveSeatResponse, error) {
	message, err := b.crawler.ReserveSeat(ctx, stuID, devID, start, end)
	if err != nil {
		b.log.Errorf("reserve seats(stu_id:%v) failed: %v", stuID, err)
		return nil, err
	}
	return &pb.ReserveSeatResponse{Message: message}, nil
}

func (b *libraryBiz) GetSeatRecord(ctx context.Context, stuID string) (*pb.GetSeatRecordResponse, error) {
	records, err := b.crawler.GetRecord(ctx, stuID)
	if err != nil {
		b.log.Errorf("get records(stu_id:%v) failed: %v", stuID, err)
		return nil, err
	}
	return &pb.GetSeatRecordResponse{
		Record: b.converter.ConvertRecords(records),
	}, nil
}

func (b *libraryBiz) GetHistory(ctx context.Context, stuID string) (*pb.GetHistoryResponse, error) {
	history, err := b.crawler.GetHistory(ctx, stuID)
	if err != nil {
		b.log.Errorf("get history(stu_id:%v) failed: %v", stuID, err)
		return nil, err
	}
	return &pb.GetHistoryResponse{
		History: b.converter.ConvertHistory(history),
	}, nil
}

func (b *libraryBiz) GetCreditPoint(ctx context.Context, stuID string) (*pb.GetCreditPointResponse, error) {
	creditPoints, err := b.crawler.GetCreditPoint(ctx, stuID)
	if err != nil {
		b.log.Errorf("get credit points(stu_id:%v) failed: %v", stuID, err)
		return nil, err
	}
	return &pb.GetCreditPointResponse{
		CreditSummary: &pb.CreditSummary{
			System: creditPoints.Summary.System,
			Remain: creditPoints.Summary.Remain,
			Total:  creditPoints.Summary.Total,
		},
		CreditRecord: b.converter.ConvertCreditRecords(creditPoints.Records),
	}, nil
}

func (b *libraryBiz) GetDiscussion(ctx context.Context, stuID, classID, date string) (*pb.GetDiscussionResponse, error) {
	discussions, err := b.crawler.GetDiscussion(ctx, stuID, classID, date)
	if err != nil {
		b.log.Errorf("get discussions(stu_id:%v) failed: %v", stuID, err)
		return nil, err
	}
	return &pb.GetDiscussionResponse{
		Discussions: b.converter.ConvertDiscussions(discussions),
	}, nil
}

func (b *libraryBiz) SearchUser(ctx context.Context, stuID, studentID string) (*pb.SearchUserResponse, error) {
	user, err := b.crawler.SearchUser(ctx, stuID, studentID)
	if err != nil {
		b.log.Errorf("search user(stu_id:%v for student_id:%v) failed: %v", stuID, studentID, err)
		return nil, err
	}
	return &pb.SearchUserResponse{
		Id:    user.ID,
		Pid:   user.Pid,
		Name:  user.Name,
		Label: user.Label,
	}, nil
}

func (b *libraryBiz) ReserveDiscussion(ctx context.Context, stuID, devID, labID, kindID, title, start, end string, list []string) (*pb.ReserveDiscussionResponse, error) {
	message, err := b.crawler.ReserveDiscussion(ctx, stuID, devID, labID, kindID, title, start, end, list)
	if err != nil {
		b.log.Errorf("reserve discussion(stu_id:%v) failed: %v", stuID, err)
		return nil, err
	}
	return &pb.ReserveDiscussionResponse{Message: message}, nil
}

func (b *libraryBiz) CancelReserve(ctx context.Context, stuID, id string) (*pb.CancelReserveResponse, error) {
	message, err := b.crawler.CancelReserve(ctx, stuID, id)
	if err != nil {
		b.log.Errorf("cancel reserve(stu_id:%v id:%v) failed: %v", stuID, id, err)
		return nil, err
	}
	return &pb.CancelReserveResponse{Message: message}, nil
}

func (b *libraryBiz) ReserveSeatRandomly(ctx context.Context, stuID, roomID, start, end string) (*pb.ReserveSeatRamdonlyResponse, error) {
	seatDevID, err := b.SeatRepo.FindFirstAvailableSeat(ctx, roomID, start, end)
	if err != nil {
		return nil, err
	}

	resp, err := b.ReserveSeat(ctx, stuID, seatDevID, start, end)
	if err != nil {
		b.log.Errorf("Ramdonly reserve(stu_id:%v id:%v) failed: %v", stuID, seatDevID, err)
		return nil, err
	}

	return &pb.ReserveSeatRamdonlyResponse{
		Message: resp.Message,
	}, err
}
