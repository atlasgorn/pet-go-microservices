package core

import (
	"context"
)

type Searcher interface {
	Search(ctx context.Context, limit int, phrase string) ([]PbComic, error)
	ISearch(ctx context.Context, limit int, phrase string) ([]PbComic, error)
}

type DB interface {
	SearchComics(ctx context.Context, limit int, phrase []string) ([]PbComic, error)
	GetAllComics(ctx context.Context) ([]Comics, error)
}

type Words interface {
	Norm(ctx context.Context, phrase string) ([]string, error)
}

type Indexer interface {
	ISearchComics(ctx context.Context, limit int, phrase []string) ([]PbComic, error)
	Build(ctx context.Context) error
}
