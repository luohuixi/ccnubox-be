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
	cacheKeyRoomSeatsFmt   = "lib:room:%s:seats"
	cacheKeyRoomSeatsTsFmt = "lib:room:%s:seats:ts"
	// 硬过期，保证夜间不丢缓存
	seatsHardTTL = 24 * time.Hour
	// 软过期，超时则视为需要刷新
	seatsFreshness = 30 * time.Second
)

// func (c *Crawler) SeatJSONCrawler(ctx context.Context, cookie string, roomid string) (JSON string, roomID string, err error) {
//  baseURL := "http://kjyy.ccnu.edu.cn/ClientWeb/pro/ajax/device.aspx"

//  date := time.Now().Format("2006-01-02")

//  params := url.Values{}
//  params.Set("classkind", "8")
//  params.Set("room_id", roomid)
//  params.Set("date", date)
//  params.Set("act", "get_rsv_sta")

//  fullURL := baseURL + "?" + params.Encode()

//  req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
//  if err != nil {
//     log.Fatal("创建请求失败:", err)
//  }

//  req.Header.Set("cookie", cookie)

//  resp, err := c.client.Do(req)
//  if err != nil {
//     return "", "", errcode.ErrCrawler
//  }
//  defer resp.Body.Close()

//  body, err := io.ReadAll(resp.Body)
//  if err != nil {
//     return "", "", err
//  }

//  return string(body), "", err
// }

func (c *SeatRepo) SeatToCache(ctx context.Context, JSON string, roomID string) error {
	// 用 map 解析最外层，只取 data
	var parsed map[string]json.RawMessage
	err := json.Unmarshal([]byte(JSON), &parsed)
	if err != nil {
		return err
	}

	// data 是一个数组
	var seats []json.RawMessage
	err = json.Unmarshal(parsed["data"], &seats)
	if err != nil {
		return err
	}

	seatBytes, err := json.Marshal(seats)
	if err != nil {
		return err
	}

	// 这里的 data 是整个房间的座位数据，直接根据房间号存入redis
	key := fmt.Sprintf("room:%s", roomID)
	err = c.data.redis.Set(ctx, key, seatBytes, 5*time.Minute).Err()
	if err != nil {
		return err
	}

	// 遍历每一个座位 JSON
	for _, seatJSON := range seats {
		// 取出 devId
		var seatMap map[string]json.RawMessage
		err = json.Unmarshal(seatJSON, &seatMap)
		if err != nil {
			return err
		}

		var devID string
		err = json.Unmarshal(seatMap["devId"], &devID)
		if err != nil {
			return err
		}

		// 直接存入 Redis
		key = fmt.Sprintf("seat:%s", devID)
		err = c.data.redis.Set(ctx, key, seatJSON, 5*time.Minute).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

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

func (r *SeatRepo) getSeatCache(ctx context.Context, devID string) (*biz.Seat, bool, error) {
	key := fmt.Sprintf("seat:%s", devID)
	val, err := r.data.redis.Get(ctx, key).Bytes()
	if err != nil || len(val) == 0 {
		return nil, false, nil
	}
	var s biz.Seat
	if err = json.Unmarshal(val, &s); err != nil {
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

func (r *SeatRepo) cacheRoomSeatsKey(roomID string) string {
	return fmt.Sprintf(cacheKeyRoomSeatsFmt, roomID)
}

func (r *SeatRepo) cacheRoomSeatsTsKey(roomID string) string {
	return fmt.Sprintf(cacheKeyRoomSeatsTsFmt, roomID)
}

func (r *SeatRepo) getRoomSeatsCache(ctx context.Context, roomID string) ([]*biz.Seat, time.Time, bool, error) {
	seatsKey := r.cacheRoomSeatsKey(roomID)
	tsKey := r.cacheRoomSeatsTsKey(roomID)

	// 读取 seats
	raw, err := r.data.redis.Get(ctx, seatsKey).Bytes()
	if err != nil || len(raw) == 0 {
		return nil, time.Time{}, false, nil
	}
	var seats []*biz.Seat
	if err = json.Unmarshal(raw, &seats); err != nil {
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
	tsKey := r.cacheRoomSeatsTsKey(roomID)

	b, err := json.Marshal(seats)
	if err != nil {
		return err
	}
	if err = r.data.redis.Set(ctx, seatsKey, b, seatsHardTTL).Err(); err != nil {
		return err
	}
	if err = r.data.redis.Set(ctx, tsKey, ts.Format(time.RFC3339Nano), seatsHardTTL).Err(); err != nil {
		return err
	}

	return nil
}

func (r *SeatRepo) delRoomSeatsCache(ctx context.Context, roomID string) error {
	seatsKey := r.cacheRoomSeatsKey(roomID)
	tsKey := r.cacheRoomSeatsTsKey(roomID)

	_, err := r.data.redis.Del(ctx, seatsKey, tsKey).Result()

	return err
}
