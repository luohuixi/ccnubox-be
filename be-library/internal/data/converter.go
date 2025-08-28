package data

import (
	"github.com/asynccnu/ccnubox-be/be-library/internal/biz"
	"github.com/asynccnu/ccnubox-be/be-library/internal/data/DO"
)

// ConvertSeat 将 data 层的 seat 转为 biz 层的 Seat
func ConvertSeat2Biz(s *DO.Seat, slots []*DO.TimeSlot) *biz.Seat {
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

func LotConvert2DataSeat(seats []*biz.Seat) ([]*DO.Seat, []*DO.TimeSlot) {
	resultSeat := make([]*DO.Seat, 0, len(seats))
	resultTimeSlots := make([]*DO.TimeSlot, 0, len(seats))

	for _, seat := range seats {
		resultSeat = append(resultSeat, Convert2DataSeat(seat))
		resultTimeSlots = append(resultTimeSlots, Convert2DataTimeSlots(seat.DevID, seat.Ts)...)
	}

	return resultSeat, resultTimeSlots
}

func Convert2DataSeat(s *biz.Seat) *DO.Seat {
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

func Convert2DataTimeSlots(seatID string, ts []*biz.TimeSlot) []*DO.TimeSlot {
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
