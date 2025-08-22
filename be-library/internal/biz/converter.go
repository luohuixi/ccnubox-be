package biz

import pb "github.com/asynccnu/ccnubox-be/be-api/gen/proto/library/v1"

type Converter struct{}

func NewConverter() *Converter {
	return &Converter{}
}

func (c *Converter) ConvertTimeSlots(src []*TimeSlot) []*pb.TimeSlot {
	if len(src) == 0 {
		return nil
	}
	result := make([]*pb.TimeSlot, 0, len(src)) // 预分配容量
	for _, ts := range src {
		result = append(result, &pb.TimeSlot{
			Start:  ts.Start,
			End:    ts.End,
			State:  ts.State,
			Owner:  ts.Owner,
			Occupy: ts.Occupy,
		})
	}
	return result
}

func (c *Converter) ConvertDiscussionTS(src []*DiscussionTS) []*pb.DiscussionTS {
	if len(src) == 0 {
		return nil
	}
	result := make([]*pb.DiscussionTS, 0, len(src))
	for _, ts := range src {
		result = append(result, &pb.DiscussionTS{
			Start:  ts.Start,
			End:    ts.End,
			State:  ts.State,
			Title:  ts.Title,
			Owner:  ts.Owner,
			Occupy: ts.Occupy,
		})
	}
	return result
}

func (c *Converter) ConvertRecords(src []*FutureRecords) []*pb.Record {
	if len(src) == 0 {
		return nil
	}
	result := make([]*pb.Record, 0, len(src))
	for _, r := range src {
		result = append(result, &pb.Record{
			Id:       r.ID,
			Owner:    r.Owner,
			Start:    r.Start,
			End:      r.End,
			TimeDesc: r.TimeDesc,
			States:   r.States,
			DevName:  r.DevName,
			RoomId:   r.RoomID,
			RoomName: r.RoomName,
			LabName:  r.LabName,
		})
	}
	return result
}

func (c *Converter) ConvertHistory(src []*HistoryRecords) []*pb.History {
	if len(src) == 0 {
		return nil
	}
	result := make([]*pb.History, 0, len(src))
	for _, h := range src {
		result = append(result, &pb.History{
			Place:      h.Place,
			Floor:      h.Floor,
			Status:     h.Status,
			Date:       h.Date,
			SubmitTime: h.SubmitTime,
		})
	}
	return result
}

func (c *Converter) ConvertCreditRecords(src []*CreditRecord) []*pb.CreditRecord {
	if len(src) == 0 {
		return nil
	}
	result := make([]*pb.CreditRecord, 0, len(src))
	for _, r := range src {
		result = append(result, &pb.CreditRecord{
			Title:    r.Title,
			Subtitle: r.Subtitle,
			Location: r.Location,
		})
	}
	return result
}

// 复合转换函数
func (c *Converter) ConvertSeats(src []*Seat) []*pb.Seat {
	if len(src) == 0 {
		return nil
	}
	result := make([]*pb.Seat, 0, len(src))
	for _, seat := range src {
		result = append(result, &pb.Seat{
			LabName:  seat.LabName,
			KindName: seat.RoomName,
			DevId:    seat.DevID,
			DevName:  seat.DevName,
			Ts:       c.ConvertTimeSlots(seat.Ts),
		})
	}
	return result
}

func (c *Converter) ConvertDiscussions(src []*Discussion) []*pb.Discussion {
	if len(src) == 0 {
		return nil
	}
	result := make([]*pb.Discussion, 0, len(src))
	for _, d := range src {
		result = append(result, &pb.Discussion{
			LabId:    d.LabID,
			LabName:  d.LabName,
			KindId:   d.KindID,
			KindName: d.KindName,
			DevId:    d.DevID,
			DevName:  d.DevName,
			TS:       c.ConvertDiscussionTS(d.TS),
		})
	}
	return result
}

// 完整响应转换
func (c *Converter) ConvertGetSeatResponse(data map[string][]*Seat) *pb.GetSeatResponse {
	if len(data) == 0 {
		return &pb.GetSeatResponse{}
	}
	result := &pb.GetSeatResponse{
		RoomSeats: make([]*pb.RoomSeat, 0, len(data)),
	}
	for roomID, seats := range data {
		result.RoomSeats = append(result.RoomSeats, &pb.RoomSeat{
			RoomId: roomID,
			Seats:  c.ConvertSeats(seats),
		})
	}
	return result
}
