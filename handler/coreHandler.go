package handler

import (
	"encoding/json"
	"github.com/rs/zerolog"
	"net/http"
	"os"
	"strconv"

	"github.com/eslami200117/ala_unlimited/service"
)

type Api struct {
	token  string
	core   *service.Core
	logger zerolog.Logger
}

func NewApi(_coer *service.Core) *Api {
	_logger := zerolog.New(os.Stderr).
		With().Str("package", "api").
		Caller().Timestamp().Logger()

	return &Api{
		core:   _coer,
		logger: _logger,
	}
}

func (api *Api) StartCore(w http.ResponseWriter, r *http.Request) {
	// Read query parameters
	maxRateStr := r.URL.Query().Get("maxRateForThread")
	durationStr := r.URL.Query().Get("duration")

	// Convert to int
	duration, err := strconv.Atoi(durationStr)
	if err != nil {
		api.logger.Error().Err(err).Str("durationStr", durationStr).Msg("durationStr must be an integer")
		http.Error(w, "Invalid duration", http.StatusBadRequest)
		return
	}

	maxRate, err := strconv.Atoi(maxRateStr)
	if err != nil {
		api.logger.Error().Err(err).Str("maxRate", maxRateStr).Msg("maxRate must be an integer")
		http.Error(w, "Invalid maxRate", http.StatusBadRequest)
		return
	}

	err = api.core.Start(maxRate, duration)
	if err != nil {
		api.logger.Error().Err(err).Msg("core start failed")
		http.Error(w, "Error in Start", http.StatusBadRequest)
		return
	}
	api.logger.Info().
		Str("remote", r.RemoteAddr).
		Str("duration", durationStr).
		Str("maxRate", maxRateStr).
		Msg("core started")

	w.WriteHeader(http.StatusOK)
}

func (api *Api) QuitCore(w http.ResponseWriter, _ *http.Request) {
	api.core.Quit()
	w.WriteHeader(http.StatusOK)
}

func (api *Api) UpdateSeller(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var update map[int]string
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	api.core.SetSellers(update)

	api.logger.Info().Msg("sellerMap updated")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("sellerMap updated"))
	if err != nil {
		api.logger.Error().Err(err).Msg("write response failed")
		return
	}

	api.logger.Info().
		Str("remote", r.RemoteAddr).
		Msg("update seller successfully")
}
