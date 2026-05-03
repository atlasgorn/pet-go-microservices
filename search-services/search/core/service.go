package core

import (
	"context"
	"log/slog"
)

type Service struct {
	log   *slog.Logger
	db    DB
	words Words
	index Indexer
}

func NewService(
	log *slog.Logger, db DB, words Words, index Indexer,
) (*Service, error) {
	return &Service{
		log:   log,
		db:    db,
		words: words,
		index: index,
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

func (s *Service) ISearch(ctx context.Context, limit int, phrase string) ([]PbComic, error) {
	normPhrase, err := s.words.Norm(ctx, phrase)
	if err != nil {
		s.log.Error("failed to norm phrase", "phrase", phrase, "error", err)
		return nil, err
	}

	comics, err := s.index.ISearchComics(ctx, limit, normPhrase)
	if err != nil {
		s.log.Error("failed to search comics", "phrase", phrase, "error", err)
		return nil, err
	}

	return comics, nil
}
