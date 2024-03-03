package good

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gomodule/redigo/redis"
)

type Store interface {
	ListGoods(ctx context.Context, params ListGoodsParams) ([]Good, error)
	CreateGood(ctx context.Context, good Good) (Good, error)
	DeleteGood(ctx context.Context, id, projectId int64) (Good, error)
	UpdateGood(ctx context.Context, id, projectId int64, params UpdateGoodParams) (Good, error)

	Count(ctx context.Context) int64
	RemovedCount(ctx context.Context) int64

	ReprioritizeGood(ctx context.Context, id, projectId int64, params ReprioritizeGoodParams) (Good, error)
	GetReprioritizedGoods(ctx context.Context, id int64) ([]ReprioritizedGood, error)
}

type Service struct {
	store Store
	cache redis.Conn
}

func NewService(store Store, cache redis.Conn) Service {
	return Service{
		store: store,
		cache: cache,
	}
}

func (s Service) ListGoods(ctx context.Context, params ListGoodsParams) ([]Good, error) {
	var goods []Good
	res, err := s.cache.Do("GET", "listGoods")
	if err != nil {
		log.Printf("error getting data from redis: %v", err)
	}

	if res == nil {
		goods, err = s.store.ListGoods(ctx, params)
		if err != nil {
			return nil, err
		}

		goodsToRedis, err := json.Marshal(goods)
		if err != nil {
			log.Printf("error marshaling data for redis: %v", err)
		} else {
			_, err = s.cache.Do("SETEX", "listGoods", ListGoodsRedisTTL, goodsToRedis)
			if err != nil {
				log.Printf("error putting data to redis: %v", err)
			}
		}

		return goods, nil
	}

	err = json.Unmarshal(res.([]byte), &goods)
	if err != nil {
		log.Printf("error unmarshaling data from redis: %v", err)

		// get goods from db if unmarshalling from redis is unsuccessful
		goods, err = s.store.ListGoods(ctx, params)
		if err != nil {
			return nil, err
		}
	}

	return goods, nil
}

func (s Service) CreateGood(ctx context.Context, params CreateGoodParams) (Good, error) {
	good := Good{
		Id:          params.Id,
		ProjectId:   params.ProjectId,
		Name:        params.Name,
		Description: params.Description,
		Removed:     false,
		CreatedAt:   time.Now(),
	}

	res, err := s.store.CreateGood(ctx, good)
	if err != nil {
		return Good{}, err
	}

	return res, nil
}

func (s Service) DeleteGood(ctx context.Context, params QueryParams) (Good, error) {
	good, err := s.store.DeleteGood(ctx, params.Id, params.ProjectId)
	if err != nil {
		return Good{}, err
	}

	ok, err := s.cache.Do("DEL", "listGoods")
	log.Printf("invalidating cache, ok: %v, err: %v", ok, err)

	return good, nil
}

func (s Service) UpdateGood(
	ctx context.Context,
	id,
	projectId int64,
	params UpdateGoodParams,
) (Good, error) {
	good, err := s.store.UpdateGood(ctx, id, projectId, params)
	if err != nil {
		return Good{}, err
	}

	ok, err := s.cache.Do("DEL", "listGoods")
	log.Printf("invalidating cache, ok: %v, err: %v", ok, err)

	return good, nil
}

func (s Service) ReprioritizeGood(
	ctx context.Context,
	id,
	projectId int64,
	params ReprioritizeGoodParams,
) ([]ReprioritizedGood, error) {
	good, err := s.store.ReprioritizeGood(ctx, id, projectId, params)
	if err != nil {
		return make([]ReprioritizedGood, 0), err
	}

	goods, err := s.store.GetReprioritizedGoods(ctx, good.Id)
	if err != nil {
		return make([]ReprioritizedGood, 0), err
	}

	ok, err := s.cache.Do("DEL", "listGoods")
	log.Printf("invalidating cache, ok: %v, err: %v", ok, err)

	return goods, nil
}
