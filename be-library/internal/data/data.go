package data

import (
	"github.com/asynccnu/ccnubox-be/be-library/internal/conf"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewLibraryCrawler, NewDelayKafka)

// Data 做CURD时使用该框架
type Data struct {
	db      *gorm.DB
	log     *log.Helper
	crawler *Crawler
	redis   *redis.Client
}

// NewData .
func NewData(c *conf.Data, logger log.Logger, db *gorm.DB, crawler *Crawler) (*Data, func(), error) {
	data := &Data{
		log:     log.NewHelper(logger),
		db:      db,
		crawler: crawler,
	}

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return data, cleanup, nil
}

// 迁移
func (d *Data) migrate() error {
	d.log.Info("Starting database migration...")

	err := d.db.AutoMigrate(
		&seat{},
	)

	if err != nil {
		return err
	}

	return nil
}
