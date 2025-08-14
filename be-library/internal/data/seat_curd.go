package data

import "context"

// CreateSeat 新增座位
func (r *seatRepo) CreateSeat(ctx context.Context, s *seat) error {
	return r.data.db.WithContext(ctx).Create(s).Error
}

// GetSeatByDevID 根据 DevID 查询座位
func (r *seatRepo) GetSeatByDevID(ctx context.Context, devID string) (*seat, error) {
	var s seat
	err := r.data.db.WithContext(ctx).Where("dev_id = ?", devID).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// UpdateSeatStatus 根据 DevID 更新状态
func (r *seatRepo) UpdateSeatStatus(ctx context.Context, devID string, status string) error {
	return r.data.db.WithContext(ctx).
		Model(&seat{}).
		Where("dev_id = ?", devID).
		Update("status", status).Error
}

// DeleteSeatByDevID 删除座位
func (r *seatRepo) DeleteSeatByDevID(ctx context.Context, devID string) error {
	return r.data.db.WithContext(ctx).Where("dev_id = ?", devID).Delete(&seat{}).Error
}

// CreateTimeSlot 新增时间段
func (r *seatRepo) CreateTimeSlot(ctx context.Context, ts *timeSlot) error {
	return r.data.db.WithContext(ctx).Create(ts).Error
}

// GetTimeSlotsBySeatID 查询座位所有时间段
func (r *seatRepo) GetTimeSlotsBySeatID(ctx context.Context, devID string) ([]timeSlot, error) {
	var ts []timeSlot
	err := r.data.db.WithContext(ctx).Where("dev_id = ?", devID).Find(&ts).Error
	if err != nil {
		return nil, err
	}
	return ts, nil
}

// UpdateTimeSlot 更新时间段信息
func (r *seatRepo) UpdateTimeSlot(ctx context.Context, ts *timeSlot) error {
	return r.data.db.WithContext(ctx).Save(ts).Error
}

// DeleteTimeSlot 删除时间段
func (r *seatRepo) DeleteTimeSlot(ctx context.Context, id uint) error {
	return r.data.db.WithContext(ctx).Where("id = ?", id).Delete(&timeSlot{}).Error
}
