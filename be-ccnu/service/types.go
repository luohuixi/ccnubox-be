package service

import (
	"context"
	"time"

	"github.com/asynccnu/ccnubox-be/be-ccnu/crawler"
	"github.com/asynccnu/ccnubox-be/be-ccnu/pkg/logger"
)

type CCNUService interface {
	LoginCCNU(ctx context.Context, studentId string, password string) (bool, error)
	GetXKCookie(ctx context.Context, studentId string, password string) (string, error)
	GetLibraryCookie(ctx context.Context, studentId, password string) (string, error)
}

// 这里直接依赖 passport 是不是太过于简单粗暴没做好解耦
type ccnuService struct {
	passport *crawler.Passport
	timeout  time.Duration
	l        logger.Logger
}

func NewCCNUService(l logger.Logger, passport *crawler.Passport) CCNUService {
	return &ccnuService{
		passport: passport,
		timeout:  time.Minute * 2,
		l:        l,
	}
}
