package main

import (
	"url-shortening-service/internal/application"
	"url-shortening-service/internal/domain"
	"url-shortening-service/internal/infrastructure/http"
	"url-shortening-service/internal/infrastructure/mocks"
)

func main() {
	localCache := mocks.NewLocalCache()
	localStorage := mocks.NewLocalStorage()
	logger := domain.StdoutLogger
	idGenerator := mocks.NewSlowSafeIDGenerator()

	getUrlCase := application.NewUrlGetter(localCache, localStorage, logger)
	shortenUrlCase := application.NewUrlShortener(idGenerator, localStorage)

	server := http.NewSimpleServer(*shortenUrlCase, *getUrlCase, logger, "8080")
	logger.Info("Starting server")
	server.Start()
	logger.Info("Server closed")
}
