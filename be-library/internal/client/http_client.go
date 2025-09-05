package client

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"
)

// CookieClient 封装带cookiejar的HTTP客户端
type CookieClient struct {
	client *http.Client
}

// NewCookieClient 创建带cookiejar的HTTP客户端
func NewCookieClient(cookieString string) (*CookieClient, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
			DisableKeepAlives:   false,
		},
		Timeout: 30 * time.Second,
	}

	cc := &CookieClient{
		client: client,
	}

	// 解析并设置初始cookie到华师图书馆域名
	if err = cc.setCookiesFromString(cookieString); err != nil {
		return nil, err
	}

	return cc, nil
}

// setCookiesFromString 解析cookie字符串并设置到jar
func (cc *CookieClient) setCookiesFromString(cookieString string) error {
	if cookieString == "" {
		return nil
	}

	// 华师图书馆的基础URL
	baseURL, err := url.Parse("http://kjyy.ccnu.edu.cn")
	if err != nil {
		return err
	}

	var cookies []*http.Cookie
	for _, part := range strings.Split(cookieString, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if idx := strings.Index(part, "="); idx > 0 {
			name := strings.TrimSpace(part[:idx])
			value := strings.TrimSpace(part[idx+1:])
			cookies = append(cookies, &http.Cookie{
				Name:  name,
				Value: value,
			})
		}
	}

	cc.client.Jar.SetCookies(baseURL, cookies)
	return nil
}

// DoWithContext 执行带上下文的HTTP请求
func (cc *CookieClient) DoWithContext(ctx context.Context, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)

	// 设置标准请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	return cc.client.Do(req)
}

// GetCookies 获取当前的cookies
func (cc *CookieClient) GetCookies() []*http.Cookie {
	baseURL, _ := url.Parse("http://kjyy.ccnu.edu.cn")
	return cc.client.Jar.Cookies(baseURL)
}

// CookiePool 管理CookieClient实例池
type CookiePool struct {
	pool   sync.Map // map[string]*CookieClient
	expiry time.Duration
}

// NewCookiePool 创建新的CookiePool
func NewCookiePool(expiry time.Duration) *CookiePool {
	pool := &CookiePool{
		expiry: expiry,
	}

	// 启动清理goroutine
	go pool.cleanup()
	return pool
}

// GetClient 获取或创建CookieClient
func (cp *CookiePool) GetClient(cookieString string) (*CookieClient, error) {
	if client, ok := cp.pool.Load(cookieString); ok {
		return client.(*CookieClient), nil
	}

	client, err := NewCookieClient(cookieString)
	if err != nil {
		return nil, err
	}

	cp.pool.Store(cookieString, client)
	return client, nil
}

// cleanup 定期清理过期的客户端
func (cp *CookiePool) cleanup() {
	ticker := time.NewTicker(cp.expiry)
	defer ticker.Stop()

	for range ticker.C {
		cp.pool.Range(func(key, value interface{}) bool {
			cp.pool.Delete(key)
			return true
		})
	}
}
