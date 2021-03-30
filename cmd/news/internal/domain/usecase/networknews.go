package usecase

import (
	"context"
	"github.com/shipa988/hw_otus_architect/internal/domain/entity"
	"time"
)

type NetworkNews interface {
	GetNews(ctx context.Context, userID uint64) ([]entity.News, error)
	InvalidateNewsCache(ctx context.Context,batchSize int, batchFlush time.Duration) error
}
