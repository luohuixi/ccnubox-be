package biz

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
