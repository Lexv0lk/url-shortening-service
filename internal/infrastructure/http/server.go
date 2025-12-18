package http

import (
	"net/http"
	"sync"
	"url-shortening-service/internal/domain"
	"url-shortening-service/internal/infrastructure/http/handlers"
)

// HandlersServer is the HTTP server that handles all URL shortening service endpoints.
// It registers handlers for URL creation, retrieval, update, deletion, and statistics.
type HandlersServer struct {
	mux             *http.ServeMux
	urlAdder        domain.UrlShortener
	urlGetter       domain.UrlGetter
	urlUpdater      domain.UrlUpdater
	urlDeleter      domain.UrlDeleter
	statsSender     domain.StatisticsSender
	statsCalculator domain.StatisticsCalculator
	logger          domain.Logger
	port            string

	once *sync.Once
}

// NewSimpleServer creates a new HandlersServer instance with all required dependencies.
func NewSimpleServer(
	urlAdder domain.UrlShortener,
	urlGetter domain.UrlGetter,
	urlUpdater domain.UrlUpdater,
	urlDeleter domain.UrlDeleter,
	statsSender domain.StatisticsSender,
	statsCalculator domain.StatisticsCalculator,
	logger domain.Logger,
	port string,
) *HandlersServer {
	return &HandlersServer{
		mux:             http.NewServeMux(),
		urlAdder:        urlAdder,
		urlGetter:       urlGetter,
		urlUpdater:      urlUpdater,
		urlDeleter:      urlDeleter,
		statsSender:     statsSender,
		statsCalculator: statsCalculator,
		logger:          logger,
		once:            &sync.Once{},
		port:            port,
	}
}

// Start starts the HTTP server. This method is safe to call multiple times;
// the server will only be started once.
// The server listens on the configured port and blocks until an error occurs.
func (s *HandlersServer) Start() {
	s.once.Do(s.startServer)
}

func (s *HandlersServer) startServer() {
	mux := http.NewServeMux()
	shortenUrlHandler := handlers.NewAddUrlHandler(s.urlAdder, s.logger)
	redirectHandler := handlers.NewRedirectHandler(s.urlGetter, s.statsSender, s.logger)
	updateUrlHandler := handlers.NewUpdateUrlHandler(s.urlUpdater, s.logger)
	deleteUrlHandler := handlers.NewDeleteUrlHandler(s.urlDeleter, s.logger)
	statsHandler := handlers.NewStatsShowHandler(s.statsCalculator, s.logger)

	mux.HandleFunc(domain.ShortenUrlAddress, shortenUrlHandler.Create)
	mux.HandleFunc(domain.RedirectAddress, redirectHandler.Redirect)
	mux.HandleFunc(domain.UpdateUrlAddress, updateUrlHandler.Update)
	mux.HandleFunc(domain.DeleteUrlAddress, deleteUrlHandler.Delete)
	mux.HandleFunc(domain.StatsUrlAddress, statsHandler.Show)

	if err := http.ListenAndServe(":"+s.port, mux); err != nil {
		s.logger.Error("Failed to start HTTP server: " + err.Error())
	}
}
