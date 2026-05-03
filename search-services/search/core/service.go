package core

import (
	"context"
	"log/slog"
)

type Service struct {
	log   *slog.Logger
	db    DB
	words Words
}

func NewService(
	log *slog.Logger, db DB, words Words,
) (*Service, error) {
	return &Service{
		log:   log,
		db:    db,
		words: words,
	}, nil
}

func (s *Service) Search(ctx context.Context, limit int, phrase string) ([]PbComic, error) {
	normPhrase, err := s.words.Norm(ctx, phrase)
	if err != nil {
		s.log.Error("failed to norm phrase", "phrase", phrase, "error", err)
		return nil, err
	}

	comics, err := s.db.SearchComics(ctx, limit, normPhrase)
	if err != nil {
		s.log.Error("failed to search comics", "phrase", phrase, "error", err)
		return nil, err
	}

	return comics, nil
}
