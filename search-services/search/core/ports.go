package core

import (
	"context"
)

type Searcher interface {
	Search(ctx context.Context, limit int, phrase string) ([]PbComic, error)
}

type DB interface {
	SearchComics(ctx context.Context, limit int, phrase []string) ([]PbComic, error)
}

type Words interface {
	Norm(ctx context.Context, phrase string) ([]string, error)
}
