package data

import "github.com/asynccnu/ccnubox-be/be-library/internal/biz"

// ConvertSeat 将 data 层的 seat 转为 biz 层的 Seat
func ConvertSeat2Biz(s *seat, slots []*timeSlot) *biz.Seat {
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

func LotConvert2DataSeat(seats []*biz.Seat) ([]*seat, []*timeSlot) {
	resultSeat := make([]*seat, 0, len(seats))
	resultTimeSlots := make([]*timeSlot, 0, len(seats))

	for _, seat := range seats {
		resultSeat = append(resultSeat, Convert2DataSeat(seat))
		resultTimeSlots = append(resultTimeSlots, Convert2DataTimeSlots(seat.DevID, seat.Ts)...)
	}

	return resultSeat, resultTimeSlots
}

func Convert2DataSeat(s *biz.Seat) *seat {
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

	return &seat{
		LabName:  s.LabName,
		RoomID:   s.RoomID,
		RoomName: s.RoomName,
		DevID:    s.DevID,
		DevName:  s.DevName,
		Status:   status,
	}
}

func Convert2DataTimeSlots(seatID string, ts []*biz.TimeSlot) []*timeSlot {
	result := make([]*timeSlot, 0, len(ts))
	for _, t := range ts {
		result = append(result, &timeSlot{
			DevID: seatID,
			Start: t.Start,
			End:   t.End,
		})
	}
	return result
}
