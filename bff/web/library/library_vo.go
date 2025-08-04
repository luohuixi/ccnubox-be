package library

type GetSeatRequest struct {
	RoomID string `json:"room_id"`
	StuID  string `json:"stu_id"`
}

type GetSeatResponse struct {
	Rooms []Room `json:"rooms"`
}

type Room struct {
	RoomID string `json:"room_id"`
	Seats  []Seat `json:"seats"`
}

type Seat struct {
	LabName   string     `json:"labName"`
	KindName  string     `json:"kindName"`
	DevID     string     `json:"devId"`
	DevName   string     `json:"devName"`
	TimeSlots []TimeSlot `json:"ts"`
}

type TimeSlot struct {
	Start  string `json:"start"`
	End    string `json:"end"`
	State  string `json:"state"`
	Owner  string `json:"owner"`
	Occupy bool   `json:"occupy"`
}

type ReserveSeatRequest struct {
	DevID string `json:"dev_id"`
	Start string `json:"start"`
	End   string `json:"end"`
	StuID string `json:"stu_id"`
}

type ReserveSeatResponse struct {
	Message string `json:"message"`
}

type GetSeatRecordRequest struct {
	StuID string `json:"stu_id"`
}

type GetSeatRecordResponse struct {
	Records []Record `json:"records"`
}

type Record struct {
	ID       string `json:"id"`
	Owner    string `json:"owner"`
	Start    string `json:"start"`
	End      string `json:"end"`
	TimeDesc string `json:"timeDesc"`
	Occur    string `json:"occur"`
	States   string `json:"states"`
	DevName  string `json:"devName"`
	RoomID   string `json:"roomId"`
	RoomName string `json:"roomName"`
	LabName  string `json:"labName"`
}

type GetHistoryRequest struct {
	StuID string `json:"stu_id"`
}

type GetHistoryResponse struct {
	Histories []History `json:"history"`
}

type History struct {
	Place      string `json:"place"`
	Floor      string `json:"floor"`
	Status     string `json:"status"`
	Date       string `json:"date"`
	SubmitTime string `json:"submitTime"`
}

type CancelSeatRequest struct {
	ID    string `json:"id"`
	StuID string `json:"stu_id"`
}

type CancelSeatResponse struct {
	Message string `json:"message"`
}

type GetCreditPointRequest struct {
	StuID string `json:"stu_id"`
}

type GetCreditPointResponse struct {
	CreditPoints CreditPoints
}

type CreditPoints struct {
	Summary CreditSummary  `json:"summary"`
	Records []CreditRecord `json:"records"`
}

type CreditSummary struct {
	System string `json:"system"` // 个人预约制度
	Remain string `json:"remain"`
	Total  string `json:"total"`
}

type CreditRecord struct {
	Title    string `json:"title"`    // 原因标题
	Subtitle string `json:"subtitle"` // 扣分及时间
	Location string `json:"location"` // 地点及备注
}

type GetDiscussionRequest struct {
	ClassID string `json:"class_id"`
	Date    string `json:"date"`
	StuID   string `json:"stu_id"`
}

type GetDiscussionResponse struct {
	Discussions []Discussion
}

type Discussion struct {
	LabName  string `json:"labName"`
	KindName string `json:"kindName"`
	DevID    string `json:"devId"`
	DevName  string `json:"devName"`
	TS       []DiscussionTS
}

type DiscussionTS struct {
	Start  string `json:"start"`
	End    string `json:"end"`
	State  string `json:"state"`
	Title  string `json:"title"`
	Owner  string `json:"owner"`
	Occupy bool   `json:"occupy"`
}

type SearchUserRequest struct {
	StudentID string `json:"student_id"`
	StuID     string `json:"stu_id"`
}

type SearchUserResponse struct {
	Search Search
}

type Search struct {
	ID    string `json:"id"`
	Pid   string `json:"Pid"`
	Name  string `json:"name"`
	Label string `json:"label"`
}

type ReserveDiscussionRequest struct {
	DevID  string   `json:"dev_id"`
	LabID  string   `json:"lab_id"`
	KindID string   `json:"kind_id"`
	Title  string   `json:"title"`
	Start  string   `json:"start"`
	End    string   `json:"end"`
	List   []string `json:"list"`
	StuID  string   `json:"stu_id"`
}

type ReserveDiscussionResponse struct {
	Message string `json:"message"`
}

type CancelDiscussionRequest struct {
	ID    string `json:"id"`
	StuID string `json:"stu_id"`
}

type CancelDiscussionResponse struct {
	Message string `json:"message"`
}
