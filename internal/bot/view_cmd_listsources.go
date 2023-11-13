package bot

import (
	"app/internal/botkit"
	"app/internal/botkit/markup"
	"app/internal/model"
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/samber/lo"
)

type SourceListener interface {
	Sources(ctx context.Context) ([]model.Source, error)
}

func ViewCmdListSources(lister SourceListener) botkit.ViewFunc {
	return func (ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
		sources, err := lister.Sources(ctx)

		if err != nil {
			return err
		}

		var (
			sourceInfos = lo.Map(sources, func(source model.Source, _ int) string { return formatSource(source) }) //складываем форматир мета инфо для источников
			msgText     = fmt.Sprintf(
				"Список источников \\(всего %d\\):\n\n%s",
				len(sources),
				strings.Join(sourceInfos, "\n\n"),
			)
		)

		reply := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
		reply.ParseMode = "MarkdownV2" 

		if _, err := bot.Send(reply); err != nil {
			return err
		}

		return nil
	}
}

func formatSource(source model.Source) string {
	return fmt.Sprintf(
		"🌐 *%s*\nID: `%d`\nURL фида: %s",
		markup.EscapeForMarkdown(source.Name),
		source.ID,
		markup.EscapeForMarkdown(source.FeedUrl),
	)
}