package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"url-shortening-service/internal/domain"
)

type StatsShowHandler struct {
	statsCalculator domain.StatisticsCalculator
	logger          domain.Logger
}

func NewStatsShowHandler(statsCalculator domain.StatisticsCalculator, logger domain.Logger) *StatsShowHandler {
	return &StatsShowHandler{
		statsCalculator: statsCalculator,
		logger:          logger,
	}
}

func (h *StatsShowHandler) Show(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue(domain.UrlTokenStr)

	stats, err := h.statsCalculator.CalculateStatistics(r.Context(), token)
	if errors.Is(err, &domain.TokenNonExistingError{}) {
		http.Error(w, "Statistics not found for the given URL token", http.StatusNotFound)
		return
	} else if err != nil {
		h.logger.Error("Failed to calculate statistics: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(stats); err != nil {
		h.logger.Error(fmt.Sprintf("Failed to encode response: %v", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(buf.Bytes())
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to write response: %v", err))
	}
}
