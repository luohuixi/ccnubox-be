package data

import (
	"github.com/asynccnu/ccnubox-be/be-library/internal/biz"
	"github.com/asynccnu/ccnubox-be/be-library/internal/data/DO"
)

// ConvertSeat 将 data 层的 seat 转为 biz 层的 Seat
func ConvertDOSeatBiz(s *DO.Seat, slots []*DO.TimeSlot) *biz.Seat {
	ts := make([]*biz.TimeSlot, 0, len(slots))
	for _, slot := range slots {
		if slot.DevID != s.DevID {
			continue
		}
		ts = append(ts, &biz.TimeSlot{
			Start:  slot.Start,
			End:    slot.End,
			State:  s.Status,                // 或者有其它逻辑决定 state
			Owner:  "",                      // 是否需要 Owner?
			Occupy: s.Status != "available", // 示例逻辑
		})
	}

	return &biz.Seat{
		LabName:  s.LabName,
		RoomID:   s.RoomID,
		RoomName: s.RoomName,
		DevID:    s.DevID,
		DevName:  s.DevName,
		Ts:       ts,
	}
}

func LotConvertBizSeatDO(seats []*biz.Seat) ([]*DO.Seat, []*DO.TimeSlot) {
	resultSeat := make([]*DO.Seat, 0, len(seats))
	resultTimeSlots := make([]*DO.TimeSlot, 0, len(seats))

	for _, seat := range seats {
		resultSeat = append(resultSeat, ConvertBizSeatDO(seat))
		resultTimeSlots = append(resultTimeSlots, ConvertBizTimeSlotsDO(seat.DevID, seat.Ts)...)
	}

	return resultSeat, resultTimeSlots
}

func ConvertBizSeatDO(s *biz.Seat) *DO.Seat {
	status := "available"
	if len(s.Ts) > 0 {
		// 简单逻辑：如果有占用的 timeSlot，就标记为 busy，待替换
		for _, t := range s.Ts {
			if t.Occupy {
				status = "busy"
				break
			}
		}
	}

	return &DO.Seat{
		LabName:  s.LabName,
		RoomID:   s.RoomID,
		RoomName: s.RoomName,
		DevID:    s.DevID,
		DevName:  s.DevName,
		Status:   status,
	}
}

func ConvertBizTimeSlotsDO(seatID string, ts []*biz.TimeSlot) []*DO.TimeSlot {
	result := make([]*DO.TimeSlot, 0, len(ts))
	for _, t := range ts {
		result = append(result, &DO.TimeSlot{
			DevID: seatID,
			Start: t.Start,
			End:   t.End,
		})
	}
	return result
}

func ConvertDOFutureRecordsBiz(dos []DO.FutureRecord) []*biz.FutureRecords {
	out := make([]*biz.FutureRecords, 0, len(dos))
	for _, d := range dos {
		out = append(out, &biz.FutureRecords{
			ID:       d.ID,
			Owner:    d.Owner,
			Start:    d.Start,
			End:      d.End,
			TimeDesc: d.TimeDesc,
			States:   d.States,
			DevName:  d.DevName,
			RoomID:   d.RoomID,
			RoomName: d.RoomName,
			LabName:  d.LabName,
		})
	}
	return out
}

func ConvertBizFutureRecordsDO(stuID string, list []*biz.FutureRecords) []DO.FutureRecord {
	dos := make([]DO.FutureRecord, 0, len(list))
	for _, it := range list {
		dos = append(dos, DO.FutureRecord{
			StuID:    stuID,
			ID:       it.ID,
			Owner:    it.Owner,
			Start:    it.Start,
			End:      it.End,
			TimeDesc: it.TimeDesc,
			States:   it.States,
			DevName:  it.DevName,
			RoomID:   it.RoomID,
			RoomName: it.RoomName,
			LabName:  it.LabName,
		})
	}
	return dos
}

func ConvertDOHistoryRecordsBiz(dos []DO.HistoryRecord) []*biz.HistoryRecords {
	out := make([]*biz.HistoryRecords, 0, len(dos))
	for _, d := range dos {
		out = append(out, &biz.HistoryRecords{
			Place:      d.Place,
			Floor:      d.Floor,
			Status:     d.Status,
			Date:       d.Date,
			SubmitTime: d.SubmitTime,
		})
	}
	return out
}

func ConvertBizHistoryRecordsDO(stuID string, list []*biz.HistoryRecords) []DO.HistoryRecord {
	dos := make([]DO.HistoryRecord, 0, len(list))
	for _, it := range list {
		dos = append(dos, DO.HistoryRecord{
			StuID:      stuID,
			Place:      it.Place,
			Floor:      it.Floor,
			Status:     it.Status,
			Date:       it.Date,
			SubmitTime: it.SubmitTime,
		})
	}
	return dos
}

func ConvertDOCreditPointsBiz(summary *DO.CreditSummary, records []DO.CreditRecord) *biz.CreditPoints {
	if summary == nil {
		return &biz.CreditPoints{Summary: nil, Records: nil}
	}
	out := &biz.CreditPoints{
		Summary: &biz.CreditSummary{
			System: summary.System,
			Remain: summary.Remain,
			Total:  summary.Total,
		},
		Records: make([]*biz.CreditRecord, 0, len(records)),
	}
	for _, r := range records {
		out.Records = append(out.Records, &biz.CreditRecord{
			Title:    r.Title,
			Subtitle: r.Subtitle,
			Location: r.Location,
		})
	}
	return out
}

func ConvertBizCreditPointsDO(stuID string, cp *biz.CreditPoints) (*DO.CreditSummary, []DO.CreditRecord) {
	if cp == nil || cp.Summary == nil {
		return nil, nil
	}
	sum := &DO.CreditSummary{
		StuID:  stuID,
		System: cp.Summary.System,
		Remain: cp.Summary.Remain,
		Total:  cp.Summary.Total,
	}
	var recs []DO.CreditRecord
	for _, r := range cp.Records {
		recs = append(recs, DO.CreditRecord{
			StuID:    stuID,
			Title:    r.Title,
			Subtitle: r.Subtitle,
			Location: r.Location,
		})
	}
	return sum, recs
}

func ConvertDODiscussionBiz(dos []*DO.Discussion) []*biz.Discussion {
	out := make([]*biz.Discussion, 0, len(dos))
	for _, d := range dos {
		item := &biz.Discussion{
			LabID:    d.LabID,
			LabName:  d.LabName,
			KindID:   d.KindID,
			KindName: d.KindName,
			DevID:    d.DevID,
			DevName:  d.DevName,
			TS:       make([]*biz.DiscussionTS, 0, len(d.TS)),
		}
		for _, t := range d.TS {
			item.TS = append(item.TS, &biz.DiscussionTS{
				Start:  t.Start,
				End:    t.End,
				State:  t.State,
				Title:  t.Title,
				Owner:  t.Owner,
				Occupy: t.Occupy,
			})
		}
		out = append(out, item)
	}
	return out
}

func ConvertBizDiscussionDO(list []*biz.Discussion) []*DO.Discussion {
	out := make([]*DO.Discussion, 0, len(list))
	for _, d := range list {
		item := &DO.Discussion{
			LabID:    d.LabID,
			LabName:  d.LabName,
			KindID:   d.KindID,
			KindName: d.KindName,
			DevID:    d.DevID,
			DevName:  d.DevName,
			TS:       make([]*DO.DiscussionTS, 0, len(d.TS)),
		}
		for _, t := range d.TS {
			item.TS = append(item.TS, &DO.DiscussionTS{
				Start:  t.Start,
				End:    t.End,
				State:  t.State,
				Title:  t.Title,
				Owner:  t.Owner,
				Occupy: t.Occupy,
			})
		}
		out = append(out, item)
	}
	return out
}
