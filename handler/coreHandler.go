package handler

import (
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

func NewApi(notifi *service.Core) *Api {
	_logger := zerolog.New(os.Stderr).
		With().Str("package", "api").
		Caller().Timestamp().Logger()

	return &Api{
		core:   notifi,
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
	w.WriteHeader(http.StatusOK)
}
