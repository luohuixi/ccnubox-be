package crawler

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	CASURL_LIBRARY = loginCCNUPassPortURL + "?service=http://kjyy.ccnu.edu.cn/loginall.aspx?page="
	PG_URL_LIBRARY = "http://kjyy.ccnu.edu.cn/"
)

type Library struct {
	Client *http.Client
}

func NewLibrary(client *http.Client) *Library {
	return &Library{
		Client: client,
	}
}

// 1.LoginLibrary 图书馆系统登录系统
func (c *Library) LoginLibrary(ctx context.Context) error {
	request, err := http.NewRequestWithContext(ctx, "POST", CASURL_LIBRARY, nil)
	if err != nil {
		return err
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.109 Safari/537.36")

	resp, err := c.Client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// 2.GetCookieFromLibrarySystem 从图书馆系统中提取 Cookie
func (c *Library) GetCookieFromLibrarySystem() (string, error) {
	parsedURL, err := url.Parse(PG_URL_LIBRARY)
	if err != nil {
		return "", fmt.Errorf("解析 URL 出错: %v", err)
	}

	cookies := c.Client.Jar.Cookies(parsedURL)
	var cookieStr strings.Builder
	for i, cookie := range cookies {
		cookieStr.WriteString(cookie.Name)
		cookieStr.WriteString("=")
		cookieStr.WriteString(cookie.Value)
		if i != len(cookies)-1 {
			cookieStr.WriteString("; ")
		}
	}

	return cookieStr.String(), nil
}
