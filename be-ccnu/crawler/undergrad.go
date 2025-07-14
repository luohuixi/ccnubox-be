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
	loginCCNUPassPortURL = "https://account.ccnu.edu.cn/cas/login"
	pgUrl                = "http://xk.ccnu.edu.cn/jwglxt"
)

// 存放本科生院相关的爬虫
type UnderGrad struct{}

func NewUnderGrad() *UnderGrad {
	return &UnderGrad{}
}

// 1.前置请求,从html中提取相关参数
func (c *UnderGrad) GetParamsFromHtml(client *http.Client) (*AccountRequestParams, error) {
	var JSESSIONID string
	var lt string
	var execution string
	var _eventId string

	params := &AccountRequestParams{}

	// 初始化 http request
	request, err := http.NewRequest("GET", loginCCNUPassPortURL, nil)
	if err != nil {
		return params, err
	}

	// 发起请求
	resp, err := client.Do(request)
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
func (c *UnderGrad) LoginCCNUPassport(ctx context.Context, client *http.Client, studentId string, password string, params *AccountRequestParams) (*http.Client, error) {

	v := url.Values{}
	v.Set("username", studentId)
	v.Set("password", password)
	v.Set("lt", params.lt)
	v.Set("execution", params.execution)
	v.Set("_eventId", params._eventId)
	v.Set("submit", params.submit)

	request, err := http.NewRequest("POST", loginCCNUPassPortURL+";jsessionid="+params.JSESSIONID, strings.NewReader(v.Encode()))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.109 Safari/537.36")
	request.WithContext(ctx)

	//发送请求
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	//如果捕获到关键字说明是账号密码错误
	if strings.Contains(string(res), "您输入的用户名或密码有误") {
		return nil, INCorrectPASSWORD
	}

	// 如果没有捕获到账号密码错误但是也没设置设置Cookie,说明是你师系统有问题
	if len(resp.Header.Get("Set-Cookie")) == 0 {
		return nil, err
	}

	////获取 Cookie 中的 CASTGC，这是用于单点登录的凭证
	//var CASTGC string
	//for _, cookie := range resp.Cookies() {
	//	if cookie.Name == "CASTGC" {
	//		CASTGC = cookie.Value
	//	}
	//}
	//if CASTGC == "" {
	//	return client, "", err
	//}

	return client, nil
}

// 3.LoginPUnderGradSystem 教务系统模拟登录
func (c *UnderGrad) LoginUnderGradSystem(client *http.Client) (*http.Client, error) {

	//华师本科生院登陆
	request, err := http.NewRequest("GET", loginCCNUPassPortURL+"?service=http%3A%2F%2Fxk.ccnu.edu.cn%2Fsso%2Fpziotlogin", nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.109 Safari/537.36")

	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return client, nil
}

// 4. GetCookieFromUnderGradSystem 返回本科生院的Cookie
func (c *UnderGrad) GetCookieFromUnderGradSystem(client *http.Client) (string, error) {

	// 构造用于获取 Cookie 的 URL（即请求时的登录 URL）
	u, err := url.Parse(pgUrl)
	if err != nil {
		return "", err
	}

	// 从 CookieJar 中获取这个域名的 Cookie
	cookies := client.Jar.Cookies(u)
	if len(cookies) == 0 {
		return "", errors.New("no cookies found after login")
	}

	// 拼接所有 Cookie 为字符串格式
	var cookieStr string
	for _, cookie := range cookies {
		if cookie.Name == "JSESSIONID" {
			cookieStr = cookie.Value
		}
	}

	// 返回拼接好的 Cookie 字符串
	return fmt.Sprintf("JSESSIONID=%s", cookieStr), nil

}

type AccountRequestParams struct {
	lt         string
	execution  string
	_eventId   string
	submit     string
	JSESSIONID string
}
