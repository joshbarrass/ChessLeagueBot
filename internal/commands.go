package internal

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type CommandFunc func(tgbotapi.Update) (tgbotapi.Message, error)

// CommandUnknown is for any unknown command
func CommandUnknown(update tgbotapi.Update) (tgbotapi.MessageConfig, error) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command")
	return msg, nil
}

// CommandStart is the /start command
func CommandStart(update tgbotapi.Update) (tgbotapi.MessageConfig, error) {
	msg, _ := CommandInfo(update)
	msg.Text = fmt.Sprintf("Welcome to the Chess League Bot!\n\n%s", msg.Text)
	return msg, nil
}

// CommandInfo is the /info command
func CommandInfo(update tgbotapi.Update) (tgbotapi.MessageConfig, error) {
	text := `To get started, add the bot to a chat and use the /newleague command to start a chess league in that chat. Your players can then use the /joinleague command to enter the league.

You can check out the source code for this bot on <a href="https://github.com/joshbarrass/ChessLeagueBot">GitHub</a>`
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "html"
	return msg, nil
}
