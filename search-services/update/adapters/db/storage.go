package db

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"yadro.com/course/update/core"
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

func (db *DB) Add(ctx context.Context, comics core.Comics) error {
	_, err := db.conn.NamedExecContext(ctx, "INSERT INTO comics (id, url, words) VALUES (:id, :url, :words)", comics)
	return err
}

func (db *DB) Stats(ctx context.Context) (core.DBStats, error) {
	var stats core.DBStats

	err := db.conn.GetContext(ctx, &stats.ComicsFetched, "SELECT COUNT(*) FROM comics")
	if err != nil {
		return core.DBStats{}, fmt.Errorf("failed to count comics: %w", err)
	}

	var words []string
	err = db.conn.SelectContext(ctx, &words, "SELECT words FROM comics")
	if err != nil {
		return core.DBStats{}, fmt.Errorf("failed to get words: %w", err)
	}

	totalWords := 0
	uniqueWordsMap := make(map[string]bool)

	for _, wordString := range words {
		splitWords := strings.Fields(wordString)
		totalWords += len(splitWords)

		for _, w := range splitWords {
			uniqueWordsMap[w] = true
		}
	}

	stats.WordsTotal = totalWords
	stats.WordsUnique = len(uniqueWordsMap)

	return stats, nil
}

func (db *DB) IDs(ctx context.Context) ([]int, error) {
	var ids []int
	if err := db.conn.SelectContext(ctx, &ids, "SELECT id FROM comics"); err != nil {
		return nil, err
	}
	return ids, nil
}

func (db *DB) Drop(ctx context.Context) error {
	_, err := db.conn.ExecContext(ctx, "DELETE FROM comics")
	return err
}
