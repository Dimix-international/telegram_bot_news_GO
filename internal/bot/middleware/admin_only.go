package middleware

import (
	"app/internal/botkit"
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

//для проверки роли пользователя бота

func AddminOnly(channelID int64, next botkit.ViewFunc) botkit.ViewFunc {
	return func (ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
		// получим список админов
		admines, err := bot.GetChatAdministrators(
			tgbotapi.ChatAdministratorsConfig{
				ChatConfig: tgbotapi.ChatConfig{
					ChatID: channelID,
				},
			},
		)

		if err != nil {
			return err
		}

		for _, admin := range admines {
			//проверяем тот кто отправил комманду является ли админом
			if admin.User.ID == update.Message.From.ID {
				return next(ctx, bot, update)
			}
		}
		
		//прерываем выполнение view функции, т.к. он не админ
		if _, err := bot.Send(tgbotapi.NewMessage(
			update.FromChat().ID,
			"У вас нет прав на выполнение этой команды.",
		)); err != nil {
			return err
		}

		return nil

	}
}