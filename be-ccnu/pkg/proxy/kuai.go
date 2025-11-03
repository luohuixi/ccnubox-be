package proxy

import (
	"fmt"
	"github.com/spf13/viper"
	"golang.org/x/net/proxy"
	"net/http"
	"os"
)

func NewKuaiHTTPClient() *http.Client {
	// 用户名密码认证(私密代理/独享代理)
	var cfg struct {
		UserName string `json:"username"`
		Password string `json:"password"`
		Proxy    string `json:"proxy"`
	}
	err := viper.UnmarshalKey("kuai", &cfg)
	if err != nil {
		panic(err)
	}

	auth := proxy.Auth{
		User:     cfg.UserName,
		Password: cfg.Password,
	}

	proxy_str := cfg.Proxy

	// 设置代理
	dialer, err := proxy.SOCKS5("tcp", proxy_str, &auth, proxy.Direct)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// 请求目标网页
	client := &http.Client{
		Transport: &http.Transport{Dial: dialer.Dial},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}
	return client
}
