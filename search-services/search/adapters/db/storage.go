package db

import (
	"context"
	"log/slog"

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
