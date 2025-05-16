package main

import (
	"fmt"
	"net/http"

	"github.com/eslami200117/ala_unlimited/config"
	"github.com/eslami200117/ala_unlimited/pkg/comm"

	"github.com/eslami200117/ala_unlimited/handler"
	"github.com/eslami200117/ala_unlimited/server"
	"github.com/eslami200117/ala_unlimited/service"
)

func main() {
	logger := comm.Logger("main")
	conf, err := config.LoadConfig()
	if err != nil {
		logger.Error().
			Err(err).
			Msg("failed to load config")
	}

	coreService := service.NewCore(conf)

	go server.NewGRPCServer().StartGRPC(coreService)

	api := handler.NewApi(coreService)
	r := server.NewChiServer()
	r.Initialize(api)

	logger.Info().
		Msg(fmt.Sprintf("Listening on port %s", conf.Port))
	logger.Fatal().
		Err(http.ListenAndServe(fmt.Sprintf(":%s", conf.Port), r)).
		Msg("Failed to start server")
}
