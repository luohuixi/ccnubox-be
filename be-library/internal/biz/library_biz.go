package biz

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type libraryBiz struct {
	crawler          LibraryCrawler
	log              *log.Helper
	SeatRepo         SeatRepo
	RecordRepo       RecordRepo
	CreditPointsRepo CreditPointsRepo
}

func NewLibraryBiz(crawler LibraryCrawler, logger log.Logger, seatRepo SeatRepo, recordRepo RecordRepo, creditPointsRepo CreditPointsRepo) LibraryBiz {
	return &libraryBiz{
		crawler:          crawler,
		log:              log.NewHelper(logger),
		SeatRepo:         seatRepo,
		RecordRepo:       recordRepo,
		CreditPointsRepo: creditPointsRepo,
	}
}

func (b *libraryBiz) GetSeat(ctx context.Context, stuID string) (map[string][]*Seat, error) {
	data, err := b.SeatRepo.GetSeatInfos(ctx, stuID)
	if err != nil {
		b.log.Errorf("get seats from cache(stu_id:%v) failed: %v", stuID, err)
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
	// 去重并持久化
	if err = b.RecordRepo.UpsertFutureRecords(ctx, stuID, records); err != nil {
		b.log.Errorf("persist future records(stu_id:%v) failed: %v", stuID, err)
		return nil, err
	}
	// 从数据库读取去重后的数据
	result, err := b.RecordRepo.ListFutureRecords(ctx, stuID)
	if err != nil {
		b.log.Errorf("list future records(stu_id:%v) failed: %v", stuID, err)
		return nil, err
	}
	return result, nil
}

func (b *libraryBiz) GetHistory(ctx context.Context, stuID string) ([]*HistoryRecords, error) {
	history, err := b.crawler.GetHistory(ctx, stuID)
	if err != nil {
		b.log.Errorf("get history(stu_id:%v) failed: %v", stuID, err)
		return nil, err
	}
	// 去重并持久化
	if err = b.RecordRepo.UpsertHistoryRecords(ctx, stuID, history); err != nil {
		b.log.Errorf("persist history records(stu_id:%v) failed: %v", stuID, err)
		return nil, err
	}
	// 从数据库读取去重后的数据
	result, err := b.RecordRepo.ListHistoryRecords(ctx, stuID)
	if err != nil {
		b.log.Errorf("list history records(stu_id:%v) failed: %v", stuID, err)
		return nil, err
	}
	return result, nil
}

func (b *libraryBiz) GetCreditPoint(ctx context.Context, stuID string) (*CreditPoints, error) {
	creditPoints, err := b.crawler.GetCreditPoint(ctx, stuID)
	if err != nil {
		b.log.Errorf("get credit points(stu_id:%v) failed: %v", stuID, err)
		return nil, err
	}
	// 去重并持久化
	if err = b.CreditPointsRepo.UpsertCreditPoint(ctx, stuID, creditPoints); err != nil {
		b.log.Warnf("persist credit points(stu_id:%v) failed: %v", stuID, err)
		return nil, err
	}
	// 从数据库读取去重后的数据
	result, err := b.CreditPointsRepo.ListCreditPoint(ctx, stuID)
	if err != nil {
		b.log.Errorf("list credit points(stu_id:%v) failed: %v", stuID, err)
		return nil, err
	}
	return result, nil
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

// 2025-09-02 20:00
func (b *libraryBiz) ReserveSeatRandomly(ctx context.Context, stuID, start, end string) (string, error) {
	layout := "2006-01-02 15:04"
	tStart, _ := time.Parse(layout, start)
	tEnd, _ := time.Parse(layout, end)

	qStart := tStart.Hour()*100 + tStart.Minute()
	qEnd := tEnd.Hour()*100 + tEnd.Minute()

	// 查找空闲预约
	seatDevID, isExist, err := b.SeatRepo.FindFirstAvailableSeat(ctx, int64(qStart), int64(qEnd))
	if err != nil {
		return "", err
	}
	if !isExist {
		return "", errors.New("available seat unfound")
	}

	parts := strings.Split(seatDevID, ":")

	// 执行预约操作
	msg, err := b.ReserveSeat(ctx, stuID, parts[1], start, end)
	if err != nil {
		b.log.Errorf("Randomly reserve(stu_id:%v seatid:%v) failed: %v", stuID, parts[1], err)
		return "", err
	}
	return msg, nil
}
