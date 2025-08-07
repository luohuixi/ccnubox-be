package biz

import (
	"context"

	pb "github.com/asynccnu/ccnubox-be/be-api/gen/proto/library/v1"
)

type LibraryBiz interface {
	GetSeat(ctx context.Context, stuID string) (*pb.GetSeatResponse, error)
	ReserveSeat(ctx context.Context, stuID, devID, start, end string) (*pb.ReserveSeatResponse, error)
	GetSeatRecord(ctx context.Context, stuID string) (*pb.GetSeatRecordResponse, error)
	GetHistory(ctx context.Context, stuID string) (*pb.GetHistoryResponse, error)
	GetCreditPoint(ctx context.Context, stuID string) (*pb.GetCreditPointResponse, error)
	GetDiscussion(ctx context.Context, stuID, classID, date string) (*pb.GetDiscussionResponse, error)
	SearchUser(ctx context.Context, stuID, studentID string) (*pb.SearchUserResponse, error)
	ReserveDiscussion(ctx context.Context, stuID, devID, labID, kindID, title, start, end string, list []string) (*pb.ReserveDiscussionResponse, error)
	CancelReserve(ctx context.Context, stuID, id string) (*pb.CancelReserveResponse, error)
}

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
