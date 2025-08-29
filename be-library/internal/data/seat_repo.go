package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/asynccnu/ccnubox-be/be-library/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type seat struct {
	ID       uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	LabName  string `gorm:"size:100;not null" json:"lab_name"`
	RoomID   string `gorm:"size:100;not null" json:"kind_id"`
	RoomName string `gorm:"size:150;not null" json:"kind_name"`
	DevID    string `gorm:"size:50;not null;uniqueIndex" json:"dev_id"`
	DevName  string `gorm:"size:50;not null" json:"dev_name"`
	Status   string `json:"status"`
}

type timeSlot struct {
	ID    uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	DevID string `gorm:"index;not null" json:"seat_id"`
	Start string `gorm:"not null" json:"start"`
	End   string `gorm:"not null" json:"end"`
}

type SeatRepo struct {
	data *Data
	log  *log.Helper
}

func NewSeatRepo(data *Data, logger log.Logger) biz.SeatRepo {
	return &SeatRepo{
		log:  log.NewHelper(logger),
		data: data,
	}
}

// 弄个管理员账号来进行持续爬虫
func (r *SeatRepo) SaveRoomSeatsInRedis(ctx context.Context, stuID string) error {
	allSeats, err := r.data.crawler.GetSeatInfos(ctx, stuID)
	if err != nil {
		return err
	}

	// 按房间存储 房间里的所有座位数据
	for roomId, seats := range allSeats {
		key := fmt.Sprintf("room:%s", roomId)

		// seatID : seatJson
		hash := make(map[string]string)
		for _, seat := range seats {
			seatID := seat.DevID
			seatJson, err := json.Marshal(seat)
			if err != nil {
				r.log.Errorf("marshal seat error := %v", err)
				return err
			}
			hash[seatID] = string(seatJson)
		}

		// 存入 Redis
		// RoomID : {N1111: json1 N2222: json2}
		err := r.data.redis.HSet(ctx, key, seats).Err()
		if err != nil {
			r.log.Errorf("HSet room:%s error: %v", roomId, err)
			return err
		}
	}

	r.log.Infof("All seats saved in Redis successfully")
	return nil
}

func (r *SeatRepo) getRoomSeats(ctx context.Context, roomID int) ([]*biz.Seat, error) {
	roomKey := fmt.Sprintf("room:%d", roomID)

	data, err := r.data.redis.HGetAll(ctx, roomKey).Result()
	if err != nil {
		r.log.Errorf("get seatinfo from redis error (room_id := %s)", roomKey)
		return nil, err
	}

	seats := []*biz.Seat{}
	for _, v := range data {
		var s biz.Seat
		err := json.Unmarshal([]byte(v), &s)
		if err == nil {
			seats = append(seats, &s)
		}
	}
	return seats, nil
}

// 将座位信息分块存入redis里
// func (r *SeatRepo) SyncFromCrawlerInCache(ctx context.Context, roomID string, cookie string) error {
// 	SeatJson, roomID, err := r.data.crawler.SeatJSONCrawler(ctx, cookie, roomID)
// 	if err != nil {
// 		return err
// 	}

// 	err = r.SeatToCache(ctx, SeatJson, roomID)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

func (r *SeatRepo) SyncSeatsIntoSQL(ctx context.Context, roomID string, stuID string, seats []*biz.Seat) error {
	dataSeats, dataTimeslots := LotConvert2DataSeat(seats)
	err := r.SaveSeatsAndTimeSlots(ctx, dataSeats, dataTimeslots)
	if err != nil {
		r.log.Errorf("save seats and timeslots failed(room_id: %v) stu_id: %v", roomID, stuID)
		return err
	}

	r.log.Infof("save seats and timeslots successed(room_id: %v) stu_id: %v", roomID, stuID)
	return nil
}

func (r *SeatRepo) GetByRoom(ctx context.Context, roomID string) (string, error) {
	json, err := r.getSeatJSONFromCacheByDevID(ctx, roomID, false)
	if err != nil {
		return "", err
	}
	return json, nil
}

// 待优化
func (r *SeatRepo) FindFirstAvailableSeat(ctx context.Context, roomID, start, end string) (string, error) {
	var seatDevID string
	// 待优化
	subQuery := r.data.db.Model(&timeSlot{}).
		Select("1").
		Where("time_slots.dev_id = seats.id").
		Where("start < ?", end).
		Where("end > ?", start)

	err := r.data.db.WithContext(ctx).
		Model(&seat{}).
		Where("room_id = ?", roomID).
		Where("NOT EXISTS (?)", subQuery).
		Limit(1).
		Pluck("dev_id", &seatDevID).Error

	if err != nil {
		return "", err
	}
	if seatDevID == "" {
		return "", fmt.Errorf("no available seat found")
	}
	return seatDevID, nil
}

// Get 获取单个座位信息
func (r *SeatRepo) Get(ctx context.Context, devID string) (*biz.Seat, error) {
	// 先从缓存获取
	cacheKey := fmt.Sprintf("seat:%s", devID)
	cached, err := r.data.redis.Get(ctx, cacheKey).Result()
	if err == nil && cached != "" {
		var s biz.Seat
		if err := json.Unmarshal([]byte(cached), &s); err == nil {
			return &s, nil
		}
	}

	r.log.Errorf("Error getting seatInfo from redis")
	// 从数据库获取兜底
	result, err := r.getSeatFromSQL(ctx, devID)

	// 写入缓存
	if data, err := json.Marshal(result); err == nil {
		r.data.redis.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return result, nil
}

func (r *SeatRepo) getSeatFromSQL(ctx context.Context, devID string) (*biz.Seat, error) {
	var seatModel seat
	if err := r.data.db.WithContext(ctx).
		Where("dev_id = ?", devID).
		First(&seatModel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	ts, err := r.GetTimeSlotsBySeatID(ctx, devID)
	if err != nil {
		return nil, err
	}

	// 转换为业务模型
	result := ConvertSeat2Biz(&seatModel, ts)

	return result, nil
}

func (r *SeatRepo) toBizSeat(s *seat, ts []timeSlot) *biz.Seat {
	bizTs := make([]*biz.TimeSlot, len(ts))
	for _, t := range ts {
		bizT := &biz.TimeSlot{
			Start: t.Start,
			End:   t.End,
		}
		bizTs = append(bizTs, bizT)
	}

	result := &biz.Seat{
		DevID:    s.DevID,
		DevName:  s.DevName,
		LabName:  s.LabName,
		RoomID:   s.RoomID,
		RoomName: s.RoomName,
		Ts:       bizTs,
	}

	return result
}

func (r *SeatRepo) SaveSeatsAndTimeSlots(ctx context.Context, seats []*seat, timeSlots []*timeSlot) error {
	// 使用事务保证 seat timeSlot 插入数据一致性
	return r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 批量插入 seat
		if len(seats) > 0 {
			if err := tx.Create(&seats).Error; err != nil {
				return err
			}
		}

		// 批量插入 timeSlot
		if len(timeSlots) > 0 {
			if err := tx.Create(&timeSlots).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
