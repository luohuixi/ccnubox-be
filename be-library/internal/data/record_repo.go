package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/asynccnu/ccnubox-be/be-library/internal/biz"
	"github.com/asynccnu/ccnubox-be/be-library/internal/data/DO"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm/clause"
)

const (
	futureRecordKeyPrefix  = "lib:future:records:"
	futureRecordTTL        = 60 * time.Second
	historyRecordKeyPrefix = "lib:history:records:"
	historyRecordTTL       = 60 * time.Second
)

type recordRepo struct {
	data *Data
}

func NewRecordRepo(data *Data) biz.RecordRepo {
	return &recordRepo{
		data: data,
	}
}

// 未来预约缓存
func (r *recordRepo) futureRecordKey(stuID string) string {
	return fmt.Sprintf("%s%s", futureRecordKeyPrefix, stuID)
}

func (r *recordRepo) getFutureRecordsCache(ctx context.Context, stuID string) ([]*biz.FutureRecords, bool, error) {
	key := r.futureRecordKey(stuID)
	val, err := r.data.redis.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var out []*biz.FutureRecords
	if err = json.Unmarshal(val, &out); err != nil {
		return nil, false, err
	}
	return out, true, nil
}

func (r *recordRepo) setFutureRecordsCache(ctx context.Context, stuID string, list []*biz.FutureRecords) error {
	key := r.futureRecordKey(stuID)
	data, err := json.Marshal(list)
	if err != nil {
		return err
	}
	return r.data.redis.Set(ctx, key, data, futureRecordTTL).Err()
}

func (r *recordRepo) delFutureRecordsCache(ctx context.Context, stuID string) error {
	key := r.futureRecordKey(stuID)
	return r.data.redis.Del(ctx, key).Err()
}

// UpsertFutureRecords 复合唯一键去重,写库成功后删除缓存
func (r *recordRepo) UpsertFutureRecords(ctx context.Context, stuID string, list []*biz.FutureRecords) error {
	if len(list) == 0 {
		return nil
	}
	dos := ConvertBizFutureRecordsDO(stuID, list)

	if err := r.data.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "stu_id"},
				{Name: "start"},
				{Name: "end"},
			},
			DoUpdates: clause.AssignmentColumns([]string{
				"remote_id", "owner", "time_desc", "states", "dev_name", "room_id", "room_name", "lab_name",
			}),
		}).
		Create(&dos).Error; err != nil {
		return err
	}
	
	// 写库后删缓存
	if err := r.delFutureRecordsCache(ctx, stuID); err != nil {
		r.data.log.Warnf("del future records cache(stu_id:%s) failed: %v", stuID, err)
	}
	return nil
}

// ListFutureRecords 先读缓存,未命中则查库并写回缓存
func (r *recordRepo) ListFutureRecords(ctx context.Context, stuID string) ([]*biz.FutureRecords, error) {
	if cached, ok, err := r.getFutureRecordsCache(ctx, stuID); err == nil && ok {
		return cached, nil
	} else if err != nil {
		r.data.log.Warnf("get future records cache(stu_id:%s) err: %v", stuID, err)
	}

	var dos []DO.FutureRecord
	if err := r.data.db.WithContext(ctx).
		Where("stu_id = ?", stuID).
		Order("start DESC").
		Find(&dos).Error; err != nil {
		return nil, err
	}

	out := ConvertDOFutureRecordsBiz(dos)

	// 回填缓存
	if err := r.setFutureRecordsCache(ctx, stuID, out); err != nil {
		r.data.log.Warnf("set future records cache(stu_id:%s) err: %v", stuID, err)
	}
	return out, nil
}

// 历史预约缓存
func (r *recordRepo) historyRecordKey(stuID string) string {
	return fmt.Sprintf("%s%s", historyRecordKeyPrefix, stuID)
}

func (r *recordRepo) getHistoryRecordCache(ctx context.Context, stuID string) ([]*biz.HistoryRecords, bool, error) {
	key := r.historyRecordKey(stuID)
	val, err := r.data.redis.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var out []*biz.HistoryRecords
	if err = json.Unmarshal(val, &out); err != nil {
		return nil, false, err
	}
	return out, true, nil
}

func (r *recordRepo) setHistoryRecordCache(ctx context.Context, stuID string, list []*biz.HistoryRecords) error {
	key := r.historyRecordKey(stuID)
	data, err := json.Marshal(list)
	if err != nil {
		return err
	}
	return r.data.redis.Set(ctx, key, data, historyRecordTTL).Err()
}

func (r *recordRepo) delHistoryRecordCache(ctx context.Context, stuID string) error {
	key := r.historyRecordKey(stuID)
	return r.data.redis.Del(ctx, key).Err()
}

func (r *recordRepo) UpsertHistoryRecords(ctx context.Context, stuID string, list []*biz.HistoryRecords) error {
	if len(list) == 0 {
		return nil
	}
	dos := ConvertBizHistoryRecordsDO(stuID, list)

	if err := r.data.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "stu_id"},
				{Name: "submit_time"},
			},
			DoUpdates: clause.AssignmentColumns([]string{
				"place", "floor", "status", "date",
			}),
		}).Create(&dos).Error; err != nil {
		return err
	}
	// 写库后删缓存
	if err := r.delHistoryRecordCache(ctx, stuID); err != nil {
		r.data.log.Warnf("del history record cache(stu_id:%s) err: %v", stuID, err)
	}
	return nil
}

// ListHistoryRecords 先读缓存,未命中则查库并写回缓存
func (r *recordRepo) ListHistoryRecords(ctx context.Context, stuID string) ([]*biz.HistoryRecords, error) {
	if cached, ok, err := r.getHistoryRecordCache(ctx, stuID); err == nil && ok {
		return cached, nil
	} else if err != nil {
		r.data.log.Warnf("get history record cache(stu_id:%s) err: %v", stuID, err)
	}

	var dos []DO.HistoryRecord
	if err := r.data.db.WithContext(ctx).
		Where("stu_id = ?", stuID).
		Order("submit_time DESC").
		Find(&dos).Error; err != nil {
		return nil, err
	}

	out := ConvertDOHistoryRecordsBiz(dos)

	// 回填缓存
	if err := r.setHistoryRecordCache(ctx, stuID, out); err != nil {
		return nil, err
	}

	return out, nil
}
