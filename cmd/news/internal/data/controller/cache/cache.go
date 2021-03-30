package cache

import (
	"context"
	"encoding/json"
	lru "github.com/hashicorp/golang-lru"
	"github.com/pkg/errors"
	"github.com/shipa988/hw_otus_architect/internal/domain/entity"
	"github.com/shipa988/hw_otus_architect/internal/domain/usecases"
	"time"
)

type newsCache struct {
	//mx sync.RWMutex
	//cache    map[uint64]*newsStack
	lruCache *lru.Cache
	queue    usecases.Queue
	size int
}

var _ entity.NewsCache = (*newsCache)(nil)

// NewNewsCache create NewsCache for getting parsed individual news for each subscriber and caching it
func NewNewsCache(queue usecases.Queue, size int) (*newsCache,error) {
	lru,err:=lru.New(size)
	if err != nil {
		return nil, err
	}
	return &newsCache{
		//cache: make(map[uint64]*newsStack, size),
		queue: queue,
		lruCache:lru,
		size: size,
	},nil
}

func (c newsCache) Invalidate(ctx context.Context,batchSize int, batchFlush time.Duration) error {
	err:=c.queue.Pull(ctx, batchFlush, batchSize, func(ctx context.Context, msgs [][]byte) error {
		sn := &entity.SubNews{}
		var merr error
		m := make(map[uint64][]entity.News,batchSize)
		for _, msg := range msgs {
			err := json.Unmarshal(msg, sn) //todo:easyjson
			if err != nil {
				merr=err
			}
			m[sn.SubId] = append(m[sn.SubId], sn.News)
		}
		for id, news := range m {
			c.PutNews(ctx,id,news)
		}
		return merr
	})
	if err != nil {
		return errors.Wrap(err, "can't Invalidate News cache")
	}
	return nil
}

func (c *newsCache) ContainsSubscriber(ctx context.Context, subscriberId uint64) bool {
	//_, ok := c.cache[subscriberId]
	return c.lruCache.Contains(subscriberId)
}

func (c *newsCache) GetNews(ctx context.Context, subscriberId uint64) []entity.News {
	/*c.mx.RLock()
	defer c.mx.RUnlock()
	if v, ok := c.cache[subscriberId]; ok {
		return v.arr
	}
	return nil*/
	n,ok:=c.lruCache.Get(subscriberId)
	if ok{
		if v, isnews := n.(*newsStack); isnews {
			return v.arr
		}
	}
	return nil
}

func (c *newsCache) PutNews(ctx context.Context, subscriberId uint64, news []entity.News) {
	//c.mx.Lock()
	//defer c.mx.Unlock()
	n, ok := c.lruCache.Peek(subscriberId)
	if ok {
		if v, isnews := n.(*newsStack); isnews {
			for _, onenews := range news {
				v.Push(onenews)
			}
			c.lruCache.Add(subscriberId, v)
		}
	}else {
		v:=newNewsStack(c.size)
		for _, on := range news {
			v.Push(on)
		}
		c.lruCache.Add(subscriberId, v)
	}
	/*for _, onenews := range news {

		c.cache[subscriberId].Push(onenews)
	}*/
}

func (c *newsCache) PutOneNews(ctx context.Context, subscriberId uint64, news entity.News) {
	/*c.mx.Lock()
	defer c.mx.Unlock()
	c.cache[subscriberId].Push(news)*/
	n, ok := c.lruCache.Peek(subscriberId)
	if ok {
		if v, isnews := n.(*newsStack); isnews {
			v.Push(news)
			c.lruCache.Add(subscriberId, v)
		}
	}else {
		v:=newNewsStack(c.size)
		v.Push(news)
		c.lruCache.Add(subscriberId, v)
	}
}
