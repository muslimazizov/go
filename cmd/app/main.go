package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"gogogo/config"
	"gogogo/internal/good"
	"gogogo/pkg/brokers/nats"
	"gogogo/pkg/cache/redis"
	"gogogo/pkg/store/postgres"

	"github.com/go-chi/chi/v5"
	redix "github.com/gomodule/redigo/redis"
)

func main() {
	cfg := config.New()

	// postgres
	store, err := postgres.New(cfg)
	if err != nil {
		log.Fatalf("cant create conn to db: %v", err)
	}
	defer store.Pool.Close()

	// redis
	r := redis.New(cfg)
	redisConn := r.Pool.Get()
	err = store.Pool.Ping(context.Background())
	if err != nil {
		log.Fatalf("cant ping redis: %v", err)
	}
	defer func(redisConn redix.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Printf("cant close redis conn: %v", err)
		}
	}(redisConn)

	// nats
	n, err := nats.New(cfg)
	if err != nil {
		log.Fatalf("cant create conn to nats: %v", err)
	}
	defer n.Conn.Close()

	goodStore := good.NewStore(store.Pool, n.Conn)
	goodService := good.NewService(goodStore, redisConn)

	router := chi.NewRouter()
	router.Get("/goods/list", good.ListGoods(goodService))
	router.Post("/good/create", good.CreateGood(goodService))
	router.Delete("/good/delete", good.DeleteGood(goodService))
	router.Patch("/good/update", good.UpdateGood(goodService))
	router.Patch("/good/reprioritize", good.ReprioritizeGood(goodService))

	server := http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Handler: router,
	}

	fmt.Println("starting server", server.Addr)
	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("server error: %v\n", err)
	}
}
