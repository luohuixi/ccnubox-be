package biz

import (
	"context"
	"time"

	pb "github.com/asynccnu/ccnubox-be/be-api/gen/proto/library/v1"
	"github.com/go-kratos/kratos/v2/log"
)

type libraryBiz struct {
	crawler   LibraryCrawler
	ccnu      CCNUServiceProxy
	converter *Converter
	waitTime  time.Duration
	log       *log.Helper
}

func NewLibraryBiz(crawler LibraryCrawler, ccnu CCNUServiceProxy, logger log.Logger, waitTime time.Duration) LibraryBiz {
	return &libraryBiz{
		crawler:   crawler,
		ccnu:      ccnu,
		converter: NewConverter(),
		waitTime:  waitTime,
		log:       log.NewHelper(logger),
	}
}

func (b *libraryBiz) GetSeat(ctx context.Context, stuID string) (*pb.GetSeatResponse, error) {
	cookie, err := b.getCookie(ctx, stuID)
	if err != nil {
		b.log.Errorf("Error getting cookie(stu_id:%v) from other service: %v", stuID, err)
		return nil, err
	}
	data, err := b.crawler.GetSeatInfos(ctx, cookie)
	if err != nil {
		b.log.Errorf("craw seats(stu_id:%v cookie:%v) failed: %v", stuID, cookie, err)
		return nil, err
	}
	return b.converter.ConvertGetSeatResponse(data), nil
}

func (b *libraryBiz) ReserveSeat(ctx context.Context, stuID, devID, start, end string) (*pb.ReserveSeatResponse, error) {
	cookie, err := b.getCookie(ctx, stuID)
	if err != nil {
		b.log.Errorf("Error getting cookie(stu_id:%v) from other service: %v", stuID, err)
		return nil, err
	}
	message, err := b.crawler.ReserveSeat(ctx, cookie, devID, start, end)
	if err != nil {
		b.log.Errorf("reserve seats(stu_id:%v cookie:%v) failed: %v", stuID, cookie, err)
		return nil, err
	}
	return &pb.ReserveSeatResponse{Message: message}, nil
}

func (b *libraryBiz) GetSeatRecord(ctx context.Context, stuID string) (*pb.GetSeatRecordResponse, error) {
	cookie, err := b.getCookie(ctx, stuID)
	if err != nil {
		b.log.Errorf("Error getting cookie(stu_id:%v) from other service: %v", stuID, err)
		return nil, err
	}
	records, err := b.crawler.GetRecord(ctx, cookie)
	if err != nil {
		b.log.Errorf("get records(stu_id:%v cookie:%v) failed: %v", stuID, cookie, err)
		return nil, err
	}
	return &pb.GetSeatRecordResponse{
		Record: b.converter.ConvertRecords(records),
	}, nil
}

func (b *libraryBiz) GetHistory(ctx context.Context, stuID string) (*pb.GetHistoryResponse, error) {
	cookie, err := b.getCookie(ctx, stuID)
	if err != nil {
		b.log.Errorf("Error getting cookie(stu_id:%v) from other service: %v", stuID, err)
		return nil, err
	}
	history, err := b.crawler.GetHistory(ctx, cookie)
	if err != nil {
		b.log.Errorf("get history(stu_id:%v cookie:%v) failed: %v", stuID, cookie, err)
		return nil, err
	}
	return &pb.GetHistoryResponse{
		History: b.converter.ConvertHistory(history),
	}, nil
}

func (b *libraryBiz) GetCreditPoint(ctx context.Context, stuID string) (*pb.GetCreditPointResponse, error) {
	cookie, err := b.getCookie(ctx, stuID)
	if err != nil {
		b.log.Errorf("Error getting cookie(stu_id:%v) from other service: %v", stuID, err)
		return nil, err
	}
	creditPoints, err := b.crawler.GetCreditPoint(ctx, cookie)
	if err != nil {
		b.log.Errorf("get credit points(stu_id:%v cookie:%v) failed: %v", stuID, cookie, err)
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
	cookie, err := b.getCookie(ctx, stuID)
	if err != nil {
		b.log.Errorf("Error getting cookie(stu_id:%v) from other service: %v", stuID, err)
		return nil, err
	}
	discussions, err := b.crawler.GetDiscussion(ctx, cookie, classID, date)
	if err != nil {
		b.log.Errorf("get discussions(stu_id:%v cookie:%v) failed: %v", stuID, cookie, err)
		return nil, err
	}
	return &pb.GetDiscussionResponse{
		Discussions: b.converter.ConvertDiscussions(discussions),
	}, nil
}

func (b *libraryBiz) SearchUser(ctx context.Context, stuID, studentID string) (*pb.SearchUserResponse, error) {
	cookie, err := b.getCookie(ctx, stuID)
	if err != nil {
		b.log.Errorf("Error getting cookie(stu_id:%v) from other service: %v", stuID, err)
		return nil, err
	}
	user, err := b.crawler.SearchUser(ctx, cookie, studentID)
	if err != nil {
		b.log.Errorf("search user(stu_id:%v cookie:%v student_id:%v) failed: %v", stuID, cookie, studentID, err)
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
	cookie, err := b.getCookie(ctx, stuID)
	if err != nil {
		b.log.Errorf("Error getting cookie(stu_id:%v) from other service: %v", stuID, err)
		return nil, err
	}
	message, err := b.crawler.ReserveDiscussion(ctx, cookie, devID, labID, kindID, title, start, end, list)
	if err != nil {
		b.log.Errorf("reserve discussion(stu_id:%v cookie:%v) failed: %v", stuID, cookie, err)
		return nil, err
	}
	return &pb.ReserveDiscussionResponse{Message: message}, nil
}

func (b *libraryBiz) CancelReserve(ctx context.Context, stuID, id string) (*pb.CancelReserveResponse, error) {
	cookie, err := b.getCookie(ctx, stuID)
	if err != nil {
		b.log.Errorf("Error getting cookie(stu_id:%v) from other service: %v", stuID, err)
		return nil, err
	}
	message, err := b.crawler.CancelReserve(ctx, cookie, id)
	if err != nil {
		b.log.Errorf("cancel reserve(stu_id:%v cookie:%v id:%v) failed: %v", stuID, cookie, id, err)
		return nil, err
	}
	return &pb.CancelReserveResponse{Message: message}, nil
}

// 提取公共的获取cookie逻辑，并添加日志记录
func (b *libraryBiz) getCookie(ctx context.Context, stuID string) (string, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, b.waitTime)
	defer cancel()
	
	getCookieStart := time.Now()
	cookie, err := b.ccnu.GetLibraryCookie(timeoutCtx, stuID)
	if err != nil {
		return "", err
	}
	
	b.log.Infof("Get cookie (stu_id:%v) from other service, cost %v", stuID, time.Since(getCookieStart))
	return cookie, nil
}
