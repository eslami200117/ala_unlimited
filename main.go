package main

import (
	"log"
	"net/http"

	"github.com/eslami200117/ala_unlimited/handler"
	"github.com/eslami200117/ala_unlimited/server"
	"github.com/eslami200117/ala_unlimited/service"
)

const (
	TELEGRAM_BOT_TOKEN = "6288180955:AAHXFhXjmwDDIuWi42ki3MGRBKgkZVanrhQ"
	TELEGRAM_CHAT_ID   = "1785120852"
)

func main() {
	notifService := service.NewCore(TELEGRAM_BOT_TOKEN, TELEGRAM_CHAT_ID)
	api := handler.NewApi(notifService)
	r := server.NewAlaServer()
	r.Initialize(api)
	log.Println("Server started on port 3000")
	log.Fatal(http.ListenAndServe(":3000", r))
}
