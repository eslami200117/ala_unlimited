package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	TelegramBotToken string
	TelegramChatID   string
	TelegramAPIURL   string
	DigiKalaAPIURL   string
	Port             string
	CheckInterval    time.Duration
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return nil, err
	}

	tbt := os.Getenv("TELEGRAM_BOT_TOKEN")
	tci := os.Getenv("TELEGRAM_CHAT_ID")
	tau := os.Getenv("TELEGRAM_API_URL")
	dau := os.Getenv("DIGIKALA_API_URL")
	port := os.Getenv("PORT")
	check, err := strconv.Atoi(os.Getenv("CHEKING_INTERVAL"))
	if err != nil {
		log.Fatal(err)
	}
	return &Config{
		TelegramBotToken: tbt,
		TelegramChatID:   tci,
		TelegramAPIURL:   tau,
		DigiKalaAPIURL:   dau,
		Port:             port,
		CheckInterval:    time.Duration(check),
	}, nil

}
