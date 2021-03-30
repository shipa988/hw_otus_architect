package usecases

import (
	"context"
	"github.com/shipa988/hw_otus_architect/internal/domain/entity"
	"time"
)

type NewsPulisher interface {
	SaveNews(ctx context.Context, news entity.News) error
}

type NewsSubscriber interface {
	StartGettingNews(ctx context.Context, batchSize int, batchFlush time.Duration)
}