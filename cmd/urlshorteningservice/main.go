package main

import (
	"context"
	"url-shortening-service/internal/application"
	"url-shortening-service/internal/domain"
	"url-shortening-service/internal/infrastructure/http"
	"url-shortening-service/internal/infrastructure/mocks"
	rediswrap "url-shortening-service/internal/infrastructure/redis"

	"github.com/redis/go-redis/v9"
)

func main() {
	logger := domain.StdoutLogger
	localStorage := mocks.NewLocalStorage()

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	localCache := rediswrap.NewRedisStorage(redisClient, logger)
	idGenerator, err := rediswrap.NewRedisIdGenerator(context.Background(), redisClient, localStorage)
	if err != nil {
		logger.Error("Failed to create Redis ID generator")
		return
	}

	getUrlCase := application.NewUrlGetter(localCache, localStorage, logger)
	shortenUrlCase := application.NewUrlShortener(idGenerator, localStorage)

	server := http.NewSimpleServer(*shortenUrlCase, *getUrlCase, logger, "8080")
	logger.Info("Starting server")
	server.Start()
	logger.Info("Server closed")
}
