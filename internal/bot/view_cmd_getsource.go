package bot

import (
	"app/internal/botkit"
	"app/internal/model"
	"context"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type SourceProvider interface {
	SourceById(ctx context.Context, id int64) (*model.Source, error)
}

func VewCmdGetSource(provider SourceProvider) botkit.ViewFunc {
	return func (ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
		idStr := update.Message.CommandArguments()

		id, err := strconv.ParseInt(idStr, 10, 64)

		if err != nil {
			return err
		}

		source, err := provider.SourceById(ctx, id)

		if err != nil {
			return err
		}

		reply := tgbotapi.NewMessage(update.Message.Chat.ID, formatSource(*source))
		reply.ParseMode = "MarkdownV2" 

		if _, err := bot.Send(reply); err != nil {
			return err
		}

		return nil
	}
}