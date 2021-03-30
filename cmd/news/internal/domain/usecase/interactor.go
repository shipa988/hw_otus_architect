package usecase

import (
	"context"
	"github.com/pkg/errors"
	"github.com/shipa988/hw_otus_architect/internal/data/controller/log"
	"github.com/shipa988/hw_otus_architect/internal/domain/entity"
	"sort"
	"time"
)

const limitFriends = 50

const (
	errHash    = "can't hash pass"
	errCompare = "can't compare pass and hash"
)

var _ NetworkNews = (*Interactor)(nil)

type Interactor struct {
	userRepo  entity.UserRepository
	newsRepo  entity.NewsRepository
	newsCache entity.NewsCache
	cacheSize int
}

func (i Interactor) InvalidateNewsCache(ctx context.Context, batchSize int, batchFlush time.Duration) error {
	err := i.newsCache.Invalidate(ctx, batchSize, batchFlush)
	if err != nil {
		return errors.Wrap(err, "can't update news cache")
	}
	return nil
}

func NewInteractor(userRepo entity.UserRepository, newsRepo entity.NewsRepository, newsCache entity.NewsCache, cacheSize int) *Interactor {
	return &Interactor{userRepo: userRepo, newsRepo: newsRepo, newsCache: newsCache, cacheSize: cacheSize}
}

func (i Interactor) GetNews(ctx context.Context, userID uint64) ([]entity.News, error) {
	news := i.newsCache.GetNews(ctx, userID)
	if len(news) == 0 {
		ok := true
		var last_friend_id uint64
		for ok {
			friends, err := i.userRepo.GetFriendsById(ctx, userID, limitFriends, last_friend_id)
			if err != nil {
				log.Error(errors.Wrap(err, "can't GetFriendsById"))
			}
			if len(friends) == 0 {
				ok = false
				break
			}
			for _, friend := range friends {
				user, err := i.userRepo.GetUserById(ctx, friend.Id)
				if err != nil {
					log.Error(errors.Wrap(err, "can't StartGettingNews"))
				}
				newsDB, err := i.newsRepo.GetNews(ctx, friend.Id, i.cacheSize)
				if err != nil {
					log.Error(errors.Wrap(err, "can't StartGettingNews"))
				}
				for _, n := range newsDB {
					n.AuthorName = user.Name
					n.AuthorGen = string(user.Gen)
					n.AuthorSurName = user.SurName
					news = append(news, n)
				}
				//news = append(news, newsDB...)
				if len(news) >= i.cacheSize {
					ok = false
					break
				}
				last_friend_id = friend.Id
			}
			if len(friends) < limitFriends {
				ok = false
				break
			}
		}
		i.newsCache.PutNews(ctx, userID, news)
	}
	sort.Slice(news, func(i, j int) bool {
		return news[j].Time.Before(news[i].Time)
	})
	return news, nil
}
