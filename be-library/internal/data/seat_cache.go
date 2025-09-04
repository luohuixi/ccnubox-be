package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/asynccnu/ccnubox-be/be-library/internal/biz"
	"github.com/asynccnu/ccnubox-be/be-library/internal/data/DO"
)

const (
	// HSet seatID : seatJson
	cacheKeyRoomSeatsFmt = "room:%s:seats"
	// ZSet
	cacheKeySeatTsFmt = "room:%s:seat:%s:times"
	// HSeat roomID : timestamp
	cacheKeyRoomTsFmt = "room:%s:time"
	// 硬过期，保证夜间不丢缓存
	hardTTL = 24 * time.Hour
	// 软过期，超时则视为需要刷新
	freshness = 30 * time.Second
)

// 布尔为真即查询单个座位，布尔为假即查询一整个房间
func (r *SeatRepo) getSeatJSONFromCacheByDevID(ctx context.Context, ID string, DevOrRoom bool) (string, error) {
	if true {
		val, err := r.data.redis.Get(ctx, fmt.Sprintf("seat:%s", ID)).Result()
		if err != nil {
			return "", err
		}
		return val, nil
	} else {
		val, err := r.data.redis.Get(ctx, fmt.Sprintf("room:%s", ID)).Result()
		if err != nil {
			return "", err
		}
		return val, nil
	}

}

// 单个座位的 JSON 解码为结构体
func (r *SeatRepo) seatJsonToSeat(ctx context.Context, JSON string) (*DO.Seat, error) {
	var seat DO.Seat
	err := json.Unmarshal([]byte(JSON), &seat)
	if err != nil {
		return nil, err
	}
	return &seat, nil
}

// 清除座位缓存
func (r *SeatRepo) clearSeatCache(ctx context.Context, devID string) {
	keys := []string{
		fmt.Sprintf("seat:%s", devID),
	}
	for _, key := range keys {
		r.data.redis.Del(ctx, key)
	}
}

// 清除房间缓存
func (r *SeatRepo) clearRoomCache(ctx context.Context, roomID string) {
	keys := []string{
		fmt.Sprintf("room:%s", roomID),
	}
	for _, key := range keys {
		r.data.redis.Del(ctx, key)
	}
}

// 封装：单座位缓存 get/set/del
func (r *SeatRepo) getSeatCache(ctx context.Context, devID string) (*biz.Seat, bool, error) {
	key := fmt.Sprintf("seat:%s", devID)
	val, err := r.data.redis.Get(ctx, key).Bytes()
	if err != nil || len(val) == 0 {
		return nil, false, nil
	}
	var s biz.Seat
	if err := json.Unmarshal(val, &s); err != nil {
		return nil, false, err
	}
	return &s, true, nil
}

func (r *SeatRepo) setSeatCache(ctx context.Context, devID string, seat *biz.Seat, ttl time.Duration) error {
	key := fmt.Sprintf("seat:%s", devID)
	b, err := json.Marshal(seat)
	if err != nil {
		return err
	}
	return r.data.redis.Set(ctx, key, b, ttl).Err()
}

func (r *SeatRepo) delSeatCache(ctx context.Context, devID string) error {
	key := fmt.Sprintf("seat:%s", devID)
	return r.data.redis.Del(ctx, key).Err()
}

// 房间级缓存 key 生成（与 seat_repo.go 中的 fmt 保持一致）
func (r *SeatRepo) cacheRoomSeatsKey(roomID string) string {
	return fmt.Sprintf(cacheKeyRoomSeatsFmt, roomID)
}
func (r *SeatRepo) cacheSeatTsKey(roomID, seatID string) string {
	return fmt.Sprintf(cacheKeySeatTsFmt, roomID, seatID)
}
func (r *SeatRepo) cacheRoomTsKey(roomID string) string {
	return fmt.Sprintf(cacheKeyRoomTsFmt)
}

// 封装：房间级缓存 get/set/del（seats + ts）
func (r *SeatRepo) getRoomSeatsCache(ctx context.Context, roomID string) ([]*biz.Seat, time.Time, bool, error) {
	seatsKey := r.cacheRoomSeatsKey(roomID)

	// 读取 seats
	raw, err := r.data.redis.Get(ctx, seatsKey).Bytes()
	if err != nil || len(raw) == 0 {
		return nil, time.Time{}, false, nil
	}
	var seats []*biz.Seat
	if err := json.Unmarshal(raw, &seats); err != nil {
		return nil, time.Time{}, false, err
	}

	// 读取 ts
	tsStr, err := r.data.redis.Get(ctx, tsKey).Result()
	if err != nil || tsStr == "" {
		return seats, time.Time{}, true, nil
	}
	ts, err := time.Parse(time.RFC3339Nano, tsStr)
	if err != nil {
		return seats, time.Time{}, true, nil
	}

	return seats, ts, true, nil
}

func (r *SeatRepo) setRoomSeatsCache(ctx context.Context, roomID string, seats []*biz.Seat, ts time.Time) error {
	seatsKey := r.cacheRoomSeatsKey(roomID)
	tsKey := r.cacheRoomTsKey(roomID)

	b, err := json.Marshal(seats)
	if err != nil {
		return err
	}
	if err := r.data.redis.Set(ctx, seatsKey, b, hardTTL).Err(); err != nil {
		return err
	}
	if err := r.data.redis.Set(ctx, tsKey, ts.Format(time.RFC3339Nano), hardTTL).Err(); err != nil {
		return err
	}
	return nil
}

func (r *SeatRepo) delRoomSeatsCache(ctx context.Context, roomID string) error {
	seatsKey := r.cacheRoomSeatsKey(roomID)
	tsKey := r.cacheRoomTsKey(roomID)
	_, err := r.data.redis.Del(ctx, seatsKey, tsKey).Result()
	return err
}
