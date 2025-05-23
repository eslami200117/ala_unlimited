package telegrambot

import (
	"bufio"
	"encoding/json"
	"os"
	"strconv"
	"strings"

	"github.com/eslami200117/ala_unlimited/pkg/comm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
)

type TelBot struct {
	bot    *tgbotapi.BotAPI
	logger zerolog.Logger
	admins map[int64]bool
	reqChn chan string
	resChn chan string
	state  map[int64]string
}

func NewTelBot(token string, debug bool, _reqChn, _resChn chan string) TelBot {
	_logger := comm.Logger("telegram_bot")
	_bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		_logger.Fatal().Err(err).Msg("failed to create bot")
	}
	_admins := make(map[int64]bool)
	_bot.Debug = debug
	_logger.Info().Msgf("Authorized on account %s", _bot.Self.UserName)
	_logger.Info().Msgf("Bot ID: %d", _bot.Self.ID)

	return TelBot{
		bot:    _bot,
		logger: _logger,
		admins: _admins,
		reqChn: _reqChn,
		resChn: _resChn,
		state:  make(map[int64]string),
	}
}

func (t TelBot) RunBot() {
	t.loadAllowedUsers()
	defer t.saveAllowedUsers()
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 10

	updates := t.bot.GetUpdatesChan(u)

	for update := range updates {
		var userID int64
		if update.Message != nil {
			userID = update.Message.From.ID
		} else if update.CallbackQuery != nil {
			userID = update.CallbackQuery.From.ID
		}

		if _, ok := t.admins[userID]; !ok {
			t.logger.Warn().Msgf("User %d is not allowed", userID)
			continue
		}

		t.handle(update)
	}
}

// Load allowed users from the file into memory
func (t TelBot) loadAllowedUsers() {
	file, err := os.Open("admins.txt")
	if err != nil {
		if os.IsNotExist(err) {
			t.logger.Warn().Msg("admins.txt not found, starting with empty list")
			return
		}
		t.logger.Fatal().Err(err).Msg("Error opening admins.txt")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		id, err := strconv.ParseInt(line, 10, 64)
		if err != nil {
			t.logger.Warn().Str("line", line).Msg("Skipping invalid admin ID")
			continue
		}
		t.admins[id] = true
	}

}

// Save allowed users back to the file
func (t TelBot) saveAllowedUsers() {
	t.logger.Info().Msg("Saving allowed users")

	ids := make([]int64, 0, len(t.admins))
	for id := range t.admins {
		if !t.admins[id] {
			continue
		}
		ids = append(ids, id)
	}

	data, err := json.MarshalIndent(ids, "", "  ")
	if err != nil {
		t.logger.Error().Err(err).Msg("Error marshaling allowed users")
		return
	}

	if err := os.WriteFile("admins.txt", data, 0644); err != nil {
		t.logger.Error().Err(err).Msg("Error writing admins.txt")
	}
}
