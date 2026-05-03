package index

import (
	"context"
	"log/slog"
	"sort"
	"sync"
	"time"

	"yadro.com/course/search/core"
)

type Indexer struct {
	log   *slog.Logger
	db    core.DB
	ttl   time.Duration
	mu    sync.RWMutex
	index map[string][]int
	urls  []string
}

func New(log *slog.Logger, db core.DB, ttl time.Duration) (*Indexer, error) {
	return &Indexer{
		log:   log,
		db:    db,
		ttl:   ttl,
		index: make(map[string][]int),
		urls:  make([]string, 0),
	}, nil
}

func (i *Indexer) ISearchComics(ctx context.Context, limit int, phrase []string) ([]core.PbComic, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	matches := make(map[int]int)
	for _, word := range phrase {
		ids, ok := i.index[word]
		if !ok {
			continue
		}
		for _, id := range ids {
			matches[id]++
		}
	}

	type result struct {
		id         int
		matchCount int
	}
	var results []result
	for id, count := range matches {
		results = append(results, result{id: id, matchCount: count})
	}

	sort.Slice(results, func(a, b int) bool {
		return results[a].matchCount > results[b].matchCount
	})

	out := make([]core.PbComic, 0, min(limit, len(results)))
	for _, r := range results {
		if len(out) >= limit {
			break
		}
		if r.id < len(i.urls) {
			out = append(out, core.PbComic{
				ID:  r.id,
				URL: i.urls[r.id],
			})
		}
	}

	return out, nil
}

func (i *Indexer) build(ctx context.Context) error {
	comics, err := i.db.GetAllComics(ctx)
	if err != nil {
		return err
	}

	newIndex := make(map[string][]int)
	newURLs := make([]string, 0)

	for _, c := range comics {
		if c.ID >= len(newURLs) {
			newURLs = append(newURLs, make([]string, c.ID-len(newURLs)+1)...)
		}
		newURLs[c.ID] = c.URL

		for _, word := range c.Words {
			newIndex[word] = append(newIndex[word], c.ID)
		}
	}

	for word := range newIndex {
		ids := newIndex[word]
		sort.Ints(ids)
		newIndex[word] = ids
	}

	i.log.Debug("build successfully", "words", len(newIndex))
	i.mu.Lock()
	i.index = newIndex
	i.urls = newURLs
	i.mu.Unlock()

	return nil
}

func (i *Indexer) Run(ctx context.Context) {
	i.log.InfoContext(ctx, "running indexer")
	if err := i.build(ctx); err != nil {
		i.log.ErrorContext(ctx, "error building index", "error", err)
	} else {
		i.log.InfoContext(ctx, "index build successfully")
	}

	ticker := time.NewTicker(i.ttl)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			i.log.InfoContext(ctx, "running indexer")
			if err := i.build(ctx); err != nil {
				i.log.ErrorContext(ctx, "error building index", "error", err)
			} else {
				i.log.InfoContext(ctx, "index build successfully")
			}
		}
	}
}
