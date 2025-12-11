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
	port      string

	once *sync.Once
}

func NewSimpleServer(urlAdder application.UrlShortener, urlGetter application.UrlGetter,
	logger domain.Logger, port string) *HandlersServer {
	return &HandlersServer{
		mux:       http.NewServeMux(),
		urlAdder:  urlAdder,
		urlGetter: urlGetter,
		logger:    logger,
		once:      &sync.Once{},
		port:      port,
	}
}

func (s *HandlersServer) Start() {
	s.once.Do(s.startServer)
}

func (s *HandlersServer) startServer() {
	mux := http.NewServeMux()
	addUrlHandler := handlers.NewAddUrlHandler(s.urlAdder, s.logger)
	redirectHandler := handlers.NewRedirectHandler(s.urlGetter)

	mux.HandleFunc(domain.AddUrlAddress, addUrlHandler.Create)
	mux.HandleFunc(domain.RedirectAddress, redirectHandler.Redirect)

	if err := http.ListenAndServe(":"+s.port, mux); err != nil {
		s.logger.Error("Failed to start HTTP server: " + err.Error())
	}
}
