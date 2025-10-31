package crawler

import (
	"errors"
	"github.com/asynccnu/ccnubox-be/be-ccnu/pkg/proxy"
	"net/http"
	"net/http/cookiejar"
	"time"
)

var (
	INCorrectPASSWORD = errors.New("账号密码错误")
)

func NewCrawlerClient(t time.Duration) *http.Client {
	j, _ := cookiejar.New(&cookiejar.Options{})
	// 未配置代理时使用默认client
	client := proxy.NewShenLongHTTPClient()
	client.Jar = j
	client.Timeout = t
	return client
}
