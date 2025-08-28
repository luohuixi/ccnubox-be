package data

import (
	"context"

	"github.com/asynccnu/ccnubox-be/be-library/internal/data/DO"
)

// CreateSeat 新增座位
func (r *SeatRepo) CreateSeat(ctx context.Context, s *DO.Seat) error {
	return r.data.db.WithContext(ctx).Create(s).Error
}

// GetSeatByDevID 根据 DevID 查询座位
func (r *SeatRepo) GetSeatByDevID(ctx context.Context, devID string) (*DO.Seat, error) {
	var s DO.Seat
	err := r.data.db.WithContext(ctx).Where("dev_id = ?", devID).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// UpdateSeatStatus 根据 DevID 更新状态
func (r *SeatRepo) UpdateSeatStatus(ctx context.Context, devID string, status string) error {
	return r.data.db.WithContext(ctx).
		Model(&DO.Seat{}).
		Where("dev_id = ?", devID).
		Update("status", status).Error
}

// DeleteSeatByDevID 删除座位
func (r *SeatRepo) DeleteSeatByDevID(ctx context.Context, devID string) error {
	return r.data.db.WithContext(ctx).Where("dev_id = ?", devID).Delete(&DO.Seat{}).Error
}

// CreateTimeSlot 新增时间段
func (r *SeatRepo) CreateTimeSlot(ctx context.Context, ts *DO.TimeSlot) error {
	return r.data.db.WithContext(ctx).Create(ts).Error
}

// GetTimeSlotsBySeatID 查询座位所有时间段
func (r *SeatRepo) GetTimeSlotsBySeatID(ctx context.Context, devID string) ([]*DO.TimeSlot, error) {
	var ts []DO.TimeSlot
	err := r.data.db.WithContext(ctx).Where("dev_id = ?", devID).Find(&ts).Error
	if err != nil {
		return nil, err
	}

	// []timeSlot -> []*timeSlot
	tss := make([]*DO.TimeSlot, 0, len(ts))
	for i := range ts {
		tss = append(tss, &ts[i])
	}

	return tss, nil
}

// UpdateTimeSlot 更新时间段信息
func (r *SeatRepo) UpdateTimeSlot(ctx context.Context, ts *DO.TimeSlot) error {
	return r.data.db.WithContext(ctx).Save(ts).Error
}

// DeleteTimeSlot 删除时间段
func (r *SeatRepo) DeleteTimeSlot(ctx context.Context, id uint) error {
	return r.data.db.WithContext(ctx).Where("id = ?", id).Delete(&DO.TimeSlot{}).Error
}
