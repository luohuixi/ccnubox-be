package swag

import (
	"os"

	"github.com/asynccnu/ccnubox-be/bff/errs"
	"github.com/asynccnu/ccnubox-be/bff/pkg/ginx"
	"github.com/asynccnu/ccnubox-be/bff/web"
	"github.com/gin-gonic/gin"
)

type SwagHandler struct {
}

func NewSwagHandler() *SwagHandler {
	return &SwagHandler{
	}
}

func (c *SwagHandler) RegisterRoutes(s *gin.RouterGroup, basicAuthMiddleware gin.HandlerFunc) {
	s.GET("/swag", basicAuthMiddleware, ginx.Wrap(c.GetOpenApi3))
}

// GetOpenApi3 直接返回 YAML 原文
func (c *SwagHandler) GetOpenApi3(ctx *gin.Context) (web.Response, error) {
	filepath := "docs/openapi3.yaml"
	content, err := os.ReadFile(filepath)
	if err != nil {
		return web.Response{}, errs.OPEN_SWAG_ERROR(err)
	}

	// 返回 YAML 字符串
	ctx.String(200, string(content))
	// 为保证返回的文件纯净性，不打印通用响应体
	ctx.Abort()
	return web.Response{}, nil
}
