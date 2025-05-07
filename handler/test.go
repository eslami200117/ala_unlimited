package handler

import (
	"net/http"
	"strconv"

	"github.com/eslami200117/ala_unlimited/service"
)

type Api struct {
	token  string
	notifi *service.Core
}

func NewApi(notifi *service.Core) *Api {
	return &Api{
		notifi: notifi,
	}
}

func (api *Api) StartTest(w http.ResponseWriter, r *http.Request) {
	// Read query parameters
	maxRateStr := r.URL.Query().Get("maxRateForThread")
	durationStr := r.URL.Query().Get("duration")

	// Convert to int
	duration, err := strconv.Atoi(durationStr)
	if err != nil {
		http.Error(w, "Invalid numThread", http.StatusBadRequest)
		return
	}

	maxRate, err := strconv.Atoi(maxRateStr)
	if err != nil {
		http.Error(w, "Invalid numThread", http.StatusBadRequest)
		return
	}

	api.notifi.ReConfig(maxRate, duration)
	err = api.notifi.Start()
	if err != nil {
		http.Error(w, "Invalid numThread", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
