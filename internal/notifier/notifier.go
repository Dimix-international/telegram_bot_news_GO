package notifier

import (
	"app/internal/botkit/markup"
	"app/internal/model"
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-shiori/go-readability"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ArticleProvider interface {
	AllNotPosted(ctx context.Context, since time.Time, limit uint64) ([]model.Article, error)
	MarkPosted(ctx context.Context, id int64) error//отметка что статья опубликова
}

type Summarizer interface {
	Summarize(ctx context.Context, text string) (string, error)
}

type Notifier struct {
	articles         ArticleProvider
	summarizer Summarizer
	sendInterval time.Duration
	bot *tgbotapi.BotAPI
	lookupTimeWindow time.Duration //промежуток время назад для проверки новых статей
	channelId int64 //id канала куда постить будем статьи
}

func New(
	articleProvider ArticleProvider,
	summarizer Summarizer,
	sendInterval time.Duration,
	bot *tgbotapi.BotAPI,
	lookupTimeWindow time.Duration,
	channelId int64,
) *Notifier {
	return &Notifier {
		articles: articleProvider,
		summarizer: summarizer,
		sendInterval: sendInterval,
		bot: bot,
		lookupTimeWindow: lookupTimeWindow,
		channelId: channelId,
	}
}

func (n *Notifier) Start(ctx context.Context) error {
	ticker := time.NewTicker(n.sendInterval)
	defer ticker.Stop()

	if err := n.SelectAndSendArticle(ctx); err != nil {
		return err
	}

	for {
		select {
		case <-ticker.C:
			if err := n.SelectAndSendArticle(ctx); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (n *Notifier) SelectAndSendArticle(ctx context.Context) error {
	//будем постить одну статью
	toOneArticles, err := n.articles.AllNotPosted(ctx, time.Now().Add(-n.lookupTimeWindow), 1)

	if err != nil {
		return err
	}

	if len(toOneArticles) == 0 {
		return nil
	}

	article := toOneArticles[0]

	summary, err := n.extractSummary(ctx, article)

	if err != nil {
		return err
	}

	if err := n.sendArticle(article, summary); err != nil {
		return err
	}

	return n.articles.MarkPosted(ctx, article.ID)
}

func (n *Notifier) extractSummary(ctx context.Context, article model.Article) (string, error) {
	//проверим summary, если есть но его используем, если нету то по link получим html code страницы и получим summary через gpt
	var r io.Reader //читать будем из него html

	if article.Summary != "" {
		r = strings.NewReader(article.Summary)
	} else {
		resp, err := http.Get(article.Link)

		if err != nil {
			return "", err
		}

		defer resp.Body.Close()

		r = resp.Body
	}

	//преобразуем reader в документ
	doc, err := readability.FromReader(r, nil)

	if err != nil {
		return "", err
	}

		//readability проблема что она создает много пустых строк после убирания тегов

	summary, err := n.summarizer.Summarize(ctx, cleanText(doc.TextContent))

	
	if err != nil {
		return "", err
	}

	return "\n\n" + summary, nil
}

var redundantNewLines = regexp.MustCompile(`\n{3,}`) //регулярке удовлетворит для пустых строк от 3 штук и более

func cleanText(text string) string {
	return redundantNewLines.ReplaceAllString(text, "\n") //заменяем на одну пустую строку
}

func (n *Notifier) sendArticle(article model.Article, summary string) error {
	const msgFormat = "*%s*%s\n\n%s" //шаблон сообщения - загловок, summary и ссылка на статью

	msg := tgbotapi.NewMessage(n.channelId, fmt.Sprintf(
		msgFormat,
		markup.EscapeForMarkdown(article.Title),
		markup.EscapeForMarkdown(summary),
		markup.EscapeForMarkdown(article.Link),
	))
	msg.ParseMode = "MarkdownV2" //парсим как Markdown сообщения

	_, err := n.bot.Send(msg)
	if err != nil {
		return err
	}


	return nil
}