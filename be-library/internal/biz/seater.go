package biz

import (
	"context"
)

// domain interfaces + usecase interface

type LibraryCrawler interface {
	GetSeatInfos(ctx context.Context, cookie string) (map[string][]*Seat, error)
	ReserveSeat(ctx context.Context, cookie string, devid, start, end string) (string, error)
	GetRecord(ctx context.Context, cookie string) ([]*FutureRecords, error)
	GetHistory(ctx context.Context, cookie string) ([]*HistoryRecords, error)
	GetCreditPoint(ctx context.Context, cookie string) (*CreditPoints, error)
	GetDiscussion(ctx context.Context, cookie string, classid, date string) ([]*Discussion, error)
	SearchUser(ctx context.Context, cookie string, studentid string) (*Search, error)
	ReserveDiscussion(ctx context.Context, cookie string, devid, labid, kindid, title, start, end string, list []string) (string, error)
	CancelReserve(ctx context.Context, cookie string, id string) (string, error)
}

type LibraryUsecase interface {
	GetSeatFromCrawler(ctx context.Context, stuID string) (map[string][]*Seat, error)
	ReserveFromCrawler(ctx context.Context, stuID string, DevID, Start, End string) (string, error)
	GetRecordFromCrawler(ctx context.Context, stuID string) ([]*FutureRecords, error)
	GetHistoryFromCrawler(ctx context.Context, stuID string) ([]*HistoryRecords, error)
	GetCreditPointFromCrawler(ctx context.Context, stuID string) (*CreditPoints, error)
	GetDiscussionFromCrawler(ctx context.Context, stuID string, ClassID, Date string) ([]*Discussion, error)
	SearchUserFromCrawler(ctx context.Context, stuID string, StudentID string) (*Search, error)
	ReserveDFromCrawler(ctx context.Context, stuID string, DevID, LabID, KindID, Title, Start, End string, List []string) (string, error)
	CancelFromCrawler(ctx context.Context, stuID string, ID string) (string, error)
}
