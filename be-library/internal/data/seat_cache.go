package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/asynccnu/ccnubox-be/be-library/internal/data/DO"
)

// func (c *Crawler) SeatJSONCrawler(ctx context.Context, cookie string, roomid string) (JSON string, roomID string, err error) {
// 	baseURL := "http://kjyy.ccnu.edu.cn/ClientWeb/pro/ajax/device.aspx"

// 	date := time.Now().Format("2006-01-02")

// 	params := url.Values{}
// 	params.Set("classkind", "8")
// 	params.Set("room_id", roomid)
// 	params.Set("date", date)
// 	params.Set("act", "get_rsv_sta")

// 	fullURL := baseURL + "?" + params.Encode()

// 	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
// 	if err != nil {
// 		log.Fatal("创建请求失败:", err)
// 	}

// 	req.Header.Set("cookie", cookie)

// 	resp, err := c.client.Do(req)
// 	if err != nil {
// 		return "", "", errcode.ErrCrawler
// 	}
// 	defer resp.Body.Close()

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return "", "", err
// 	}

// 	return string(body), "", err
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
		err := json.Unmarshal(seatJSON, &seatMap)
		if err != nil {
			return err
		}

		var devID string
		err = json.Unmarshal(seatMap["devId"], &devID)
		if err != nil {
			return err
		}

		// 直接存入 Redis
		key := fmt.Sprintf("seat:%s", devID)
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
