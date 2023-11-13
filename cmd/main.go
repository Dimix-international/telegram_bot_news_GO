package main

import (
	"app/internal/bot"
	"app/internal/bot/middleware"
	"app/internal/botkit"
	"app/internal/config"
	"app/internal/fetcher"
	"app/internal/notifier"
	"app/internal/storage"
	"app/internal/summary"
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	//создаем бота api
	botAPI, err := tgbotapi.NewBotAPI(config.Get().TelegramBotToken)

	if err != nil {
		log.Printf("failed to create bot: %v", err)
		return
	}

	//подключаемся к бд

	db, err := sqlx.Connect("postgres", config.Get().DatabaseDSN)

	if err != nil {
		log.Printf("failed to connect to db: %v", err)
		return
	}

	defer db.Close()
	//инициализируем зависимости
	var (
		articleStorage = storage.NewArticleStorage(db)
		sourceStorage = storage.NewSourceStorage(db)
		channelID = config.Get().TelegramChannelID
		fetcher = fetcher.New( //забираем статьи из источников
			articleStorage,
			sourceStorage,
			config.Get().FetchInterval,
			config.Get().FilterKeywords,
		)
		notifier = notifier.New(
			articleStorage,
			summary.NewOpenAISummarizer(config.Get().OpenAIKey, config.Get().OpenAIPrompt),
			config.Get().NotificationInterval, //интервал отправки сообщений
			botAPI,
			2 * config.Get().FetchInterval, //интервал заглядывать назад для проверки новых статей
			channelID,
		)
	)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	//инициализируем бота
	newsBot := botkit.New(botAPI)
	//регестриуем view для команды start
	newsBot.RegisterCmdView("start", bot.ViewCmdStart())
	newsBot.RegisterCmdView("addsource", middleware.AddminOnly(channelID, bot.ViewCmdAddSource(sourceStorage)))
	newsBot.RegisterCmdView("listsources", middleware.AddminOnly(channelID, bot.ViewCmdListSources(sourceStorage)))
	newsBot.RegisterCmdView("deletesource", middleware.AddminOnly(channelID, bot.ViewCmdDeleteSource(sourceStorage)))
	newsBot.RegisterCmdView("getsource", middleware.AddminOnly(channelID, bot.VewCmdGetSource(sourceStorage)))

	//стартуем fetch
	go func(ctx context.Context) {
		if err := fetcher.Start(ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Printf("[ERROR] failed to run fetcher: %v", err)
				return
			}

			log.Printf("[INFO] fetcher stopped")
		}
	}(ctx)

	//cтарт notifier
	go func(ctx context.Context) {
		if err := notifier.Start(ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Printf("[ERROR] failed to run notifier: %v", err)
				return
			}

			log.Printf("[INFO] notifier stopped")
		}
	}(ctx)

	//запускаем бота
	if err := newsBot.Run(ctx); err != nil {
		log.Printf("[ERROR] failed to run botkit: %v", err)
	}
}