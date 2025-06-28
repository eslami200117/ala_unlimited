package config

import (
	"os"
	"strconv"
	"time"

	"github.com/eslami200117/ala_unlimited/pkg/comm"
	"github.com/joho/godotenv"
)

type Config struct {
	TelegramBotToken string
	TelegramChatID   string
	TelegramAPIURL   string
	DigiKalaAPIURL   string
	Port             string
	CheckInterval    time.Duration
	Debug            bool
	GRPC_PORT        string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return nil, err
	}
	logger := comm.Logger("config")
	tbt := os.Getenv("TELEGRAM_BOT_TOKEN")
	tci := os.Getenv("TELEGRAM_CHAT_ID")
	tau := os.Getenv("TELEGRAM_API_URL")
	dau := os.Getenv("DIGIKALA_API_URL")
	port := os.Getenv("PORT")
	debugMode := os.Getenv("DEBUG_MODE")
	grpcPort := os.Getenv("GRPC_PORT")
	check, err := strconv.Atoi(os.Getenv("CHEKING_INTERVAL"))
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to parse check interval")
	}
	return &Config{
		TelegramBotToken: tbt,
		TelegramChatID:   tci,
		TelegramAPIURL:   tau,
		DigiKalaAPIURL:   dau,
		Port:             port,
		CheckInterval:    time.Duration(check),
		Debug:            debugMode == "true",
		GRPC_PORT:        grpcPort,
	}, nil

}
