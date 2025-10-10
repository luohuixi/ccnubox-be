package crawler

import (
	"errors"
	"net/http"
	"net/http/cookiejar"
	"time"
)

var (
	INCorrectPASSWORD = errors.New("账号密码错误")
)

func NewCrawlerClient(t time.Duration) *http.Client {
	j, _ := cookiejar.New(&cookiejar.Options{})
	return &http.Client{
		Transport: nil,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
		Jar:     j,
		Timeout: t,
	}
}
