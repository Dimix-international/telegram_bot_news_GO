package bot

import (
	"app/internal/botkit"
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

//создаем view которая напишет helloWorld на команду start


func ViewCmdStart() botkit.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
		if _, err := bot.Send(tgbotapi.NewMessage(update.FromChat().ID, "Hello world")); err != nil {
			return err
		}

		return nil
	}
}