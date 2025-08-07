package client

import (
	"time"
	"github.com/google/wire"
)

// NewCookiePoolProvider 创建CookiePool的wire provider
func NewCookiePoolProvider() *CookiePool {
	return NewCookiePool(30 * time.Minute)
}

var ProviderSet = wire.NewSet(
	NewClient, 
	NewCCNUServiceProxy,
	NewCookiePoolProvider, // 新增CookiePool provider
)
