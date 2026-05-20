package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"stathead/internal/model"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type PlayerStorer interface {
	Search(ctx context.Context, q string) ([]model.Player, error)
	Seasons(ctx context.Context, playerID, seasonType string) ([]model.SeasonLine, error)
	GameLogs(ctx context.Context, params model.GameLogParams) ([]model.GameLog, error)
	Compare(ctx context.Context, playerIDs []string, season, seasonType string) ([]model.StatLine, error)
	Leaders(ctx context.Context, stat, season, seasonType string, limit int) ([]model.LeaderRow, error)
}

type PlayerHandler struct {
	store PlayerStorer
}

func NewPlayerHandler(store PlayerStorer) *PlayerHandler {
	return &PlayerHandler{store: store}
}

func (h *PlayerHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeError(w, http.StatusBadRequest, "q param required")
		return
	}
	players, err := h.store.Search(r.Context(), q)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, players)
}

func (h *PlayerHandler) Seasons(w http.ResponseWriter, r *http.Request) {
	playerID   := chi.URLParam(r, "playerID")
	seasonType := defaultStr(r.URL.Query().Get("season_type"), "regular")

	seasons, err := h.store.Seasons(r.Context(), playerID, seasonType)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, seasons)
}

func (h *PlayerHandler) GameLogs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	params := model.GameLogParams{
		PlayerID:   chi.URLParam(r, "playerID"),
		Season:     q.Get("season"),
		SeasonType: defaultStr(q.Get("season_type"), "regular"),
	}

	if v := q.Get("fg_pct_lt"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			params.FgPctLt = &f
		}
	}
	if v := q.Get("min_fga"); v != "" {
		params.MinFGA, _ = strconv.Atoi(v)
	}

	logs, err := h.store.GameLogs(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"count": len(logs),
		"logs":  logs,
	})
}

func (h *PlayerHandler) Compare(w http.ResponseWriter, r *http.Request) {
	ids        := r.URL.Query()["player_id"]
	season     := r.URL.Query().Get("season")
	seasonType := defaultStr(r.URL.Query().Get("season_type"), "regular")

	if len(ids) < 2 {
		writeError(w, http.StatusBadRequest, "provide at least 2 player_id params")
		return
	}

	results, err := h.store.Compare(r.Context(), ids, season, seasonType)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, results)
}

func (h *PlayerHandler) Leaders(w http.ResponseWriter, r *http.Request) {
	q          := r.URL.Query()
	stat       := q.Get("stat")
	season     := q.Get("season")
	seasonType := defaultStr(q.Get("season_type"), "regular")
	limit, _   := strconv.Atoi(defaultStr(q.Get("limit"), "10"))
	if limit > 50 { limit = 50 }

	leaders, err := h.store.Leaders(r.Context(), stat, season, seasonType, limit)
	if errors.Is(err, model.ErrInvalidStat) {
		writeError(w, http.StatusBadRequest, "invalid stat column")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"stat":    stat,
		"season":  season,
		"leaders": leaders,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func defaultStr(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}
