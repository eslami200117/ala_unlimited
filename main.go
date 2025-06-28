package main

import (
	"fmt"
	"net/http"

	"github.com/eslami200117/ala_unlimited/config"
	"github.com/eslami200117/ala_unlimited/handler"
	"github.com/eslami200117/ala_unlimited/pkg/comm"
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
	reqChn, resChn := make(chan string), make(chan string)

	//tb := telegrambot.NewTelBot(conf.TelegramBotToken, conf.Debug, reqChn, resChn)
	//go tb.RunBot()

	coreService := service.NewCore(conf, reqChn, resChn)
	go server.NewGRPCServer().StartGRPC(coreService)
	coreService.Start(250, 0)
	coreService.SetSellers(map[int]string{
		1720400: "پاورتک شاپ",
		1105946: "تک ترند",
	})

	api := handler.NewApi(coreService)
	r := server.NewChiServer()
	r.Initialize(api)

	logger.Info().
		Msg(fmt.Sprintf("Listening on port %s", conf.Port))
	logger.Fatal().
		Err(http.ListenAndServe(fmt.Sprintf(":%s", conf.Port), r)).
		Msg("Failed to start server")
}
