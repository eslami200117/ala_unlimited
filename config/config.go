package config

import (
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	TelegramBotToken string
	TelegramChatID   string
	TelegramAPIURL   string
	DigiKalaAPIURL   string
	Port             string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return nil, err
	}

	tbt := os.Getenv("TELEGRAM_BOT_TOKEN")
	tci := os.Getenv("TELEGRAM_CHAT_ID")
	tau := os.Getenv("TELEGRAM_API_URL")
	dau := os.Getenv("DG_API_URL")
	port := os.Getenv("PORT")
	return &Config{
		TelegramBotToken: tbt,
		TelegramChatID:   tci,
		TelegramAPIURL:   tau,
		DigiKalaAPIURL:   dau,
		Port:             port,
	}, nil

}
