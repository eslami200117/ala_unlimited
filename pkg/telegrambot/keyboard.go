package telegrambot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func GetMainMenu() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Start Core", "start_core"),
			tgbotapi.NewInlineKeyboardButtonData("Quit Core", "quit_core"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Update Seller", "update_seller"),
			tgbotapi.NewInlineKeyboardButtonData("Log", "log"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Add User", "add_user"),
			tgbotapi.NewInlineKeyboardButtonData("Remove Users", "remove_user"),
		),
	)
}
