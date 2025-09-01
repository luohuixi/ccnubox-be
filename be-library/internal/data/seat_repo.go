package data

import (
	"context"
	"fmt"
	"time"

	"github.com/asynccnu/ccnubox-be/be-library/internal/biz"
	"github.com/asynccnu/ccnubox-be/be-library/internal/data/DO"
	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

type SeatRepo struct {
	data    *Data
	log     *log.Helper
	sf      singleflight.Group
	crawler biz.LibraryCrawler
}

func NewSeatRepo(data *Data, logger log.Logger, crawler biz.LibraryCrawler) biz.SeatRepo {
	return &SeatRepo{
		log:     log.NewHelper(logger),
		data:    data,
		crawler: crawler,
	}
}

// 弄个管理员账号来进行持续爬虫
func (r *SeatRepo) SaveRoomSeatsInRedis(ctx context.Context, stuID string) error {
	ttl := r.data.cfg.Redis.Ttl

	allSeats, err := r.crawler.GetSeatInfos(ctx, stuID)
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
		err := r.data.redis.HSet(ctx, key, hash).Err()
		if err != nil {
			r.log.Errorf("HSet room:%s error: %v", roomId, err)
			return err
		}

		// 设置 TTL , 过时自动删除捏
		err = r.data.redis.Expire(ctx, key, ttl.AsDuration()).Err()
		if err != nil {
			r.log.Errorf("Expire room:%s error: %v", roomId, err)
			return err
		}
	}

	r.log.Infof("All seats saved in Redis successfully")
	return nil
}

func (r *SeatRepo) GetSeatsByRoom(ctx context.Context, roomID string) ([]*biz.Seat, error) {
	roomKey := fmt.Sprintf("room:%s", roomID)

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

// 待优化x
func (r *SeatRepo) FindFirstAvailableSeat(ctx context.Context, roomID, start, end string) (string, error) {
	var seatDevID string
	// 待优化
	subQuery := r.data.db.Model(&DO.TimeSlot{}).
		Select("1").
		Where("time_slots.dev_id = seats.id").
		Where("start < ?", end).
		Where("end > ?", start)

	err := r.data.db.WithContext(ctx).
		Model(&DO.Seat{}).
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

// func (r *SeatRepo) getSeatFromSQL(ctx context.Context, devID string) (*biz.Seat, error) {
// 	var seatModel seat
// 	if err := r.data.db.WithContext(ctx).
// 		Where("dev_id = ?", devID).
// 		First(&seatModel).Error; err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return nil, nil
// 		}
// 		return nil, err
// 	}
// 	ts, err := r.GetTimeSlotsBySeatID(ctx, devID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// 转换为业务模型
// 	result := ConvertSeat2Biz(&seatModel, ts)

// 	return result, nil
// }

func (r *SeatRepo) toBizSeat(s *DO.Seat, ts []*DO.TimeSlot) *biz.Seat {
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

func (r *SeatRepo) SaveSeatsAndTimeSlots(ctx context.Context, seats []*DO.Seat, timeSlots []*DO.TimeSlot) error {
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

// GetSeatInfos 按楼层查缓存
func (r *SeatRepo) GetSeatInfos(ctx context.Context, stuID string) (map[string][]*biz.Seat, error) {
	now := time.Now()
	result := make(map[string][]*biz.Seat, len(biz.RoomIDs))

	// 是否有房间命中缓存
	hitAny := false
	// 是否需要后台刷新
	needRefresh := false

	for _, roomID := range biz.RoomIDs {
		seats, ts, ok, err := r.getRoomSeatsCache(ctx, roomID)
		if err != nil {
			r.log.Warnf("get room seats cache(room_id:%s) err: %v", roomID, err)
			needRefresh = true
			continue
		}
		if !ok {
			needRefresh = true
			continue
		}

		// 命中缓存
		result[roomID] = seats
		hitAny = true

		// 判断软过期
		if ts.IsZero() || now.Sub(ts) > freshness {
			needRefresh = true
		}
	}

	if hitAny {
		// 返回缓存同时在后台刷新
		if needRefresh {
			go func() {
				bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				_, _, _ = r.sf.Do("lib:getSeatInfos:refresh", func() (interface{}, error) {
					_, err := r.refreshAll(bgCtx, stuID)
					return nil, err
				})
			}()
		}
		return result, nil
	}

	// 走到这里说明完全没有缓存,阻塞一次并拉取座位信息
	val, err, _ := r.sf.Do("lib:getSeatInfos:refresh", func() (interface{}, error) {
		ctx2, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		return r.refreshAll(ctx2, stuID)
	})
	if err != nil {
		return nil, err
	}
	return val.(map[string][]*biz.Seat), nil
}

// refreshAll 从爬虫获取所有房间最新座位信息并回填缓存与时间戳
func (r *SeatRepo) refreshAll(ctx context.Context, stuID string) (map[string][]*biz.Seat, error) {
	data, err := r.crawler.GetSeatInfos(ctx, stuID)
	if err != nil {
		return nil, err
	}
	now := time.Now()

	for roomID, seats := range data {
		// 回填缓存
		if err := r.setRoomSeatsCache(ctx, roomID, seats, now); err != nil {
			r.log.Warnf("set room seats cache(room_id:%s) err: %v", roomID, err)
		}
	}
	return data, nil
}
