package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"

	ccnuv1 "github.com/asynccnu/ccnubox-be/be-api/gen/proto/ccnu/v1"
	"github.com/asynccnu/ccnubox-be/be-ccnu/pkg/errorx"
	"github.com/asynccnu/ccnubox-be/be-ccnu/tool"
)

// 定义错误,这里将kratos的error作为一个重要部分传入,此处的错误并不直接在service中去捕获,而是选择在更底层的爬虫去捕获,因为爬虫的错误处理非常复杂
var (
	CCNUSERVER_ERROR = func(err error) error {
		return errorx.New(ccnuv1.ErrorCcnuserverError("ccnu服务器错误"), "ccnuServer", err)
	}

	Invalid_SidOrPwd_ERROR = func(err error) error {
		return errorx.New(ccnuv1.ErrorInvalidSidOrPwd("账号密码错误"), "user", err)
	}

	SYSTEM_ERROR = func(err error) error {
		return errorx.New(ccnuv1.ErrorSystemError("系统内部错误"), "system", err)
	}
)

func (c *ccnuService) GetXKCookie(ctx context.Context, studentId string, password string) (string, error) {

	//初始化client
	client := c.client()

	params, err := tool.Retry(func() (*accountRequestParams, error) {
		return c.makeAccountPreflightRequest(client)
	})
	if err != nil {
		return "", err
	}

	client, err = tool.Retry(func() (*http.Client, error) {
		loginClient, _, err := c.loginClient(ctx, client, studentId, password, params)
		return loginClient, err
	})
	if err != nil {
		return "", err
	}

	client, err = tool.Retry(func() (*http.Client, error) {
		return c.xkLoginClient(client)
	})
	if err != nil {
		return "", err
	}

	cookie, err := tool.Retry(func() (string, error) {
		return c.getBKSCookie(client)
	})
	if err != nil {
		return "", err
	}

	return cookie, nil
}

func (c *ccnuService) GetCCNUCookie(ctx context.Context, studentId string, password string) (string, error) {
	client := c.client()

	params, err := tool.Retry(func() (*accountRequestParams, error) {
		return c.makeAccountPreflightRequest(client)
	})
	if err != nil {
		return "", err
	}

	s, err := tool.Retry(func() (string, error) {
		_, s, err := c.loginClient(ctx, client, studentId, password, params)
		return s, err
	})
	if err != nil {
		return "", err
	}

	return s, nil
}

func (c *ccnuService) GetLibraryCookie(ctx context.Context, studentId, password string) (string, error) {
	client := c.client()

	// 获取登录参数
	params, err := tool.Retry(func() (*accountRequestParams, error) {
		return c.makeAccountPreflightRequest(client)
	})
	if err != nil {
		return "", err
	}

	// 执行登录到图书馆系统
	client, err = tool.Retry(func() (*http.Client, error) {
		return c.libraryLoginClient(ctx, client, studentId, password, params)
	})
	if err != nil {
		return "", err
	}

	// 获取图书馆Cookie
	cookie, err := tool.Retry(func() (string, error) {
		return c.getLibraryCookie(client)
	})
	if err != nil {
		return "", err
	}

	return cookie, nil
}

// 图书馆登录客户端
func (c *ccnuService) libraryLoginClient(ctx context.Context, client *http.Client, studentId string, password string, params *accountRequestParams) (*http.Client, error) {
	v := url.Values{}
	v.Set("username", studentId)
	v.Set("password", password)
	v.Set("lt", params.lt)
	v.Set("execution", params.execution)
	v.Set("_eventId", params._eventId)
	v.Set("submit", params.submit)

	// 登录到图书馆系统
	request, err := http.NewRequest("POST", "https://account.ccnu.edu.cn/cas/login;jsessionid="+params.JSESSIONID+"?service=http://kjyy.ccnu.edu.cn/loginall.aspx?page=", strings.NewReader(v.Encode()))
	if err != nil {
		return nil, SYSTEM_ERROR(err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.109 Safari/537.36")
	request.WithContext(ctx)

	// 发送请求
	resp, err := client.Do(request)
	if err != nil {
		return nil, CCNUSERVER_ERROR(err)
	}
	defer resp.Body.Close()

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, CCNUSERVER_ERROR(err)
	}

	// 检查登录是否成功
	if strings.Contains(string(res), "您输入的用户名或密码有误") {
		return nil, Invalid_SidOrPwd_ERROR(errors.New("学号或密码错误"))
	}

	// 如果没有设置Cookie，说明系统有问题
	if len(resp.Header.Get("Set-Cookie")) == 0 {
		return nil, CCNUSERVER_ERROR(errors.New("登录失败，未获取到Cookie"))
	}

	libraryReq, err := http.NewRequest("GET", "http://kjyy.ccnu.edu.cn/clientweb/default.aspx", nil)
	if err != nil {
		return nil, SYSTEM_ERROR(err)
	}

	libraryReq.WithContext(ctx)

	libraryResp, err := client.Do(libraryReq)
	if err != nil {
		return nil, CCNUSERVER_ERROR(err)
	}
	defer libraryResp.Body.Close()

	if libraryResp.StatusCode != http.StatusOK {
		return nil, CCNUSERVER_ERROR(fmt.Errorf("访问图书馆系统失败，状态码: %d", libraryResp.StatusCode))
	}

	return client, nil
}

// 获取图书馆Cookie
func (c *ccnuService) getLibraryCookie(client *http.Client) (string, error) {
	// 构造用于获取Cookie的URL
	libraryURL := "http://kjyy.ccnu.edu.cn/"
	u, err := url.Parse(libraryURL)
	if err != nil {
		return "", SYSTEM_ERROR(err)
	}

	// 从CookieJar中获取这个域名的Cookie
	cookies := client.Jar.Cookies(u)
	if len(cookies) == 0 {
		return "", SYSTEM_ERROR(errors.New("no cookies found after login"))
	}

	// 拼接所有Cookie为字符串格式
	var cookieStrs []string
	for _, cookie := range cookies {
		cookieStrs = append(cookieStrs, fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
	}

	// 返回拼接好的Cookie字符串
	finalCookie := strings.Join(cookieStrs, "; ")
	return finalCookie, nil
}

// getBKSCookie 返回本科生院的Cookie
func (c *ccnuService) getBKSCookie(client *http.Client) (string, error) {

	// 构造用于获取 Cookie 的 URL（即请求时的登录 URL）
	loginURL := "http://kjyy.ccnu.edu.cn/"
	u, err := url.Parse(loginURL)
	if err != nil {
		return "", SYSTEM_ERROR(err)
	}

	// 从 CookieJar 中获取这个域名的 Cookie
	cookies := client.Jar.Cookies(u)
	if len(cookies) == 0 {
		return "", SYSTEM_ERROR(errors.New("no cookies found after login"))
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

// 2.xkLoginClient 教务系统模拟登录
func (c *ccnuService) xkLoginClient(client *http.Client) (*http.Client, error) {

	//华师本科生院登陆
	request, err := http.NewRequest("GET", "https://account.ccnu.edu.cn/cas/login?service=http%3A%2F%2Fxk.ccnu.edu.cn%2Fsso%2Fpziotlogin", nil)
	if err != nil {
		return nil, CCNUSERVER_ERROR(err)
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.109 Safari/537.36")

	resp, err := client.Do(request)
	if err != nil {
		return nil, CCNUSERVER_ERROR(err)
	}
	defer resp.Body.Close()

	return client, nil
}

// 1.登陆ccnu通行证
func (c *ccnuService) loginClient(ctx context.Context, client *http.Client, studentId string, password string, params *accountRequestParams) (*http.Client, string, error) {

	v := url.Values{}
	v.Set("username", studentId)
	v.Set("password", password)
	v.Set("lt", params.lt)
	v.Set("execution", params.execution)
	v.Set("_eventId", params._eventId)
	v.Set("submit", params.submit)

	request, err := http.NewRequest("POST", "https://account.ccnu.edu.cn/cas/login;jsessionid="+params.JSESSIONID, strings.NewReader(v.Encode()))
	if err != nil {
		return nil, "", SYSTEM_ERROR(err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.109 Safari/537.36")
	request.WithContext(ctx)
	// 创建一个带jar的客户端
	j, _ := cookiejar.New(&cookiejar.Options{})
	client.Jar = j
	//发送请求
	resp, err := client.Do(request)
	if err != nil {
		return nil, "", CCNUSERVER_ERROR(err)
	}
	defer resp.Body.Close()

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", CCNUSERVER_ERROR(err)
	}
	t := time.Now()
	////如果捕获到关键字说明是账号密码错误
	if strings.Contains(string(res), "您输入的用户名或密码有误") {
		return nil, "", Invalid_SidOrPwd_ERROR(errors.New("学号或密码错误"))
	}
	fmt.Println(time.Now().Sub(t))
	// 如果没有捕获到账号密码错误但是也没设置设置Cookie,说明是你师系统有问题
	if len(resp.Header.Get("Set-Cookie")) == 0 {
		return nil, "", CCNUSERVER_ERROR(err)
	}

	//获取 Cookie 中的 CASTGC，这是用于单点登录的凭证
	var CASTGC string
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "CASTGC" {
			CASTGC = cookie.Value
		}
	}
	if CASTGC == "" {
		return client, "", CCNUSERVER_ERROR(err)
	}

	return client, CASTGC, nil
}

type accountRequestParams struct {
	lt         string
	execution  string
	_eventId   string
	submit     string
	JSESSIONID string
}

// 0.前置请求,从html中提取相关参数
func (c *ccnuService) makeAccountPreflightRequest(client *http.Client) (*accountRequestParams, error) {
	var JSESSIONID string
	var lt string
	var execution string
	var _eventId string

	params := &accountRequestParams{}

	// 初始化 http request
	request, err := http.NewRequest("GET", "https://account.ccnu.edu.cn/cas/login", nil)
	if err != nil {
		return params, SYSTEM_ERROR(err)
	}

	// 发起请求
	resp, err := client.Do(request)
	if err != nil {
		return params, SYSTEM_ERROR(err)
	}
	defer resp.Body.Close()

	// 读取 MsgContent
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return params, SYSTEM_ERROR(err)
	}

	// 获取 Cookie 中的 JSESSIONID
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "JSESSIONID" {
			JSESSIONID = cookie.Value
		}
	}

	if JSESSIONID == "" {
		return params, SYSTEM_ERROR(errors.New("Can not get JSESSIONID"))
	}

	// 正则匹配 HTML 返回的表单字段
	ltReg := regexp.MustCompile("name=\"lt\".+value=\"(.+)\"")
	executionReg := regexp.MustCompile("name=\"execution\".+value=\"(.+)\"")
	_eventIdReg := regexp.MustCompile("name=\"_eventId\".+value=\"(.+)\"")

	bodyStr := string(body)

	ltArr := ltReg.FindStringSubmatch(bodyStr)
	if len(ltArr) != 2 {
		return params, CCNUSERVER_ERROR(errors.New("Can not get lt"))
	}
	lt = ltArr[1]

	execArr := executionReg.FindStringSubmatch(bodyStr)
	if len(execArr) != 2 {
		return params, CCNUSERVER_ERROR(errors.New("Can not get execution"))
	}

	execution = execArr[1]

	_eventIdArr := _eventIdReg.FindStringSubmatch(bodyStr)
	if len(_eventIdArr) != 2 {
		return params, CCNUSERVER_ERROR(errors.New("Can not get _eventId"))
	}
	_eventId = _eventIdArr[1]

	params.lt = lt
	params.execution = execution
	params._eventId = _eventId
	params.submit = "LOGIN"
	params.JSESSIONID = JSESSIONID

	return params, nil
}

// -1.前置工作,用于初始化client
func (c *ccnuService) client() *http.Client {
	j, _ := cookiejar.New(&cookiejar.Options{})
	return &http.Client{
		Transport: nil,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
		Jar:     j,
		Timeout: c.timeout,
	}
}
