package crawler

import (
	"context"
	"github.com/asynccnu/ccnubox-be/be-classlist/test"
	"io/ioutil"
	"testing"
)

func Test_extractCourseInfo(t *testing.T) {

	path := "./test.html"

	// 读取文件内容
	content, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	t.Logf("length: %v", len(content))

	c := NewClassCrawler2(test.NewLogger())

	classes, err := c.extractCourses("2025", "1", string(content))
	if err != nil {
		t.Fatalf("failed to extract classes: %v", err)
	}
	for _, class := range classes {
		t.Log(class)
	}
}

func Test_Crawler2(t *testing.T) {
	c := NewClassCrawler2(test.NewLogger())
	test_cookie := "bzb_jsxsd=3E32EA47C5848F74CDC0DF2FC15F2141; __root_domain_v=.ccnu.edu.cn; _qddaz=QD.813351615258360; SERVERID=jw1; bzb_njw=A7B26A39D263A42418262D2E84DD7EDC; SERVERIDgld=pc1"
	a, b, err := c.GetClassInfosForUndergraduate(context.Background(), "2023214414", "2025", "1", test_cookie)
	if err != nil {
		t.Fatalf("failed to crawl: %v", err)
	}
	for _, info := range a {
		t.Log(info)
	}
	for _, info := range b {
		t.Log(info)
	}

}
