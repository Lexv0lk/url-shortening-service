package http

import (
	"net/http"
	"sync"
	"url-shortening-service/internal/application"
	"url-shortening-service/internal/domain"
	"url-shortening-service/internal/infrastructure/http/handlers"
)

type HandlersServer struct {
	mux       *http.ServeMux
	urlAdder  application.UrlShortener
	urlGetter application.UrlGetter
	logger    domain.Logger

	once *sync.Once
}

func NewSimpleServer(urlAdder application.UrlShortener, urlGetter application.UrlGetter, logger domain.Logger) *HandlersServer {
	return &HandlersServer{
		mux:       http.NewServeMux(),
		urlAdder:  urlAdder,
		urlGetter: urlGetter,
		logger:    logger,
	}
}

func (s *HandlersServer) Start() {
	s.once.Do(s.startServer)
}

func (s *HandlersServer) startServer() {
	mux := http.NewServeMux()
	addUrlHandler := handlers.NewAddUrlHandler(s.urlAdder, s.logger)
	redirectHandler := handlers.NewRedirectHandler(s.urlGetter)

	mux.HandleFunc("POST /add-url", addUrlHandler.Create)
	mux.HandleFunc("GET /{urlToken}", redirectHandler.Redirect)

	go func() {
		if err := http.ListenAndServe(":8080", mux); err != nil {
			s.logger.Error("Failed to start HTTP server: " + err.Error())
		}
	}()
}
