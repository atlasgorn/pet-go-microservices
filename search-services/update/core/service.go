package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type Service struct {
	log         *slog.Logger
	db          DB
	xkcd        XKCD
	words       Words
	concurrency int
	status      ServiceStatus
	lock        sync.Mutex
}

func NewService(
	log *slog.Logger, db DB, xkcd XKCD, words Words, concurrency int,
) (*Service, error) {
	if concurrency < 1 {
		return nil, fmt.Errorf("wrong concurrency specified: %d", concurrency)
	}
	return &Service{
		log:         log,
		db:          db,
		xkcd:        xkcd,
		words:       words,
		concurrency: concurrency,
		status:      StatusIdle,
	}, nil
}

func (s *Service) Update(ctx context.Context) error {
	if ok := s.lock.TryLock(); !ok {
		return ErrAlreadyExists
	}
	defer s.lock.Unlock()

	s.status = StatusRunning
	defer func() {
		s.status = StatusIdle
	}()

	lastID, err := s.xkcd.LastID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get last comic ID: %w", err)
	}

	existingIDs, err := s.db.IDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get existing comic IDs: %w", err)
	}

	existingMap := make(map[int]bool, len(existingIDs))
	for _, id := range existingIDs {
		existingMap[id] = true
	}

	existingMap[404] = true // XKCD comic 404 does not exist, so we mark it as already fetched

	var idsToFetch []int
	for id := 1; id <= lastID; id++ {
		if !existingMap[id] {
			idsToFetch = append(idsToFetch, id)
		}
	}

	if len(idsToFetch) == 0 {
		s.log.InfoContext(ctx, "no new comics to fetch")
		return nil
	}

	s.log.InfoContext(ctx, "starting update", "total", len(idsToFetch))

	var mu sync.Mutex
	var errs []error

	g, ctx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(int64(s.concurrency))

	for _, id := range idsToFetch {
		g.Go(func() error {
			if err := sem.Acquire(ctx, 1); err != nil {
				return err
			}
			defer sem.Release(1)

			info, err := s.xkcd.Get(ctx, id)
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("failed to fetch comic %d: %w", id, err))
				mu.Unlock()
				return nil
			}

			text := info.Title + " " + info.Description
			words, err := s.words.Norm(ctx, text)
			if err != nil {
				if errors.Is(err, ErrPhraseTooLarge) {
					s.log.WarnContext(ctx, "comic discription is too large, using only title", "comic_id", info.ID)
					words, err = s.words.Norm(ctx, info.Title)
					if err != nil {
						mu.Lock()
						errs = append(errs, fmt.Errorf("failed to normalize title for comic %d: %w", id, err))
						mu.Unlock()
					}

				} else {
					s.log.ErrorContext(ctx, "failed to normalize words", "comic_id", info.ID, "error", err)
					mu.Lock()
					errs = append(errs, fmt.Errorf("failed to normalize title for comic %d: %w", id, err))
					mu.Unlock()
				}
			}

			if words == nil {
				s.log.WarnContext(ctx, "comic title consist of only stop words", "comic_id", info.ID)
				words = []string{""}
			}

			comic := Comics{
				ID:    id,
				URL:   info.URL,
				Words: words,
			}

			if err := s.db.Add(ctx, comic); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("failed to add comic %d to DB: %w", id, err))
				mu.Unlock()
				return nil
			}

			s.log.DebugContext(ctx, "comic fetched and saved", "id", id, "words", len(words))
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	if len(errs) > 0 {
		for _, err := range errs {
			s.log.ErrorContext(ctx, "error during update", "error", err)
		}
		return fmt.Errorf("update completed with %d errors: %w", len(errs), errs[0])
	}

	s.log.InfoContext(ctx, "update completed successfully")
	return nil
}

func (s *Service) Stats(ctx context.Context) (ServiceStats, error) {
	dbstats, err := s.db.Stats(ctx)
	if err != nil {
		return ServiceStats{}, fmt.Errorf("failed to get DB stats: %w", err)
	}
	total, err := s.xkcd.LastID(ctx)
	if err != nil {
		return ServiceStats{}, fmt.Errorf("failed to get total comics count: %w", err)
	}
	return ServiceStats{DBStats: dbstats, ComicsTotal: total - 1}, nil // comic 404 does not exist, so we subtract it from total count
}

func (s *Service) Status(ctx context.Context) ServiceStatus {
	return s.status
}

func (s *Service) Drop(ctx context.Context) error {
	if ok := s.lock.TryLock(); !ok {
		s.log.Error("service already runs update or drop")
		return ErrAlreadyExists
	}
	defer s.lock.Unlock()

	return s.db.Drop(ctx)
}
