package biz

import "context"

type CreditPoints struct {
	Summary *CreditSummary
	Records []*CreditRecord
}

type CreditSummary struct {
	System string
	Remain string
	Total  string
}

type CreditRecord struct {
	Title    string
	Subtitle string
	Location string
}

type CreditPointsRepo interface {
	UpsertCreditPoint(ctx context.Context, stuID string, list *CreditPoints) error
	ListCreditPoint(ctx context.Context, stuID string) (*CreditPoints, error)
}
