package telegrambot

import (
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (t *TelBot) handle(update tgbotapi.Update) {
	if update.Message != nil {
		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID

		switch t.state[userID] {
		case "awaiting_user_id_add":
			t.logger.Info().Msgf("Adding user: %s", update.Message.Text)
			id, err := strconv.ParseInt(update.Message.Text, 10, 64)
			if err != nil {
				t.bot.Send(tgbotapi.NewMessage(chatID, "❌ Please send a valid number."))
				return
			}
			t.admins[id] = false
			t.state[userID] = ""
			t.bot.Send(tgbotapi.NewMessage(chatID, "✅ User added."))
			t.showMenu(chatID)
		case "awaiting_user_id_remove":
			t.logger.Info().Msgf("Removing user: %s", update.Message.Text)
			id, err := strconv.ParseInt(update.Message.Text, 10, 64)
			if err != nil {
				t.bot.Send(tgbotapi.NewMessage(chatID, "❌ Please send a valid number."))
				return
			}
			delete(t.admins, id)
			t.state[userID] = ""
			t.bot.Send(tgbotapi.NewMessage(chatID, "✅ User removed."))
			t.showMenu(chatID)
		case "awaiting_core_input":
			t.reqChn <- "start"
			t.reqChn <- update.Message.Text
			res := <-t.resChn
			if res != "OK" {
				t.bot.Send(tgbotapi.NewMessage(chatID, res))
				return
			}
			t.state[userID] = ""
			t.bot.Send(tgbotapi.NewMessage(chatID, "✅ Core started."))
			t.showMenu(chatID)

		case "awaiting_seller_map":
			t.reqChn <- "set"
			t.reqChn <- update.Message.Text
			res := <-t.resChn
			if res != "OK" {
				t.bot.Send(tgbotapi.NewMessage(chatID, res))
				return
			}

			t.state[userID] = ""
			t.bot.Send(tgbotapi.NewMessage(chatID, "✅ Sellers updated."))
			t.showMenu(chatID)

		default:
			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "start":
					t.showMenu(chatID)
				}
			}
		}

	} else if update.CallbackQuery != nil {
		t.bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data))

		userID := update.CallbackQuery.From.ID
		chatID := update.CallbackQuery.Message.Chat.ID

		switch update.CallbackQuery.Data {
		case "add_user":
			t.state[userID] = "awaiting_user_id_add"
			t.bot.Send(tgbotapi.NewMessage(chatID, "👤 Please send the user ID to add:"))
		case "remove_user":
			t.state[userID] = "awaiting_user_id_remove"
			t.bot.Send(tgbotapi.NewMessage(chatID, "🗑️ Please send the user ID to remove:"))
		case "start_core":
			t.state[userID] = "awaiting_core_input"
			t.bot.Send(tgbotapi.NewMessage(chatID, "⚙️ Send `max_rate duration` (e.g., `10 30`):"))
		case "update_seller":
			t.state[userID] = "awaiting_seller_map"
			t.bot.Send(tgbotapi.NewMessage(chatID, "📦 Send each seller like:\n```\n101 John\n102 Alice\n```"))
		case "quit_core":
			t.reqChn <- "done"
			t.logger.Info().Msg("Core stopped.")
			t.bot.Send(tgbotapi.NewMessage(chatID, "🛑 Core stopped."))
			t.showMenu(chatID)
		case "log":
			t.reqChn <- "log"
			t.bot.Send(tgbotapi.NewMessage(chatID, "📜 Logs coming soon..."))
			log := <-t.resChn
			t.bot.Send(tgbotapi.NewMessage(chatID, log))
			t.showMenu(chatID)
		}

	}
	// t.showMenu(update.Message.Chat.ID)
}

func (t *TelBot) showMenu(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Choose an option:")
	msg.ReplyMarkup = GetMainMenu()
	t.bot.Send(msg)
}
