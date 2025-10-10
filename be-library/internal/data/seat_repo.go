package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/asynccnu/ccnubox-be/be-library/internal/biz"
	"github.com/asynccnu/ccnubox-be/be-library/pkg/tool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

const (
	cacheKeyRoomFmt     = "lib:room:%s"
	cacheKeyRoomSeatFmt = "lib:room:%s:seat:%s"
	// 硬过期，保证夜间不丢缓存
	seatsHardTTL = 24 * time.Hour
	// 软过期，超时则视为需要刷新
	seatsFreshness = 30 * time.Second
)

type SeatRepo struct {
	data    *Data
	sf      singleflight.Group
	crawler biz.LibraryCrawler
}

func NewSeatRepo(data *Data, crawler biz.LibraryCrawler) biz.SeatRepo {
	return &SeatRepo{
		data:    data,
		crawler: crawler,
	}
}

func (r *SeatRepo) cacheRoomSeatsKey(roomID string) string {
	return fmt.Sprintf(cacheKeyRoomFmt, roomID)
}

func (r *SeatRepo) cacheRoomSeatsTsKey(roomID string, seatID string) string {
	return fmt.Sprintf(cacheKeyRoomSeatFmt, roomID, seatID)
}

// 弄个管理员账号来进行持续爬虫
// ZADD seat:{seatID}:times startTimestamp "{start}-{end}"
// HSET roomid timestamp(UnixMilli)
func (r *SeatRepo) SaveRoomSeatsInRedis(ctx context.Context, stuID string) error {
	ttl := r.data.cfg.Redis.Ttl
	// 用pipe收集redis指令，减少网络IO造成的时间损耗
	pipe := r.data.redis.Pipeline()

	allSeats, err := r.crawler.GetSeatInfos(ctx, stuID)
	if err != nil {
		return err
	}

	// 按房间存储 房间里的所有座位数据
	for roomId, seats := range allSeats {
		key := r.cacheRoomSeatsKey(roomId)
		// seatID : seatJson
		hash := make(map[string]string)
		for _, seat := range seats {
			seatID := seat.DevID
			seatJson, err := json.Marshal(seat)
			if err != nil {
				r.data.log.Errorf("marshal seat error := %v", err)
				return err
			}
			hash[seatID] = string(seatJson)

			// 建立时间序列 zSet
			zKey := r.cacheRoomSeatsTsKey(roomId, seatID)
			var zs []redis.Z
			for _, ts := range seat.Ts {
				startUnix, _ := tool.ParseToUnix(ts.Start)
				endUnix, _ := tool.ParseToUnix(ts.End)
				// 记录每个被占用时间的开始与结束的时间戳
				zs = append(zs, redis.Z{
					// 开始时间
					Score: float64(endUnix),
					// 结束时间
					Member: float64(startUnix),
				})
			}
			if len(zs) > 0 {
				pipe.ZAdd(ctx, zKey, zs...) // 批量插入时间段
				pipe.Expire(ctx, zKey, ttl.AsDuration())
			} else if len(zs) == 0 {
				// 给未被占用的座位一个默认值，使得查询脚本能查询到空闲座位
				def := redis.Z{
					Score:  2300,
					Member: 2300,
				}

				pipe.ZAdd(ctx, zKey, def)
				pipe.Expire(ctx, zKey, ttl.AsDuration())
			}

		}
		// RoomID : {N1111: json1 N2222: json2}
		// 单个房间的座位存储
		pipe.HSet(ctx, key, hash)
		// 设置 TTL , 过时自动删除捏
		pipe.Expire(ctx, key, ttl.AsDuration()).Err()
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		r.data.log.Error("Save SeatInfo in redis ERROR:%s", err.Error())
		return err
	}

	r.data.log.Infof("All seats saved in Redis successfully")
	return nil
}

// RoomSeat 是否保留待商榷
func (r *SeatRepo) GetSeatsByRoom(ctx context.Context, roomID string) ([]*biz.Seat, error) {
	roomKey := r.cacheRoomSeatsKey(roomID)

	data, err := r.data.redis.HGetAll(ctx, roomKey).Result()
	if err != nil {
		r.data.log.Errorf("get seatinfo from redis error (room_id := %s)", roomKey)
		return nil, err
	}

	var seats []*biz.Seat
	for _, v := range data {
		var s biz.Seat
		err = json.Unmarshal([]byte(v), &s)
		if err == nil {
			seats = append(seats, &s)
		}
	}
	return seats, nil
}

// 返回 座位号 座位是否找到 err
func (r *SeatRepo) FindFirstAvailableSeat(ctx context.Context, start, end int64, roomID []string) (string, bool, error) {
	luaScript := `
		local qStart = tonumber(ARGV[1])
		local qEnd = tonumber(ARGV[2])

		-- 收集房间ID
		local roomIDs = {}
		for i = 3, #ARGV do
			table.insert(roomIDs, ARGV[i])
		end

		-- 遍历所有房间ID
		for _, roomID in ipairs(roomIDs) do
			local cursor = "0"
			repeat
				-- 只扫描当前房间下的 seat
				local pattern = "lib:room:" .. roomID .. ":seat:*"
				local scanResult = redis.call("SCAN", cursor, "MATCH", pattern, "COUNT", 100)
				cursor = scanResult[1]
				local keys = scanResult[2]

				for i = 1, #keys do
					local key = keys[i]
					local members = redis.call("ZRANGE", key, 0, -1, "WITHSCORES")

					local free = true
					for j = 2, #members, 2 do
						local startTime = tonumber(members[j - 1])
						local endTime = tonumber(members[j])
						if startTime < qEnd and endTime > qStart then
							free = false
							break
						end
					end

					if free then
						return key -- 找到空闲座位直接返回
					end
				end
			until cursor == "0"
		end

		return nil
	`
	args := make([]interface{}, 0, 2+len(roomID))
	args = append(args, start, end)
	for _, id := range roomID {
		args = append(args, id)
	}

	result, err := r.data.redis.Eval(ctx, luaScript, nil, args...).Result()
	// redis.Nil 来做无匹配座位的表示符，返回 false
	if errors.Is(err, redis.Nil) {
		r.data.log.Infof("No available seat (time:%s)", time.Now().String())
		return "", false, err
	}
	if err != nil {
		r.data.log.Errorf("Error getting first available seat from redis (time:%s)", time.Now().String())
		return "", false, err
	}

	resultStr, ok := result.(string)
	if !ok {
		r.data.log.Infof("No available seat now (time:%s)", time.Now().String())
		return "", false, fmt.Errorf("no available seat now (time:%s)", time.Now().String())
	}

	idx := strings.LastIndexByte(resultStr, ':')
	freeSeatID := resultStr[idx+1:]

	return freeSeatID, true, nil
}

// GetSeatInfos 按楼层查缓存
func (r *SeatRepo) GetSeatInfos(ctx context.Context, stuID string) (map[string][]*biz.Seat, error) {
	// now := time.Now()
	result := make(map[string][]*biz.Seat, len(biz.RoomIDs))

	// 是否有房间命中缓存
	hitAny := false
	// 是否需要后台刷新
	needRefresh := false

	// 循环每个房间
	for _, roomID := range biz.RoomIDs {
		seats, err := r.GetSeatsByRoom(ctx, roomID)
		if err != nil {
			r.data.log.Warnf("get room seats cache(room_id:%s) err: %v", roomID, err)
			needRefresh = true
			continue
		}
		// if !ok {
		// 	needRefresh = true
		// 	continue
		// }

		// 命中缓存
		result[roomID] = seats
		hitAny = true

		// // 判断软过期
		// if ts.IsZero() || now.Sub(ts) > seatsFreshness {
		// 	needRefresh = true
		// }
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
		if err = r.setRoomSeatsCache(ctx, roomID, seats, now); err != nil {
			r.data.log.Warnf("set room seats cache(room_id:%s) err: %v", roomID, err)
		}
	}
	return data, nil
}

