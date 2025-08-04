package biz

type FutureRecords struct {
	ID       string
	Owner    string
	Start    string
	End      string
	TimeDesc string
	Occur    string
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
