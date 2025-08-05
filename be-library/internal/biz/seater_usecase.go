package biz

import (
	"context"
	"time"

	"github.com/asynccnu/ccnubox-be/be-library/internal/conf"
	"github.com/go-kratos/kratos/v2/log"
)

// libraryUsecase struct + methods + NewLibraryUsecase

type libraryUsecase struct {
	crawler LibraryCrawler
	ccnu    CCNUServiceProxy
	Que     DelayQueue

	waitTime time.Duration
	log      *log.Helper
}

func NewLibraryUsecase(crawler LibraryCrawler, ccnu CCNUServiceProxy, logger log.Logger, cf *conf.Server,
	que DelayQueue) LibraryUsecase {
	waitTime := 1200 * time.Millisecond

	if cf.Grpc.Timeout.Seconds > 0 {
		waitTime = cf.Grpc.Timeout.AsDuration()
	}

	uc := &libraryUsecase{
		crawler:  crawler,
		ccnu:     ccnu,
		log:      log.NewHelper(logger),
		waitTime: waitTime,
		Que:      que,
	}

	go func() {
		if err := uc.Que.Consume("be-library-refresh-retry", uc.handleRetryMsg); err != nil {
			uc.log.Errorf("Error consuming retry message: %v", err)
		}
	}()

	return uc
}

// GetSeatFromCrawler 爬取座位信息
func (u *libraryUsecase) GetSeatFromCrawler(ctx context.Context, stuID string) (map[string][]*Seat, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, u.waitTime)
	defer cancel()

	getCookieStart := time.Now()

	cookie, err := u.ccnu.GetLibraryCookie(timeoutCtx, stuID)
	if err != nil {
		u.log.Errorf("Error getting cookie(stu_id:%v) from other service: %v", stuID, err)
		return nil, err
	}

	u.log.Infof("Get cookie (stu_id:%v) from other service,cost %v", stuID, time.Since(getCookieStart))

	seats, err := u.crawler.GetSeatInfos(ctx, cookie)
	if err != nil {
		u.log.Errorf("craw seats(stu_id:%v cookie:%v) failed: %v", stuID, cookie, err)
		return nil, err
	}

	return seats, nil
}

func (u *libraryUsecase) ReserveFromCrawler(ctx context.Context, stuID string, DevID, Start, End string) (string, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, u.waitTime)
	defer cancel()

	getCookieStart := time.Now()

	cookie, err := u.ccnu.GetLibraryCookie(timeoutCtx, stuID)
	if err != nil {
		u.log.Errorf("Error getting cookie(stu_id:%v) from other service: %v", stuID, err)
		return "", err
	}

	u.log.Infof("Get cookie (stu_id:%v) from other service,cost %v", stuID, time.Since(getCookieStart))

	result, err := u.crawler.ReserveSeat(ctx, cookie, DevID, Start, End)
	if err != nil {
		u.log.Errorf("reserve seats(stu_id:%v cookie:%v) failed: %v", stuID, cookie, err)
		return "", err
	}

	return result, nil
}

func (u *libraryUsecase) GetRecordFromCrawler(ctx context.Context, stuID string) ([]*FutureRecords, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, u.waitTime)
	defer cancel()

	getCookieStart := time.Now()

	cookie, err := u.ccnu.GetLibraryCookie(timeoutCtx, stuID)
	if err != nil {
		u.log.Errorf("Error getting cookie(stu_id:%v) from other service: %v", stuID, err)
		return nil, err
	}

	u.log.Infof("Get cookie (stu_id:%v) from other service,cost %v", stuID, time.Since(getCookieStart))

	records, err := u.crawler.GetRecord(ctx, cookie)
	if err != nil {
		u.log.Errorf("crawl seat records(stu_id:%v cookie:%v) failed: %v", stuID, cookie, err)
		return nil, err
	}

	return records, nil
}

func (u *libraryUsecase) GetHistoryFromCrawler(ctx context.Context, stuID string) ([]*HistoryRecords, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, u.waitTime)
	defer cancel()

	getCookieStart := time.Now()

	cookie, err := u.ccnu.GetLibraryCookie(timeoutCtx, stuID)
	if err != nil {
		u.log.Errorf("Error getting cookie(stu_id:%v) from other service: %v", stuID, err)
		return nil, err
	}

	u.log.Infof("Get cookie (stu_id:%v) from other service,cost %v", stuID, time.Since(getCookieStart))

	history, err := u.crawler.GetHistory(ctx, cookie)
	if err != nil {
		u.log.Errorf("crawl history records(stu_id:%v cookie:%v) failed: %v", stuID, cookie, err)
		return nil, err
	}

	return history, nil
}

func (u *libraryUsecase) GetCreditPointFromCrawler(ctx context.Context, stuID string) (*CreditPoints, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, u.waitTime)
	defer cancel()

	getCookieStart := time.Now()

	cookie, err := u.ccnu.GetLibraryCookie(timeoutCtx, stuID)
	if err != nil {
		u.log.Errorf("Error getting cookie(stu_id:%v) from other service: %v", stuID, err)
		return nil, err
	}

	u.log.Infof("Get cookie (stu_id:%v) from other service,cost %v", stuID, time.Since(getCookieStart))

	result, err := u.crawler.GetCreditPoint(ctx, cookie)
	if err != nil {
		u.log.Errorf("get credit point(stu_id:%v cookie:%v) failed: %v", stuID, cookie, err)
		return nil, err
	}

	return result, nil
}

func (u *libraryUsecase) GetDiscussionFromCrawler(ctx context.Context, stuID string, ClassID, Date string) ([]*Discussion, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, u.waitTime)
	defer cancel()

	getCookieStart := time.Now()

	cookie, err := u.ccnu.GetLibraryCookie(timeoutCtx, stuID)
	if err != nil {
		u.log.Errorf("Error getting cookie(stu_id:%v) from other service: %v", stuID, err)
		return nil, err
	}

	u.log.Infof("Get cookie (stu_id:%v) from other service,cost %v", stuID, time.Since(getCookieStart))

	result, err := u.crawler.GetDiscussion(ctx, cookie, ClassID, Date)
	if err != nil {
		u.log.Errorf("get discussion(stu_id:%v cookie:%v) failed: %v", stuID, cookie, err)
		return nil, err
	}

	return result, nil
}

func (u *libraryUsecase) SearchUserFromCrawler(ctx context.Context, stuID string, StudentID string) (*Search, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, u.waitTime)
	defer cancel()

	getCookieStart := time.Now()

	cookie, err := u.ccnu.GetLibraryCookie(timeoutCtx, stuID)
	if err != nil {
		u.log.Errorf("Error getting cookie(stu_id:%v) from other service: %v", stuID, err)
		return nil, err
	}

	u.log.Infof("Get cookie (stu_id:%v) from other service,cost %v", stuID, time.Since(getCookieStart))

	result, err := u.crawler.SearchUser(ctx, cookie, StudentID)
	if err != nil {
		u.log.Errorf("search user(stu_id:%v cookie:%v) failed: %v", stuID, cookie, err)
		return nil, err
	}

	return result, nil
}

func (u *libraryUsecase) ReserveDFromCrawler(ctx context.Context, stuID string, DevID, LabID, KindID, Title, Start, End string, List []string) (string, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, u.waitTime)
	defer cancel()

	getCookieStart := time.Now()

	cookie, err := u.ccnu.GetLibraryCookie(timeoutCtx, stuID)
	if err != nil {
		u.log.Errorf("Error getting cookie(stu_id:%v) from other service: %v", stuID, err)
		return "", err
	}

	u.log.Infof("Get cookie (stu_id:%v) from other service,cost %v", stuID, time.Since(getCookieStart))

	result, err := u.crawler.ReserveDiscussion(ctx, cookie, DevID, LabID, KindID, Title, Start, End, List)
	if err != nil {
		u.log.Errorf("reserve discussion(stu_id:%v cookie:%v) failed: %v", stuID, cookie, err)
		return "", err
	}

	return result, nil
}

func (u *libraryUsecase) CancelFromCrawler(ctx context.Context, stuID string, ID string) (string, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, u.waitTime)
	defer cancel()

	getCookieStart := time.Now()

	cookie, err := u.ccnu.GetLibraryCookie(timeoutCtx, stuID)
	if err != nil {
		u.log.Errorf("Error getting cookie(stu_id:%v) from other service: %v", stuID, err)
		return "", err
	}

	u.log.Infof("Get cookie (stu_id:%v) from other service,cost %v", stuID, time.Since(getCookieStart))

	result, err := u.crawler.CancelReserve(ctx, cookie, ID)
	if err != nil {
		u.log.Errorf("cancel discussion(stu_id:%v cookie:%v) failed: %v", stuID, cookie, err)
		return "", err
	}

	return result, nil
}

// 处理重试消息
func (u *libraryUsecase) handleRetryMsg(key, val []byte) {

}
