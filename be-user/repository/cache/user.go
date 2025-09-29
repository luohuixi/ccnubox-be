package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrKeyNotExists = redis.Nil

type UserCache interface {
	GetCookie(ctx context.Context, sid string) (string, error)
	SetCookie(ctx context.Context, sid string, cookie string) error
	GetLibraryCookie(ctx context.Context, sid string) (string, error)
	SetLibraryCookie(ctx context.Context, sid string, cookie string) error
}

type RedisUserCache struct {
	cmd redis.Cmdable
}

// GetCookie 从 Redis 获取指定 sid 对应的 cookie。
func (cache *RedisUserCache) GetCookie(ctx context.Context, sid string) (string, error) {
	// 生成缓存键
	key := cache.key(sid)
	// 获取缓存
	val, err := cache.cmd.Get(ctx, key).Result()
	if err == redis.Nil {
		// 如果缓存未命中，返回一个特定的错误
		return "", ErrKeyNotExists
	} else if err != nil {
		// 其他 Redis 错误
		return "", fmt.Errorf("failed to get value from Redis: %w", err)
	}
	return val, nil
}

// SetCookie 将 sid 和对应的 cookie 存入 Redis。
func (cache *RedisUserCache) SetCookie(ctx context.Context, sid string, cookie string) error {
	// 生成缓存键
	key := cache.key(sid)
	// 设置缓存，过期时间 5分钟 ,学校的cookie过期时间是随着访问量的变化而变化的,做一个简单的单例模式
	err := cache.cmd.Set(ctx, key, cookie, time.Minute*5).Err()
	if err != nil {
		// Redis 设置缓存失败
		return fmt.Errorf("failed to set value in Redis: %w", err)
	}
	return nil
}

// key 生成 Redis 缓存键，格式为 "ccnubox:users:{sid}"。
func (cache *RedisUserCache) key(sid string) string {
	return fmt.Sprintf("ccnubox:users:%s", sid)
}

// GetLibraryCookie 从 Redis 获取指定 sid 对应的图书馆 cookie
func (cache *RedisUserCache) GetLibraryCookie(ctx context.Context, sid string) (string, error) {
	// 生成缓存键
	key := cache.libraryKey(sid)
	// 获取缓存
	val, err := cache.cmd.Get(ctx, key).Result()
	if err == redis.Nil {
		// 如果缓存未命中，返回一个特定的错误
		return "", ErrKeyNotExists
	} else if err != nil {
		// 其他 Redis 错误
		return "", fmt.Errorf("failed to get library cookie from Redis: %w", err)
	}
	return val, nil
}

// SetLibraryCookie 将 sid 和对应的图书馆 cookie 存入 Redis
func (cache *RedisUserCache) SetLibraryCookie(ctx context.Context, sid string, cookie string) error {
	// 生成缓存键
	key := cache.libraryKey(sid)
	// 设置缓存，过期时间 5分钟
	err := cache.cmd.Set(ctx, key, cookie, time.Minute*5).Err()
	if err != nil {
		// Redis 设置缓存失败
		return fmt.Errorf("failed to set library cookie in Redis: %w", err)
	}
	return nil
}

// libraryKey 生成图书馆 Redis 缓存键，格式为 "ccnubox:library:{sid}"
func (cache *RedisUserCache) libraryKey(sid string) string {
	return fmt.Sprintf("ccnubox:library:%s", sid)
}

// NewRedisUserCache 创建一个新的 RedisUserCache 实例
func NewRedisUserCache(cmd redis.Cmdable) UserCache {
	return &RedisUserCache{cmd: cmd}
}
