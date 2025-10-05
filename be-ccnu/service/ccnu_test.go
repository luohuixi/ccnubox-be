package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/asynccnu/ccnubox-be/be-ccnu/crawler"
	"github.com/asynccnu/ccnubox-be/be-ccnu/pkg/logger"
)

type TestLogger struct {
}

func (t *TestLogger) Debug(msg string, args ...logger.Field) {

}

func (t *TestLogger) Info(msg string, args ...logger.Field) {

}

func (t *TestLogger) Warn(msg string, args ...logger.Field) {

}

func (t *TestLogger) Error(msg string, args ...logger.Field) {

}

func Test_ccnuService_getGradCookie(t *testing.T) {
	client := crawler.NewCrawlerClient()
	passport := crawler.NewPassport(client)
	testLogger := new(TestLogger)
	ccs := NewCCNUService(testLogger, passport)
	stuId, password := "", ""
	cookie, err := ccs.GetLibraryCookie(context.Background(), stuId, password)
	if err != nil {
		t.Errorf("GetXKCookie err : %v", err)
	}
	t.Log(cookie)
	fmt.Println(cookie, "如果能正常显示的话")
}
