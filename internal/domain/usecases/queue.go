package usecases

import (
	"golang.org/x/net/context"
	"time"
)

type Queue interface {
	Push(ctx context.Context, msg []byte) error
	Pull(ctx context.Context, timeToFlush time.Duration, batchSize int, clb func(ctx context.Context, msgs [][]byte) error) error
}