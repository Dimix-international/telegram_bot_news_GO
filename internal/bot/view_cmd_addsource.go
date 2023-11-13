package bot

import (
	"app/internal/botkit"
	"app/internal/model"
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type SourceStorage interface {
	Add(ctx context.Context, source model.Source) (int64, error)
}

func ViewCmdAddSource(storage SourceStorage) botkit.ViewFunc {
	type addSourceArgs struct {
		Name string `json:"name"`
		URL string `json:"url"`
	}

	//распарсим агрументы

	return func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
		args, err := botkit.ParseJSON[addSourceArgs](update.Message.CommandArguments()) //распарсим из аргументов команды

		if err != nil {
			return err
		}

		//воссоздаем мета инфу из аргументов
		source := model.Source{
			Name: args.Name,
			FeedUrl: args.URL,
		}

		sourceID, err := storage.Add(ctx, source)

		if err != nil {
			return err
		}

		//конструрируем сообщение что источник добавлен
		var (
			msgText = fmt.Sprintf(
				"Источник добавлен с ID: `%d`\\. Используйте этот ID для обновления источника или удаления\\.",
				sourceID,
			)
			reply = tgbotapi.NewMessage(update.Message.Chat.ID, msgText) //ответ
		)

		reply.ParseMode = "MarkdownV2" //даем телеграму знать что текст парсим как MarkdownV2

		//отправляем сообщение

		if _, err := bot.Send(reply); err != nil {
			return err
		}

		return nil

	}
}