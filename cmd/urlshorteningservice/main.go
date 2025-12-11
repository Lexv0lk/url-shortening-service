package main

import (
	"context"
	"os"
	"url-shortening-service/internal/application"
	"url-shortening-service/internal/domain"
	"url-shortening-service/internal/infrastructure/http"
	"url-shortening-service/internal/infrastructure/mocks"
	rediswrap "url-shortening-service/internal/infrastructure/redis"

	"github.com/redis/go-redis/v9"
)

func main() {
	redisUrl := "localhost"
	redisPort := "6379"
	serverPort := "8080"

	if envRedisUrl, found := os.LookupEnv(domain.RedisUrlEnv); found {
		redisUrl = envRedisUrl
	}
	if envRedisPort, found := os.LookupEnv(domain.RedisPortEnv); found {
		redisPort = envRedisPort
	}
	if envServerPort, found := os.LookupEnv(domain.ServerPortEnv); found {
		serverPort = envServerPort
	}

	logger := domain.StdoutLogger
	localStorage := mocks.NewLocalStorage()

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisUrl + ":" + redisPort,
	})
	localCache := rediswrap.NewRedisStorage(redisClient, logger)
	idGenerator, err := rediswrap.NewRedisIdGenerator(context.Background(), redisClient, localStorage)
	if err != nil {
		logger.Error("Failed to create Redis ID generator")
		return
	}

	getUrlCase := application.NewUrlGetter(localCache, localStorage, logger)
	shortenUrlCase := application.NewUrlShortener(idGenerator, localStorage)

	server := http.NewSimpleServer(*shortenUrlCase, *getUrlCase, logger, serverPort)
	logger.Info("Starting server")
	server.Start()
	logger.Info("Server closed")
}
