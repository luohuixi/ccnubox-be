package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/asynccnu/ccnubox-be/be-library/internal/biz"
	"github.com/asynccnu/ccnubox-be/be-library/pkg/tool"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
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
		tsKey := r.cacheRoomTsKey(roomId)
		// 存储每个获取每个房间的时间戳，用于比较软更新时使用
		pipe.HSet(ctx, tsKey, time.Now().UnixMilli())

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

			// 建立时间序列 zSet
			zKey := r.cacheSeatTsKey(seatID)
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
		r.log.Error("Save SeatInfo in redis ERROR:%s", err.Error())
		return err
	}

	r.log.Infof("All seats saved in Redis successfully")
	return nil
}

// RoomSeat 是否保留待商榷
func (r *SeatRepo) GetSeatsByRoom(ctx context.Context, roomID string, stuID string) ([]*biz.Seat, bool, error) {
	roomKey := r.cacheRoomSeatsKey(roomID)

	data, err := r.data.redis.HGetAll(ctx, roomKey).Result()
	if err != nil {
		r.log.Errorf("get seatinfo from redis error (room_id := %s)", roomKey)
		return nil, true, err
	} else if len(data) == 0 {
		r.log.Infof("get no seatinfo from redis (room_id := %s) reloading", roomKey)
		return nil, false, err
	}

	seats := []*biz.Seat{}
	for _, v := range data {
		var s biz.Seat
		err := json.Unmarshal([]byte(v), &s)
		if err == nil {
			seats = append(seats, &s)
		}
	}
	return seats, true, nil
}

// 返回 座位号 座位是否找到 err
func (r *SeatRepo) FindFirstAvailableSeat(ctx context.Context, start, end int64, roomID []string) (string, bool, error) {
	luaScript := `
		local qStart = tonumber(ARGV[1])
		local qEnd = tonumber(ARGV[2])
		local cursor = "0"

		repeat
			local scanResult = redis.call("SCAN", cursor, "MATCH", "seat:*:times", "COUNT", 100)
			cursor = scanResult[1]
			local keys = scanResult[2]

			for i=1,#keys do
				local members = redis.call("ZRANGE", keys[i], 0, -1, "WITHSCORES")
				local free = true
				for j=2,#members,2 do
					local startTime = tonumber(members[j-1])
					local endTime = tonumber(members[j])
					if startTime < qEnd and endTime > qStart then
						free = false
						break
					end
				end
				if free then
					return keys[i]  -- 返回第一个空闲座位 key
				end
			end
		until cursor == "0"

		return nil
	`
	result, err := r.data.redis.Eval(ctx, luaScript, nil, start, end).Result()
	// redis.Nil 来做无匹配座位的表示符，返回 false
	if err == redis.Nil {
		r.log.Infof("No available seat (time:%s)", time.Now().String())
		return "", false, err
	}
	if err != nil {
		r.log.Errorf("Error getting first available seat from redis (time:%s)", time.Now().String())
		return "", false, err
	}

	resultStr, ok := result.(string)
	if !ok {
		r.log.Infof("No available seat now (time:%s)", time.Now().String())
		return "", false, fmt.Errorf("no available seat now (time:%s)", time.Now().String())
	}
	return resultStr, true, nil
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
		seats, ok, err := r.GetSeatsByRoom(ctx, roomID, stuID)
		if err != nil {
			r.log.Warnf("get room seats cache(room_id:%s) err: 	 %v", roomID, err)
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
