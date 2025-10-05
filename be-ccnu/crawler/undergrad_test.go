package crawler

import (
	"context"
	"testing"
)

// 随便写的,比较随意
func Test_GetCookie(t *testing.T) {
	ug := NewUnderGrad(NewCrawlerClient())
	ctx := context.Background()
	html, err := ug.GetParamsFromHtml(ctx)
	if err != nil {
		return
	}
	err = ug.LoginCCNUPassport(ctx, "", "", html)
	if err != nil {
		return
	}
	err = ug.LoginUnderGradSystem(ctx)
	if err != nil {
		return
	}
	cookie, err := ug.GetCookieFromUnderGradSystem()
	if err != nil {
		return
	}
	t.Log(cookie)
}
