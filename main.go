package main

import (
	"log"
	"net/http"

	"github.com/eslami200117/ala_unlimited/handler"
	"github.com/eslami200117/ala_unlimited/notification"
	"github.com/eslami200117/ala_unlimited/server"
)

const (
	TELEGRAM_BOT_TOKEN = "7602044545:AAFZXb277_yJaInm3bsyNSOc9ESOrgKVmWY"
	TELEGRAM_CHAT_ID = "1785120852"
)
func main(){
	notifService := notification.NewNotifi(TELEGRAM_BOT_TOKEN, TELEGRAM_CHAT_ID)
	api := handler.NewApi(notifService)
	r := server.NewAlaServer()
	r.Initialize(api)
	log.Println("Server started on port 3000")
    log.Fatal(http.ListenAndServe(":3000", r))
}
