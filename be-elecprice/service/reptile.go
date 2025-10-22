package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// 通用 HTTP 请求函数
func sendRequest(ctx context.Context, url string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("User-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36 Edg/128.0.0.0")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("服务器返回错误状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应体失败: %w", err)
	}

	return string(body), nil
}

// 匹配正则工具
func matchRegex(input, pattern string) (map[string]string, error) {
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(input, -1)
	if matches == nil {
		return nil, errors.New("未匹配到结果")
	}
	res := make(map[string]string)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		//"xx":"123"
		res[match[1]] = match[2]
	}
	return res, nil
}

func matchRegexpOneEle(input, pattern string) (string, error) {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(input)
	if matches == nil {
		return "", errors.New("未匹配到结果")
	}
	if len(matches) < 2 {
		return "", errors.New("未匹配到结果")
	}
	return matches[1], nil
}

func filter(m map[string]string) map[string]string {
	res := make(map[string]string, len(m))
	for k, v := range m {
		if isBlackListed(v) || isEmpty(v) || isEqual(v) {
			continue
		}
		v = formatRoomInfo(v)
		res[k] = v
	}
	return res
}

// formatRoomInfo 格式化房间信息
func formatRoomInfo(name string) string {
	return trimSuffixAndPrefix(replaceAlias(removeExcessiveWord(name)))
}

// removeExcessiveWord 去除中间的多余词汇
func removeExcessiveWord(name string) string {
	for _, v := range RemoveItems {
		name = strings.ReplaceAll(name, v, "")
	}
	return name
}

// trim 去除前后缀
func trimSuffixAndPrefix(name string) string {
	for _, item := range TrimPrefixItems {
		name = strings.TrimPrefix(name, item)
	}
	for _, item := range TrimSuffixItems {
		name = strings.TrimSuffix(name, item)
	}
	return name
}

// replaceAlias 替换别名, 尽可能统一名称
func replaceAlias(name string) string {
	for k, v := range ReplaceItems {
		name = strings.ReplaceAll(name, k, v)
	}
	return name
}

// isEqual 这里面是一些意义不明的房间
func isEqual(name string) bool {
	_, ok := EqualFold[name]
	return ok
}

// isEmpty 排除 xxx空, 但保留 xxx空调
func isEmpty(name string) bool {
	return strings.Contains(name, "空") && !strings.Contains(name, "空调")
}

func isBlackListed(name string) bool {
	for _, b := range BlackList {
		if strings.Contains(name, b) {
			return true
		}
	}
	return false
}
