package biz

type Discussion struct {
	LabName  string
	KindName string
	DevID    string
	DevName  string
	TS       []*DiscussionTS
}

type DiscussionTS struct {
	Start  string
	End    string
	State  string
	Title  string
	Owner  string
	Occupy bool
}

type Search struct {
	ID    string `json:"id"`
	Pid   string `json:"Pid"`
	Name  string `json:"name"`
	Label string `json:"label"`
}
