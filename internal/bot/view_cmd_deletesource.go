package bot

import (
	"app/internal/botkit"
	"context"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type SourceDeleter interface {
	Delete(ctx context.Context, sourceID int64) error
}

func ViewCmdDeleteSource(deleter SourceDeleter) botkit.ViewFunc {
	return func (ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
		//получаем агрументы
		idStr := update.Message.CommandArguments()

		//парсим id
		id, err := strconv.ParseInt(idStr,10, 64)

		if err != nil {
			return err
		}

		if err := deleter.Delete(ctx, id); err != nil {
			return err
		}

		//создаем и отправляем сообщение что удалили
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Источник успешно удален")

		if _, err := bot.Send(msg); err != nil {
			return err
		}

		return nil
	}
}