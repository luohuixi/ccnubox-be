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
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	creditPointsKeyPrefix = "lib:credit:point:"
	creditPointsTTL       = 60 * time.Second
)

type creditPointsRepo struct {
	data *Data
}

func NewCreditPointsRepo(data *Data) biz.CreditPointsRepo {
	return &creditPointsRepo{
		data: data,
	}
}

func (r *creditPointsRepo) creditPointsKey(stuID string) string {
	return fmt.Sprintf("%s%s", creditPointsKeyPrefix, stuID)
}

func (r *creditPointsRepo) getCreditPointsCache(ctx context.Context, stuID string) (*biz.CreditPoints, bool, error) {
	key := r.creditPointsKey(stuID)
	val, err := r.data.redis.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var out *biz.CreditPoints
	if err = json.Unmarshal(val, &out); err != nil {
		return nil, false, err
	}
	return out, true, nil
}

func (r *creditPointsRepo) setCreditPointCache(ctx context.Context, stuID string, list *biz.CreditPoints) error {
	key := r.creditPointsKey(stuID)
	data, err := json.Marshal(list)
	if err != nil {
		return err
	}
	return r.data.redis.Set(ctx, key, data, creditPointsTTL).Err()
}

func (r *creditPointsRepo) delCreditPointCache(ctx context.Context, stuID string) error {
	key := r.creditPointsKey(stuID)
	return r.data.redis.Del(ctx, key).Err()
}

// UpsertCreditPoint 复合唯一键去重,写库成功后删除缓存
func (r *creditPointsRepo) UpsertCreditPoint(ctx context.Context, stuID string, list *biz.CreditPoints) error {
	if list == nil {
		return nil
	}

	sum, recs := ConvertBizCreditPointsDO(stuID, list)
	db := r.data.db.WithContext(ctx)

	// summary：stu_id 唯一，冲突更新 system/remain/total
	if sum != nil {
		if err := db.
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "stu_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"system", "remain", "total"}),
			}).
			Create(sum).Error; err != nil {
			return err
		}
	}

	// records：按 stu_id+title+subtitle+location 去重，冲突忽略
	if len(recs) > 0 {
		if err := db.
			Clauses(clause.OnConflict{
				Columns: []clause.Column{
					{Name: "stu_id"},
					{Name: "title"},
					{Name: "subtitle"},
					{Name: "location"},
				},
				DoNothing: true,
			}).
			Create(&recs).Error; err != nil {
			return err
		}
	}

	// 写库后删缓存
	if err := r.delCreditPointCache(ctx, stuID); err != nil {
		r.data.log.Warnf("del credit point cache(stu_id:%s) failed: %v", stuID, err)
	}
	return nil
}

func (r *creditPointsRepo) ListCreditPoint(ctx context.Context, stuID string) (*biz.CreditPoints, error) {
	if cached, ok, err := r.getCreditPointsCache(ctx, stuID); err == nil && ok {
		return cached, nil
	} else if err != nil {
		r.data.log.Warnf("get credit point cache(stu_id:%s) err: %v", stuID, err)
	}

	db := r.data.db.WithContext(ctx)

	// 读 summary
	var sum DO.CreditSummary
	var sumPtr *DO.CreditSummary
	if err := db.Where("stu_id = ?", stuID).First(&sum).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		// not found: sumPtr 保持为 nil
	} else {
		sumPtr = &sum
	}

	// 读 records
	var recs []DO.CreditRecord
	if err := db.Where("stu_id = ?", stuID).Find(&recs).Error; err != nil {
		return nil, err
	}

	out := ConvertDOCreditPointsBiz(sumPtr, recs)

	// 回填缓存
	if err := r.setCreditPointCache(ctx, stuID, out); err != nil {
		r.data.log.Warnf("set credit point cache(stu_id:%s) err: %v", stuID, err)
	}
	return out, nil
}
