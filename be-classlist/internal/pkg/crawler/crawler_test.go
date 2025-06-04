package crawler

import (
	"context"
	"fmt"
	"github.com/asynccnu/ccnubox-be/be-classlist/test"
	"testing"
	"time"
)

var cookie = "JSESSIONID=9923654784AEF41BD751198E64AD830B"

func TestCrawler_GetClassInfosForUndergraduate(t *testing.T) {
	crawler := NewClassCrawler(test.NewLogger())
	start := time.Now()
	infos, scs, err := crawler.GetClassInfosForUndergraduate(context.Background(), "testID", "2024", "2", cookie)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(fmt.Sprintf("一共耗时%v", time.Since(start)))
	t.Log(infos, scs)
}

func BenchmarkCrawler_GetClassInfosForUndergraduate(b *testing.B) {
	crawler := NewClassCrawler(test.NewLogger())

	ctx := context.Background()

	// 通常第一次调用可以预热缓存等，不纳入统计
	_, _, _ = crawler.GetClassInfosForUndergraduate(ctx, "testID", "2024", "2", cookie)

	b.ResetTimer() // 重置计时器，排除预热时间
	for i := 0; i < b.N; i++ {
		_, _, err := crawler.GetClassInfosForUndergraduate(ctx, "testID", "2024", "2", cookie)
		if err != nil {
			b.Fatal(err)
		}
	}
}
