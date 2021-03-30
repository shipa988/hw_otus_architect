package entity

import (
	"context"
	"time"
)
type SubNews struct {
	SubId uint64
	News
}

type News struct {
	Id uint64
	AuthorId uint64
	AuthorGen string
	AuthorName string
	AuthorSurName string
	Title string
	Time time.Time
	Text string
}

type NewsRepository interface {
	GetNews(ctx context.Context, authorId uint64, limit int) ([]News, error)
	SaveNews(ctx context.Context, authorId uint64, title, text string) (error)
}

type NewsCache interface {
	GetNews(ctx context.Context, subscriberId uint64) ([]News)
	PutNews(ctx context.Context, subscriberId uint64,news []News)
	Invalidate(ctx context.Context,batchSize int, batchFlush time.Duration) error
}