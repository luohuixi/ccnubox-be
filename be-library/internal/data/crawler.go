package data

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/asynccnu/ccnubox-be/be-library/internal/biz"
	"github.com/asynccnu/ccnubox-be/be-library/internal/errcode"
	"github.com/asynccnu/ccnubox-be/be-library/internal/model"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/goccy/go-json"
	"github.com/tidwall/gjson"
)

type Crawler struct {
	log    *log.Helper
	client *http.Client
}

func NewLibraryCrawler(logger log.Logger) biz.LibraryCrawler {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,              // 最大空闲连接
			IdleConnTimeout:     90 * time.Second, // 空闲连接超时
			TLSHandshakeTimeout: 10 * time.Second, // TLS握手超时
			DisableKeepAlives:   false,            // 确保不会意外关闭 Keep-Alive
		},
	}

	return &Crawler{
		log:    log.NewHelper(logger),
		client: client,
	}
}

func (c *Crawler) GetSeatInfos(ctx context.Context, cookie string) (map[string][]*biz.Seat, error) {
	var wg sync.WaitGroup
	results := make(map[string][]*biz.Seat)
	mutex := &sync.Mutex{}

	for _, roomID := range biz.RoomIDs {
		wg.Add(1)
		go func(roomID string) {
			defer wg.Done()
			seats, err := c.getSeatInfos(ctx, cookie, roomID)
			if err != nil {
				fmt.Printf("获取房间 %s 座位失败: %v", roomID, err)
				mutex.Lock()
				results[roomID] = nil
				mutex.Unlock()
				return // todo错误处理
			}
			mutex.Lock()
			results[roomID] = seats
			mutex.Unlock()
		}(roomID)
	}

	wg.Wait()
	return results, nil
}

func (c *Crawler) getSeatInfos(ctx context.Context, cookie string, roomid string) ([]*biz.Seat, error) {
	baseURL := "http://kjyy.ccnu.edu.cn/ClientWeb/pro/ajax/device.aspx"

	date := time.Now().Format("2006-01-02")

	params := url.Values{}
	params.Set("classkind", "8")
	params.Set("room_id", roomid)
	params.Set("date", date)
	params.Set("act", "get_rsv_sta")

	fullURL := baseURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		log.Fatal("创建请求失败:", err)
	}

	req.Header.Set("cookie", cookie)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errcode.ErrCrawler
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 用 gjson 解析
	data := gjson.GetBytes(body, "data")
	if !data.Exists() {
		return nil, nil
	}

	var result []*biz.Seat

	// 遍历每个 seat
	data.ForEach(func(_, item gjson.Result) bool {
		seat := &biz.Seat{
			LabName:  item.Get("labName").String(),
			KindName: item.Get("kindName").String(),
			DevID:    item.Get("devId").String(),
			DevName:  item.Get("devName").String(),
		}

		// 提取时间段 ops
		item.Get("ts").ForEach(func(_, op gjson.Result) bool {
			ts := &biz.TimeSlot{
				Start:  op.Get("start").String(),
				End:    op.Get("end").String(),
				State:  op.Get("state").String(),
				Owner:  op.Get("owner").String(),
				Occupy: op.Get("occupy").Bool(),
			}
			seat.Ts = append(seat.Ts, ts)
			return true
		})

		result = append(result, seat)
		return true
	})

	return result, nil
}

func (c *Crawler) ReserveSeat(ctx context.Context, cookie string, devid, start, end string) (string, error) {
	baseURL := "http://kjyy.ccnu.edu.cn/ClientWeb/pro/ajax/reserve.aspx"

	params := url.Values{}
	params.Set("dev_id", devid)
	params.Set("start", start)
	params.Set("end", end)
	params.Set("act", "set_resv")

	fullURL := baseURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		log.Fatal("创建请求失败:", err)
	}

	req.Header.Set("cookie", cookie)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", errcode.ErrCrawler
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var ReserveResp model.Response
	if err = json.Unmarshal(body, &ReserveResp); err != nil {
		return "", err
	}

	if ReserveResp.Ret != 1 {
		return "", fmt.Errorf(ReserveResp.Msg)
	}

	return ReserveResp.Msg, nil
}

func (c *Crawler) GetRecord(ctx context.Context, cookie string) ([]*biz.FutureRecords, error) {
	baseURL := "http://kjyy.ccnu.edu.cn/ClientWeb/pro/ajax/reserve.aspx"

	params := url.Values{}
	params.Set("act", "get_my_resv")

	fullURL := baseURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		log.Fatal("创建请求失败:", err)
	}

	req.Header.Set("cookie", cookie)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errcode.ErrCrawler
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 用 gjson 解析
	data := gjson.GetBytes(body, "data")
	if !data.Exists() {
		return nil, nil
	}

	var result []*biz.FutureRecords

	// 遍历每个record
	for _, item := range data.Array() {
		rawStates := item.Get("states").String()
		html := "<div>" + rawStates + "</div>"

		// 使用goquery提取span内的纯文本
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			continue
		}

		var plainStates []string
		doc.Find("span").Each(func(i int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if text != "" {
				plainStates = append(plainStates, text)
			}
		})

		record := &biz.FutureRecords{
			ID:       item.Get("id").String(),
			Owner:    item.Get("owner").String(),
			Start:    item.Get("start").String(),
			End:      item.Get("end").String(),
			TimeDesc: item.Get("timeDesc").String(),
			States:   strings.Join(plainStates, ","),
			DevName:  item.Get("devName").String(),
			RoomID:   item.Get("roomId").String(),
			RoomName: item.Get("roomName").String(),
			LabName:  item.Get("labName").String(),
		}

		result = append(result, record)
	}

	return result, nil
}

func (c *Crawler) GetHistory(ctx context.Context, cookie string) ([]*biz.HistoryRecords, error) {
	fullURL := "http://kjyy.ccnu.edu.cn/clientweb/m/a/resvlist.aspx"

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		log.Fatal("创建请求失败:", err)
	}

	req.Header.Set("cookie", cookie)

	//todo:去重
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errcode.ErrCrawler
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var records []*biz.HistoryRecords

	doc.Find("li.item-content").Each(func(i int, item *goquery.Selection) {
		place := item.Find(".item-title").Text()
		status := item.Find(".item-after").Text()
		date := item.Find(".item-subtitle").Text()
		submitText := item.Find(".item-text").Text()
		submitParts := strings.Split(submitText, ",")
		if len(submitParts) >= 2 {
			floor := submitParts[0]
			floor = strings.TrimSpace(floor)
			submitTime := submitParts[2]
			submitTime = strings.TrimSpace(submitTime)

			records = append(records, &biz.HistoryRecords{
				Place:      place,
				Floor:      floor,
				Status:     status,
				Date:       date,
				SubmitTime: submitTime,
			})
		}
	})

	return records, nil
}

func (c *Crawler) GetCreditPoint(ctx context.Context, cookie string) (*biz.CreditPoints, error) {
	fullURL := "http://kjyy.ccnu.edu.cn/clientweb/m/a/credit.aspx"

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		log.Fatal("创建请求失败:", err)
	}

	req.Header.Set("cookie", cookie)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errcode.ErrCrawler
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var summary *biz.CreditSummary
	doc.Find("table tbody tr").Each(func(i int, s *goquery.Selection) {
		tds := s.Find("td")
		if tds.Length() >= 3 {
			summary = &biz.CreditSummary{
				System: strings.TrimSpace(tds.Eq(0).Text()),
				Remain: strings.TrimSpace(tds.Eq(1).Text()),
				Total:  strings.TrimSpace(tds.Eq(2).Text()),
			}
		}
	})

	var records []*biz.CreditRecord
	doc.Find("#my_resv_list li").Each(func(i int, s *goquery.Selection) {
		record := &biz.CreditRecord{
			Title:    strings.TrimSpace(s.Find(".item-title").Text()),
			Subtitle: strings.TrimSpace(s.Find(".item-subtitle").Text()),
			Location: strings.TrimSpace(s.Find(".item-text").Text()),
		}
		records = append(records, record)
	})

	var result *biz.CreditPoints

	result = &biz.CreditPoints{
		Summary: summary,
		Records: records,
	}

	return result, nil
}

func (c *Crawler) GetDiscussion(ctx context.Context, cookie string, classid, date string) ([]*biz.Discussion, error) {
	baseURL := "http://kjyy.ccnu.edu.cn/ClientWeb/pro/ajax/device.aspx"

	params := url.Values{}
	params.Set("classkind", "1")
	params.Set("class_id", classid)
	params.Set("date", date)
	params.Set("act", "get_rsv_sta")

	fullURL := baseURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		log.Fatal("创建请求失败:", err)
	}

	req.Header.Set("cookie", cookie)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errcode.ErrCrawler
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	data := gjson.GetBytes(body, "data")
	if !data.Exists() {
		return nil, nil
	}

	var result []*biz.Discussion

	data.ForEach(func(_, item gjson.Result) bool {
		dis := &biz.Discussion{
			LabID:    item.Get("labId").String(),
			LabName:  item.Get("labName").String(),
			KindID:   item.Get("kindId").String(),
			KindName: item.Get("kindName").String(),
			DevID:    item.Get("devId").String(),
			DevName:  item.Get("devName").String(),
		}

		item.Get("ts").ForEach(func(_, op gjson.Result) bool {
			ts := &biz.DiscussionTS{
				Start:  op.Get("start").String(),
				End:    op.Get("end").String(),
				State:  op.Get("state").String(),
				Title:  op.Get("title").String(),
				Owner:  op.Get("owner").String(),
				Occupy: op.Get("occupy").Bool(),
			}
			dis.TS = append(dis.TS, ts)
			return true
		})

		result = append(result, dis)
		return true
	})

	return result, nil
}

func (c *Crawler) SearchUser(ctx context.Context, cookie string, studentid string) (*biz.Search, error) {
	baseURL := "http://kjyy.ccnu.edu.cn/ClientWeb/pro/ajax/data/searchAccount.aspx"

	params := url.Values{}
	params.Set("term", studentid)

	fullURL := baseURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		log.Fatal("创建请求失败:", err)
	}

	req.Header.Set("cookie", cookie)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errcode.ErrCrawler
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var search []*biz.Search
	if err := json.Unmarshal(body, &search); err != nil {
		return nil, err
	}

	return search[0], nil
}
func (c *Crawler) ReserveDiscussion(ctx context.Context, cookie string, devid, labid, kindid, title, start, end string, list []string) (string, error) {
	mbList := "$" + strings.Join(list, ",")

	baseURL := "http://kjyy.ccnu.edu.cn/ClientWeb/pro/ajax/reserve.aspx"

	params := url.Values{}
	params.Set("dev_id", devid)
	params.Set("lab_id", labid)
	params.Set("kind_id", kindid)
	params.Set("min_user", "3")
	params.Set("max_user", "4")
	params.Set("test_name", title)
	params.Set("mb_list", mbList)
	params.Set("start", start)
	params.Set("end", end)
	params.Set("act", "set_resv")

	fullURL := baseURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		log.Fatal("创建请求失败:", err)
	}

	req.Header.Set("cookie", cookie)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", errcode.ErrCrawler
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var ReserveResp model.Response
	if err = json.Unmarshal(body, &ReserveResp); err != nil {
		return "", err
	}

	if ReserveResp.Ret != 1 {
		return "", fmt.Errorf(ReserveResp.Msg)
	}

	return ReserveResp.Msg, nil
}

func (c *Crawler) CancelReserve(ctx context.Context, cookie string, id string) (string, error) {
	baseURL := "http://kjyy.ccnu.edu.cn/ClientWeb/pro/ajax/reserve.aspx"

	params := url.Values{}
	params.Set("act", "del_resv")
	params.Set("id", id)

	fullURL := baseURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		log.Fatal("创建请求失败:", err)
	}

	req.Header.Set("cookie", cookie)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", errcode.ErrCrawler
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	jsonRegexp := regexp.MustCompile(`\{[^}]+}`)
	matches := jsonRegexp.FindAll(body, -1)

	var CancelResp model.Response
	for _, m := range matches {
		if err = json.Unmarshal(m, &CancelResp); err != nil {
			continue // 忽略无效块
		}
		if CancelResp.Ret == 1 {
			return CancelResp.Msg, nil
		}
		return "", fmt.Errorf(CancelResp.Msg)
	}

	return CancelResp.Msg, nil
}
