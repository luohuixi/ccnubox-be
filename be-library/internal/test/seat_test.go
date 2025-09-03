package test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/asynccnu/ccnubox-be/be-library/internal/biz"
	"github.com/asynccnu/ccnubox-be/be-library/internal/client"
	"github.com/asynccnu/ccnubox-be/be-library/internal/conf"
	"github.com/asynccnu/ccnubox-be/be-library/internal/crawler"
	"github.com/asynccnu/ccnubox-be/be-library/internal/data"
	"github.com/asynccnu/ccnubox-be/be-library/internal/registry"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
)

var confPath string

func init() {
	flag.StringVar(&confPath, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

// 全局 repo
var repo *data.SeatRepo
var bizz biz.LibraryBiz

// TestMain 在所有测试前初始化依赖
func TestMain(m *testing.M) {
	flag.Parse()

	c := config.New(
		config.WithSource(file.NewSource(confPath)),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	// 初始化 DB + Redis
	db, err := data.NewDB(bc.Data)
	if err != nil {
		panic(err)
	}
	rdb := data.NewRedisDB(bc.Data, log.NewStdLogger(os.Stdout))

	confServer := bc.Server
	confRegistry := bc.Registry
	logger := log.NewStdLogger(os.Stdout)

	cookiePool := client.NewCookiePoolProvider()
	etcdRegistry := registry.NewRegistrarServer(confRegistry, logger)
	userServiceClient, err := client.NewClient(etcdRegistry, confRegistry, logger)
	if err != nil {
		panic(err)
	}
	ccnuServiceProxy := client.NewCCNUServiceProxy(userServiceClient)
	duration := biz.NewWaitTime(confServer)
	libraryCrawler := crawler.NewLibraryCrawler(logger, cookiePool, ccnuServiceProxy, duration)

	d, err := data.NewData(bc.Data, log.NewStdLogger(os.Stdout), db, rdb)
	if err != nil {
		panic(err)
	}

	repo = data.NewSeatRepo(d, log.NewStdLogger(os.Stdout), libraryCrawler).(*data.SeatRepo)
	bizz = biz.NewLibraryBiz(libraryCrawler, log.NewStdLogger(os.Stdout), repo)

	// 执行测试
	m.Run()
}

// 10s -> 8s
func TestSaveRoomSeatsInRedis(t *testing.T) {
	stuID := "2024214744"
	ctx := context.Background()

	err := repo.SaveRoomSeatsInRedis(ctx, stuID)
	if err != nil {
		panic(err)
	}
}

func TestGetSeat(t *testing.T) {
	ctx := context.Background()
	roomID := "101699179"

	seats, err := repo.GetSeatsByRoom(ctx, roomID)
	if err != nil {
		panic(err)
	}

	for _, seat := range seats {
		fmt.Println(seat)
	}
}

func TestFindFirstAvailbleSeat(t *testing.T) {
	ctx := context.Background()
	devid, _, err := repo.FindFirstAvailableSeat(ctx, 2000, 2100)
	if err != nil {
		panic(err)
	}
	fmt.Println(devid)
}

func TestReserveSeatRandomly(t *testing.T) {
	id := "2024214744"
	start := "2025-09-02 20:00"
	end := "2025-09-02 21:00"
	ctx := context.Background()
	msg, err := bizz.ReserveSeatRandomly(ctx, id, start, end)
	if err != nil {
		panic(err)
	}
	fmt.Println(msg)
}
