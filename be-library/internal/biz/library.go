package biz

import (
	"context"
)

type LibraryBiz interface {
	GetSeat(ctx context.Context, stuID string) (map[string][]*Seat, error)
	ReserveSeat(ctx context.Context, stuID, devID, start, end string) (string, error)
	GetSeatRecord(ctx context.Context, stuID string) ([]*FutureRecords, error)
	GetHistory(ctx context.Context, stuID string) ([]*HistoryRecords, error)
	GetCreditPoint(ctx context.Context, stuID string) (*CreditPoints, error)
	GetDiscussion(ctx context.Context, stuID, classID, date string) ([]*Discussion, error)
	SearchUser(ctx context.Context, stuID, studentID string) (*Search, error)
	ReserveDiscussion(ctx context.Context, stuID, devID, labID, kindID, title, start, end string, list []string) (string, error)
	CancelReserve(ctx context.Context, stuID, id string) (string, error)
	ReserveSeatRandomly(ctx context.Context, stuID, roomID, start, end string) (string, error)
}

type LibraryCrawler interface {
	GetSeatInfos(ctx context.Context, stuID string) (map[string][]*Seat, error)
	ReserveSeat(ctx context.Context, stuID string, devid, start, end string) (string, error)
	GetRecord(ctx context.Context, stuID string) ([]*FutureRecords, error)
	GetHistory(ctx context.Context, stuID string) ([]*HistoryRecords, error)
	GetCreditPoint(ctx context.Context, stuID string) (*CreditPoints, error)
	GetDiscussion(ctx context.Context, stuID string, classid, date string) ([]*Discussion, error)
	SearchUser(ctx context.Context, stuID string, studentid string) (*Search, error)
	ReserveDiscussion(ctx context.Context, stuID string, devid, labid, kindid, title, start, end string, list []string) (string, error)
	CancelReserve(ctx context.Context, stuID string, id string) (string, error)
}
