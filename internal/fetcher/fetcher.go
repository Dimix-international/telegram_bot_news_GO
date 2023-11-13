package fetcher

import (
	"app/internal/model"
	"app/internal/source"
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/tomakado/containers/set"
)

type ArticlesStorage interface {
	Store(ctx context.Context, article model.Article) error
}

type SourceProvider interface {
	Sources(ctx context.Context) ([]model.Source, error)
}

type Source interface {
	ID() int64
	Name() string
	Fetch(ctx context.Context) ([]model.Item, error)
}

type Fetcher struct {
	articles ArticlesStorage
	sources SourceProvider

	fethcInterval time.Duration
	filterKeyword []string
}

//типа конструктор
func New(
	articlesStorage ArticlesStorage,
	sourceProvider SourceProvider,
	fetchIinterval time.Duration,
	filterKeyword []string,
) *Fetcher {
	return &Fetcher{
		articles: articlesStorage,
		sources: sourceProvider,
		fethcInterval: fetchIinterval,
		filterKeyword: filterKeyword,
	}
}

//метод worker который по интервалу забирает статьи
func (f *Fetcher) Start(ctx context.Context) error {
	ticker := time.NewTicker(f.fethcInterval)
	defer ticker.Stop()

	if err := f.Fetch(ctx); err != nil {
		return err
	}

	for {
		select {
		case <- ctx.Done():
			//контекст завершен или отменен
			return ctx.Err()
		case <- ticker.C:
			//прошел интервал
			if err := f.Fetch(ctx); err != nil {
				return err
			}
		}
	}
}

func (f *Fetcher) Fetch(ctx context.Context) error {
	sources, err := f.sources.Sources(ctx)
	//sources - список моделей источника
	if err != nil {
		return err
	}

	//делаем wait group т.к. источников много и они не влиял
	var wg sync.WaitGroup

	for _, sourceModel := range sources {
		wg.Add(1)

		rssSource := source.NewRSSSourceFromModel(sourceModel)

		//передаем интерфейс источника
		go func(source Source) {
			defer wg.Done()

			items, err := source.Fetch(ctx)

			if err != nil {
				log.Printf("[ERROR] failed to fetch items from source %q: %v", source.Name(), err)
				return
			}

			//сохраняем в бд

			if err := f.processItems(ctx, source, items); err != nil {
				log.Printf("[ERROR] failed to process items from source %q: %v", source.Name(), err)
				return
			}


		}(rssSource)
	}

	wg.Wait()
	return nil
}

func (f *Fetcher) processItems(ctx context.Context,source Source, items []model.Item) error {

	for _, item := range items {
		item.Date = item.Date.UTC()

		if f.itemShouldBeSkipped(item) {
			continue
		}

		//сохраняем
		if err := f.articles.Store(ctx, model.Article{
			SourceID: source.ID(),
			Title: item.Title,
			Link: item.Link,
			Summary: item.Summary,
			PublishedAt: item.Date,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (f *Fetcher) itemShouldBeSkipped(item model.Item) bool {
	categoriesSet := set.New(item.Categories...)

	for _, keyword := range f.filterKeyword {
		
		titleContainsKeyword := strings.Contains(strings.ToLower(item.Title), keyword)

		if categoriesSet.Contains(keyword) || titleContainsKeyword{
			return true
		}
	}

	return false
}