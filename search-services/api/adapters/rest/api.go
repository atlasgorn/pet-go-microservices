package rest

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"yadro.com/course/api/core"
)

func NewMetricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

type Authenticator interface {
	Login(user, password string) (string, error)
}

func NewLoginHandler(log *slog.Logger, auth Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

type pingResponse struct {
	Replies map[string]string `json:"replies"`
}

func NewPingHandler(log *slog.Logger, pingers map[string]core.Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reply := pingResponse{
			Replies: make(map[string]string),
		}
		for name, pinger := range pingers {
			if err := pinger.Ping(r.Context()); err != nil {
				reply.Replies[name] = "unavailable"
			} else {
				reply.Replies[name] = "ok"
			}
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			log.Error("cannot encode ping response", "error", err)
		}
	}
}

func NewUpdateHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log.Debug("received update request")

		err := updater.Update(ctx)
		if err != nil {
			if errors.Is(err, core.ErrAlreadyExists) {
				w.WriteHeader(http.StatusAccepted)
				return
			}
			log.Error("update failed", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

type updateStats struct {
	WordsTotal    int `json:"words_total"`
	WordsUnique   int `json:"words_unique"`
	ComicsFetched int `json:"comics_fetched"`
	ComicsTotal   int `json:"comics_total"`
}

func NewUpdateStatsHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		stats, err := updater.Stats(ctx)
		if err != nil {
			log.Error("cannot get stats", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		replyStats := updateStats{
			WordsTotal:    stats.WordsTotal,
			WordsUnique:   stats.WordsUnique,
			ComicsFetched: stats.ComicsFetched,
			ComicsTotal:   stats.ComicsTotal,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(replyStats); err != nil {
			log.Error("cannot encode response", "error", err)
		}
	}
}

type statusResponse struct {
	Status core.UpdateStatus `json:"status"`
}

func NewUpdateStatusHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		status, err := updater.Status(ctx)
		if err != nil {
			log.Error("cannot get update status", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		response := statusResponse{Status: status}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Error("cannot encode response", "error", err)
		}
	}
}

func NewDropHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if err := updater.Drop(ctx); err != nil {
			log.Error("cannot drop database", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

type searchResponse struct {
	Comics []pbComic `json:"comics"`
	Total  int       `json:"total"`
}

type pbComic struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

const defaultLimit = 10

func NewSearchHandler(log *slog.Logger, searcher core.Searcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil {
			limit = defaultLimit
		}

		if limit <= 0 {
			http.Error(w, "invalid limit parameter", http.StatusBadRequest)
			return
		}

		phrase := r.URL.Query().Get("phrase")
		if phrase == "" {
			http.Error(w, "missing phrase parameter", http.StatusBadRequest)
			return
		}

		comics, err := searcher.Search(ctx, limit, phrase)
		if err != nil {
			log.Error("cannot search", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		replyComics := make([]pbComic, len(comics))
		for i, comic := range comics {
			replyComics[i] = pbComic{ID: comic.ID, URL: comic.URL}
		}
		response := searchResponse{
			Comics: replyComics,
			Total:  len(comics),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Error("cannot encode response", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func NewISearchHandler(log *slog.Logger, searcher core.Searcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil {
			limit = defaultLimit
		}

		if limit <= 0 {
			http.Error(w, "invalid limit parameter", http.StatusBadRequest)
			return
		}

		phrase := r.URL.Query().Get("phrase")
		if phrase == "" {
			http.Error(w, "missing phrase parameter", http.StatusBadRequest)
			return
		}

		comics, err := searcher.ISearch(ctx, limit, phrase)
		if err != nil {
			log.Error("cannot search", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		replyComics := make([]pbComic, len(comics))
		for i, comic := range comics {
			replyComics[i] = pbComic{ID: comic.ID, URL: comic.URL}
		}
		response := searchResponse{
			Comics: replyComics,
			Total:  len(comics),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Error("cannot encode response", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
