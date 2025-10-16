package crawler

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/asynccnu/ccnubox-be/be-grade/repository/model"
)

const PG_URL = "https://bkzhjw.ccnu.edu.cn/"

func NewCrawlerClientWithCookieJar(t time.Duration, jar *cookiejar.Jar) *http.Client {
    client := &http.Client{
        Transport: nil,
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            return nil
        },
        Timeout: t,
    }
    if jar != nil {
        client.Jar = jar
    }
    return client
}

func NewJarWithCookie(targetURL, rawCookie string) *cookiejar.Jar {
	jar, _ := cookiejar.New(&cookiejar.Options{})
	// 设置目标域名
	u, err := url.Parse(targetURL)
	if err != nil {
		return nil
	}

	// 将字符串形式 Cookie 解析成 []*http.Cookie
	cookies := parseRawCookieString(rawCookie)
	jar.SetCookies(u, cookies)
	return jar
}

func parseRawCookieString(raw string) []*http.Cookie {
	parts := strings.Split(raw, ";")
	var cookies []*http.Cookie
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if kv := strings.SplitN(part, "=", 2); len(kv) == 2 {
			cookies = append(cookies, &http.Cookie{
				Name:  strings.TrimSpace(kv[0]),
				Value: strings.TrimSpace(kv[1]),
			})
		}
	}
	return cookies
}

func ConvertGraduateGrade(graduateGrade []GraduatePoints) []model.Grade {
	var grades []model.Grade
	for _, p := range graduateGrade {
		var xqm int64
		switch p.Xqm {
		case "3":
			xqm = 1
		case "12":
			xqm = 2
		case "16":
			xqm = 3
		}
		grades = append(grades, model.Grade{
			StudentId: p.Xh,
			JxbId:     p.JxbID,
			Kcmc:      p.Kcmc,
			Xnm:       parseInt64(p.Xnm),
			Xqm:       xqm,
			Xf:        parseFloat32(p.Xf),
			Kcxzmc:    p.Kcxzmc,
			Kclbmc:    p.Kclbmc,
			Kcbj:      p.Kcbj,
			Jd:        parseFloat32(p.Jd),
			Cj:        parseFloat32(p.Cj),
		})
	}
	return grades
}

// parseInt64 辅助函数，将字符串转换为 int64
func parseInt64(value string) int64 {
	if i, err := strconv.Atoi(value); err == nil {
		return int64(i)
	}
	return 0
}

// parseFloat32 辅助函数，将字符串转换为 float32
func parseFloat32(value string) float32 {
	if i, err := strconv.ParseFloat(value, 32); err == nil {
		return float32(i)
	}
	return 0
}

func ConvertUndergraduate(undergraduate []Grade) []model.Grade {
	var grades []model.Grade
	for _, g := range undergraduate {
		xnm, xqm := parseXnxq(g.XNXQID)
		mg := model.Grade{
			StudentId: g.XS0101ID, // 学生号
			KcId:      g.KCH,      // 课程ID
			JxbId:     g.JX0404ID, // 教学班ID
			Kcmc:      g.KCMC,     // 课程名称
			Xnm:       xnm,        // 学年
			Xqm:       xqm,        // 学期
			Xf:        g.XF,       // 学分
			Kcxzmc:    g.KCXZMC,   // 课程性质
			Kclbmc:    g.KCSX,     // 课程属性
			Kcbj:      "",         // 无对应字段
			Jd:        0,
			Cj:        g.ZCJ, // 总成绩
			// TODO平时成绩和期末成绩
		}
		grades = append(grades, mg)
	}
	return grades
}

func parseXnxq(xnxq string) (int64, int64) {
	parts := strings.Split(xnxq, "-")
	if len(parts) < 3 {
		return 0, 0
	}
	xnm, _ := strconv.ParseInt(parts[0], 10, 64)
	xqm, _ := strconv.ParseInt(parts[2], 10, 64)
	return xnm, xqm
}
