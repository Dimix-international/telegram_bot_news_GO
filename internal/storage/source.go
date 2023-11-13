package storage

import (
	"app/internal/model"
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
)

type SourcePostgresStorage struct {
	db *sqlx.DB
}

//для мапинг структур
type dbSource struct {
	ID int64 `db:"id"`
	Name string `db:"name"`
	FeedUrl string `db:"feed_url"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

func NewSourceStorage(db *sqlx.DB) *SourcePostgresStorage {
	return &SourcePostgresStorage{db: db}
} 

func (s *SourcePostgresStorage) Sources(ctx context.Context) ([]model.Source, error) {
	conn, err := s.db.Connx(ctx)

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	var sourses []dbSource

	if err := conn.SelectContext(ctx, &sourses, `SELECT * FROM sources`); err != nil {
		return nil, err
	}
	
	return lo.Map(sourses, func(source dbSource, _ int) model.Source {
		return model.Source(source)
	}), nil
}

func (s *SourcePostgresStorage) SourceById(ctx context.Context, id int64) (*model.Source, error) {
	conn, err := s.db.Connx(ctx)

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	var sourse dbSource

	if err := conn.GetContext(ctx, &sourse, `SELECT * FROM sources WHERE id = $1`, id); err != nil {
		return nil, err
	}

	return (*model.Source)(&sourse), nil //кастуем из sourse в model.Source
}

func (s *SourcePostgresStorage) Add(ctx context.Context, source model.Source) (int64, error) {
	conn, err := s.db.Connx(ctx)

	if err != nil {
		return 0, err
	}

	defer conn.Close()

	var id int64

	row := conn.QueryRowContext(
		ctx,
		`INSERT INTO sources(name, feed_url, created_at) VALUES($1, $2, $3) RETURNING id`,
		source.Name,
		source.FeedUrl,
		source.CreatedAt,
	)

	if err := row.Err(); err != nil {
		return 0, err
	}

	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (s *SourcePostgresStorage) Delete(ctx context.Context, id int64) error {
	conn, err := s.db.Connx(ctx)

	if err != nil {
		return err
	}

	defer conn.Close()

	if _, err := conn.ExecContext(ctx, `DELETE FROM sources WHERE id=$1`, id); err != nil {
		return err
	}

	return nil
}