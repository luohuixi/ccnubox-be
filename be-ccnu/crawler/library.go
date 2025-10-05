package crawler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
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

// 1.前置请求,从html中提取相关参数
func (c *Library) GetParamsFromHtml(ctx context.Context) (*AccountRequestParams, error) {
	var JSESSIONID string
	var lt string
	var execution string
	var _eventId string

	params := &AccountRequestParams{}

	// 初始化 http request
	request, err := http.NewRequestWithContext(ctx, "GET", loginCCNUPassPortURL, nil)
	if err != nil {
		return params, err
	}

	// 发起请求
	resp, err := c.Client.Do(request)
	if err != nil {
		return params, err
	}
	defer resp.Body.Close()

	// 读取 MsgContent
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return params, err
	}

	// 获取 Cookie 中的 JSESSIONID
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "JSESSIONID" {
			JSESSIONID = cookie.Value
		}
	}

	if JSESSIONID == "" {
		return params, errors.New("Can not get JSESSIONID")
	}

	// 正则匹配 HTML 返回的表单字段
	ltReg := regexp.MustCompile("name=\"lt\".+value=\"(.+)\"")
	executionReg := regexp.MustCompile("name=\"execution\".+value=\"(.+)\"")
	_eventIdReg := regexp.MustCompile("name=\"_eventId\".+value=\"(.+)\"")

	bodyStr := string(body)

	ltArr := ltReg.FindStringSubmatch(bodyStr)
	if len(ltArr) != 2 {
		return params, errors.New("Can not get lt")
	}
	lt = ltArr[1]

	execArr := executionReg.FindStringSubmatch(bodyStr)
	if len(execArr) != 2 {
		return params, errors.New("Can not get execution")
	}
	execution = execArr[1]

	_eventIdArr := _eventIdReg.FindStringSubmatch(bodyStr)
	if len(_eventIdArr) != 2 {
		return params, errors.New("Can not get _eventId")
	}
	_eventId = _eventIdArr[1]

	params.lt = lt
	params.execution = execution
	params._eventId = _eventId
	params.submit = "LOGIN"
	params.JSESSIONID = JSESSIONID

	return params, nil
}

// 2.登陆ccnu通行证
func (c *Library) LoginCCNUPassport(ctx context.Context, studentId string, password string, params *AccountRequestParams) error {
	v := url.Values{}
	v.Set("username", studentId)
	v.Set("password", password)
	v.Set("lt", params.lt)
	v.Set("execution", params.execution)
	v.Set("_eventId", params._eventId)
	v.Set("submit", params.submit)

	urlstr := loginCCNUPassPortURL + ";jsessionid=" + params.JSESSIONID
	request, err := http.NewRequestWithContext(ctx, "POST", urlstr, strings.NewReader(v.Encode()))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.109 Safari/537.36")

	resp, err := c.Client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if strings.Contains(string(res), "您输入的用户名或密码有误") {
		return INCorrectPASSWORD
	}

	if len(resp.Header.Get("Set-Cookie")) == 0 {
		return errors.New("登录失败，未返回 Cookie")
	}

	return nil
}

// 3.LoginLibrary 图书馆系统登录系统
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

// 4.GetCookieFromLibrarySystem 从图书馆系统中提取 Cookie
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
