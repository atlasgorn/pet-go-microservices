package rest

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"yadro.com/course/api/core"
)

func NewPingHandler(log *slog.Logger, pingers map[string]core.Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		replies := make(map[string]string)
		for name, pinger := range pingers {
			if err := pinger.Ping(r.Context()); err != nil {
				replies[name] = "unavailable"
			} else {
				replies[name] = "ok"
			}
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]map[string]string{"replies": replies}); err != nil {
			log.Error("cannot encode ping response", "error", err)
		}
	}
}

func NewWordsHandler(log *slog.Logger, normalizer core.Normalizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		phrase := r.URL.Query().Get("phrase")
		if phrase == "" {
			http.Error(w, "phrase parameter is required", http.StatusBadRequest)
			return
		}

		words, err := normalizer.Norm(r.Context(), phrase)
		if err != nil {
			if errors.Is(err, core.ErrPhraseTooLarge) {
				http.Error(w, core.ErrPhraseTooLarge.Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{
			"words": words,
			"total": len(words),
		}); err != nil {
			log.Error("cannot encode words response", "error", err)
		}
	}
}
