package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type BasicAuthMiddleware struct {
	Username string
	Password string
}

func NewBasicAuthMiddleware() *BasicAuthMiddleware {
	type Config struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	}

	var cfg Config

	// 从配置文件中读取 JWT 配置，并将其解码到 Config 结构体中
	err := viper.UnmarshalKey("basicAuth", &cfg)
	if err != nil {
		// 如果读取配置失败，则抛出一个 panic 错误
		panic(err)
	}
	return &BasicAuthMiddleware{
		Username: cfg.Username,
		Password: cfg.Password,
	}
}

func (m *BasicAuthMiddleware) MiddlewareFunc() gin.HandlerFunc {
	return gin.BasicAuth(gin.Accounts{m.Username: m.Password})
}
