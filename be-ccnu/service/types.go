package service

import (
	"context"
	"time"

	"github.com/asynccnu/ccnubox-be/be-ccnu/pkg/logger"
)

type CCNUService interface {
	GetCCNUCookie(ctx context.Context, studentId string, password string) (string, error)
	GetXKCookie(ctx context.Context, studentId string, password string) (string, error)
	GetLibraryCookie(ctx context.Context, studentId, password string) (string, error)
}

type ccnuService struct {
	timeout time.Duration
	l       logger.Logger
}

func NewCCNUService(l logger.Logger) CCNUService {
	return &ccnuService{
		timeout: time.Minute * 2,
		l:       l,
	}
}
