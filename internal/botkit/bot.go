package botkit

import (
	"context"
	"log"
	"runtime/debug"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

//весь код для телеграмма


type Bot struct {
	api *tgbotapi.BotAPI
	cmdViews map[string]ViewFunc //string - название команды
}

//функция которая будет реагировать на команды
//Добавление источника
// вывод всех источников
// удаление источников
type ViewFunc func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error //Update - любой event от телеграмма при взаимодейств с ботом


func New(api *tgbotapi.BotAPI) *Bot {
	return &Bot {
		api: api,
	}
}

//метод для регистрации view комады
func (b *Bot) RegisterCmdView(cmd string, view ViewFunc) {
	if b.cmdViews == nil {
		b.cmdViews = make(map[string]ViewFunc)
	}

	b.cmdViews[cmd] = view
}

//метод для обновления и роутить команды на view
func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	//перехватим паники
	defer func() {
		if p := recover(); p != nil {
			log.Printf("[ERROR] panic recovered: %v\n%s", p, string(debug.Stack())) //debug.Stack() - покажет цепочку которая вызвала панику
		}
	}()

	var view ViewFunc
	
	//если сообщение не содержит команды - выходим
	if update.Message == nil || !update.Message.IsCommand() {
		return
	}

	cmd := update.Message.Command() //вытаскиваем команду

	cmdView, ok := b.cmdViews[cmd] //пробуем достать view

	if !ok {
		return
	}

	view = cmdView

	//вызываем команду
	if err := view(ctx, b.api, update); err != nil {
		log.Printf("[ERROR] failed to execute view: %v", err)

		//пробуем отправить пользователю сообщение что произошла ошибка
		if _, err := b.api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Internal error")); err != nil { 
			log.Printf("[ERROR] failed to send error message: %v", err)
		}
	}
}

func (b *Bot) Run(ctx context.Context) error {

	u := tgbotapi.NewUpdate(0) //получаем канал в который будут писаться сообщения
	u.Timeout = 60 //таймаут в 60 сек

	//получаем канал
	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case update := <-updates:
			updateCtx, updateCancel := context.WithTimeout(ctx, 5 * time.Second)
			b.handleUpdate(updateCtx, update)
			updateCancel()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}