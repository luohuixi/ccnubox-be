package DO

type Discussion struct {
	LabID    string
	LabName  string
	KindID   string
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
	ID    string
	Pid   string
	Name  string
	Label string
}
