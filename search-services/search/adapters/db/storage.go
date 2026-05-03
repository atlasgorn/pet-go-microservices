package db

import (
	"context"
	"log/slog"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"yadro.com/course/search/core"
)

type DB struct {
	log  *slog.Logger
	conn *sqlx.DB
}

func New(log *slog.Logger, address string) (*DB, error) {
	db, err := sqlx.Connect("pgx", address)
	if err != nil {
		log.Error("connection problem", "address", address, "error", err)
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	return &DB{
		log:  log,
		conn: db,
	}, nil
}

func (db *DB) SearchComics(ctx context.Context, limit int, phrase []string) ([]core.PbComic, error) {
	var comics []core.PbComic

	query := `
		SELECT id, url
		FROM comics
		WHERE words && $1::text[]
		ORDER BY array_length(array(SELECT unnest(words) INTERSECT SELECT unnest($1::text[])), 1) DESC,
				array_length(words, 1) ASC
		LIMIT $2
	`

	err := db.conn.SelectContext(ctx, &comics, query, pq.Array(phrase), limit)
	if err != nil {
		return nil, err
	}

	return comics, nil
}

func (db *DB) ISearchComics(ctx context.Context, limit int, phrase []string) ([]core.PbComic, error) {
	var comics []core.PbComic

	query := `
		SELECT c.id, c.url
		FROM search_index si
		CROSS JOIN LATERAL unnest(si.comic_ids) AS comic_id
		JOIN comics c ON c.id = comic_id
		WHERE si.word = ANY($1)
		GROUP BY c.id, c.url
		ORDER BY COUNT(DISTINCT si.word) DESC, array_length(c.words, 1) ASC
		LIMIT $2
	`

	err := db.conn.SelectContext(ctx, &comics, query, pq.Array(phrase), limit)
	if err != nil {
		return nil, err
	}

	return comics, nil
}

func (db *DB) Index(ctx context.Context) error {
	query := `
		TRUNCATE search_index;

		INSERT INTO search_index (word, comic_ids)
		SELECT
			word,
			array_agg(DISTINCT comic_id ORDER BY comic_id) AS comic_ids
		FROM (
			SELECT
				id AS comic_id,
				unnest(words) AS word
			FROM comics
		) AS flattened
		GROUP BY word;
	`

	_, err := db.conn.ExecContext(ctx, query)
	return err
}

type comics struct {
	ID    int
	URL   string
	Words pq.StringArray
}

func (db *DB) GetAllComics(ctx context.Context) ([]core.Comics, error) {
	var comics []comics
	query := `SELECT id, url, words FROM comics`
	if err := db.conn.SelectContext(ctx, &comics, query); err != nil {
		return nil, err
	}

	result := make([]core.Comics, len(comics))
	for i, c := range comics {
		result[i] = core.Comics{
			ID:    c.ID,
			URL:   c.URL,
			Words: []string(c.Words),
		}
	}
	return result, nil
}
