package queue

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/shipa988/hw_otus_architect/internal/data/controller/log"
	"github.com/shipa988/hw_otus_architect/internal/domain/entity"
	"github.com/shipa988/hw_otus_architect/internal/domain/usecases"
	"time"
)

type newsHub struct {
	newsQueue usecases.Queue
	subRepo   entity.UserRepository
	subQueue  usecases.Queue
}

func NewPublisher(newsQueue usecases.Queue) *newsHub {
	return &newsHub{newsQueue: newsQueue}
}

func NewSubscriber(newsQueue usecases.Queue, subRepo entity.UserRepository, subQueue usecases.Queue) *newsHub {
	return &newsHub{newsQueue: newsQueue, subRepo: subRepo, subQueue: subQueue}
}

func (h newsHub) StartGettingNews(ctx context.Context, batchSize int, batchFlush time.Duration) {
	h.newsQueue.Pull(ctx, batchFlush, batchSize, func(ctx context.Context, msgs [][]byte) error {
		n := &entity.News{}
		var merr error
		for _, msg := range msgs {
			err := json.Unmarshal(msg, n) //todo:easyjson
			if err != nil {
				merr = err
			}

			user, err := h.subRepo.GetUserById(ctx, n.AuthorId)
			if err != nil {
				log.Error(errors.Wrap(err, "can't StartGettingNews"))
			}
			n.AuthorGen=string(user.Gen)
			n.AuthorName=user.Name
			n.AuthorSurName=user.SurName

			friends, err := h.subRepo.GetSubscribersIdById(ctx, n.AuthorId)
			if err != nil {
				log.Error(errors.Wrap(err, "can't GetFriendsById"))
			}
			for _, friendId := range friends {
				nb, err := json.Marshal(entity.SubNews{ //todo:easyjson
					SubId: friendId,
					News:  *n,
				})
				if err != nil {
					log.Error(errors.Wrap(err, "can't Marshal SubNews"))
				}
				err = h.subQueue.Push(ctx, nb)
				if err != nil {
					log.Error(errors.Wrap(err, "can't Push SubNews"))
				}
			}
		}
		return merr
	})
}

func (h newsHub) SaveNews(ctx context.Context, news entity.News) error {
	nb, err := json.Marshal(news)
	if err != nil {
		return errors.Wrap(err, "can't Marshal News")
	}
	h.newsQueue.Push(ctx, nb)
	if err != nil {
		return errors.Wrap(err, "can't Push News")
	}
	return nil
}
