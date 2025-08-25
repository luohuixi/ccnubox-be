package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

type libraryBiz struct {
	crawler  LibraryCrawler
	SeatRepo SeatRepo
	log      *log.Helper
}

func NewLibraryBiz(crawler LibraryCrawler, logger log.Logger, seatRepo SeatRepo) LibraryBiz {
	return &libraryBiz{
		crawler:  crawler,
		log:      log.NewHelper(logger),
		SeatRepo: seatRepo,
	}
}

func (b *libraryBiz) GetSeat(ctx context.Context, stuID string) (map[string][]*Seat, error) {
	data, err := b.crawler.GetSeatInfos(ctx, stuID)
	if err != nil {
		b.log.Errorf("craw seats(stu_id:%v) failed: %v", stuID, err)
		return nil, err
	}
	return data, nil
}

func (b *libraryBiz) ReserveSeat(ctx context.Context, stuID, devID, start, end string) (string, error) {
	message, err := b.crawler.ReserveSeat(ctx, stuID, devID, start, end)
	if err != nil {
		b.log.Errorf("reserve seats(stu_id:%v) failed: %v", stuID, err)
		return "", err
	}
	return message, nil
}

func (b *libraryBiz) GetSeatRecord(ctx context.Context, stuID string) ([]*FutureRecords, error) {
	records, err := b.crawler.GetRecord(ctx, stuID)
	if err != nil {
		b.log.Errorf("get records(stu_id:%v) failed: %v", stuID, err)
		return nil, err
	}
	return records, nil
}

func (b *libraryBiz) GetHistory(ctx context.Context, stuID string) ([]*HistoryRecords, error) {
	history, err := b.crawler.GetHistory(ctx, stuID)
	if err != nil {
		b.log.Errorf("get history(stu_id:%v) failed: %v", stuID, err)
		return nil, err
	}
	return history, nil
}

func (b *libraryBiz) GetCreditPoint(ctx context.Context, stuID string) (*CreditPoints, error) {
	creditPoints, err := b.crawler.GetCreditPoint(ctx, stuID)
	if err != nil {
		b.log.Errorf("get credit points(stu_id:%v) failed: %v", stuID, err)
		return nil, err
	}
	return creditPoints, nil
}

func (b *libraryBiz) GetDiscussion(ctx context.Context, stuID, classID, date string) ([]*Discussion, error) {
	discussions, err := b.crawler.GetDiscussion(ctx, stuID, classID, date)
	if err != nil {
		b.log.Errorf("get discussions(stu_id:%v) failed: %v", stuID, err)
		return nil, err
	}
	return discussions, nil
}

func (b *libraryBiz) SearchUser(ctx context.Context, stuID, studentID string) (*Search, error) {
	user, err := b.crawler.SearchUser(ctx, stuID, studentID)
	if err != nil {
		b.log.Errorf("search user(stu_id:%v for student_id:%v) failed: %v", stuID, studentID, err)
		return nil, err
	}
	return user, nil
}

func (b *libraryBiz) ReserveDiscussion(ctx context.Context, stuID, devID, labID, kindID, title, start, end string, list []string) (string, error) {
	message, err := b.crawler.ReserveDiscussion(ctx, stuID, devID, labID, kindID, title, start, end, list)
	if err != nil {
		b.log.Errorf("reserve discussion(stu_id:%v) failed: %v", stuID, err)
		return "", err
	}
	return message, nil
}

func (b *libraryBiz) CancelReserve(ctx context.Context, stuID, id string) (string, error) {
	message, err := b.crawler.CancelReserve(ctx, stuID, id)
	if err != nil {
		b.log.Errorf("cancel reserve(stu_id:%v id:%v) failed: %v", stuID, id, err)
		return "", err
	}
	return message, nil
}

func (b *libraryBiz) ReserveSeatRamdomly(ctx context.Context, stuID, roomID, start, end string) (string, error) {
	seatDevID, err := b.SeatRepo.FindFirstAvailableSeat(ctx, roomID, start, end)
	if err != nil {
		return "", err
	}
	msg, err := b.ReserveSeat(ctx, stuID, seatDevID, start, end)
	if err != nil {
		b.log.Errorf("Ramdonly reserve(stu_id:%v id:%v) failed: %v", stuID, seatDevID, err)
		return "", err
	}
	return msg, nil
}
