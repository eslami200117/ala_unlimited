package main

import (
	"fmt"
	"github.com/eslami200117/ala_unlimited/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"

	"github.com/eslami200117/ala_unlimited/handler"
	"github.com/eslami200117/ala_unlimited/server"
	"github.com/eslami200117/ala_unlimited/service"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stderr).
		With().Str("package", "service").
		Caller().Timestamp().Logger()

	conf, err := config.LoadConfig()
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to load config")
	}

	coreService := service.NewCore(conf, logger)
	api := handler.NewApi(coreService)
	r := server.NewAlaServer()
	r.Initialize(api)
	log.Info().Msg(fmt.Sprintf("Listening on port %s", conf.Port))
	log.Fatal().
		Err(http.ListenAndServe(fmt.Sprintf(":%s", conf.Port), r)).
		Msg("Failed to start server")
}
