package biz

import "context"

type FutureRecords struct {
	ID       string
	Owner    string
	Start    string
	End      string
	TimeDesc string
	States   string
	DevName  string
	RoomID   string
	RoomName string
	LabName  string
}

type HistoryRecords struct {
	Place      string
	Floor      string
	Status     string
	Date       string
	SubmitTime string
}

type RecordRepo interface {
	UpsertFutureRecords(ctx context.Context, stuID string, list []*FutureRecords) error
	ListFutureRecords(ctx context.Context, stuID string) ([]*FutureRecords, error)
	UpsertHistoryRecords(ctx context.Context, stuID string, list []*HistoryRecords) error
	ListHistoryRecords(ctx context.Context, stuID string) ([]*HistoryRecords, error)
}
