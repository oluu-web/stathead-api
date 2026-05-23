package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
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
	Compare(ctx context.Context, playerIDs, seasons []string, seasonType string) ([]model.StatLine, error)
	Leaders(ctx context.Context, stat, season, seasonType string, limit int) ([]model.LeaderRow, error)
	GameLogsVsTeam(ctx context.Context, params model.VsTeamParams) ([]model.GameLog, error)
	GameLogsVsTeamAllTime(ctx context.Context, params model.VsTeamAllTimeParams) ([]model.VsTeamSeason, error)
	HeadToHead(ctx context.Context, playerIDA, playerIDB, seasonType, season string) (*model.HeadToHead, error)
	DefenseStats(ctx context.Context, params model.DefenseParams) ([]model.DefenseStat, error)
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
	playerID := chi.URLParam(r, "playerID")
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
	seasonStr := q.Get("season")
	var seasonYear int
	if seasonStr != "" {
		var err error
		seasonYear, err = strconv.Atoi(seasonStr)
		if err != nil {
			http.Error(w, "season must be a 4-digit year e.g. ?season=2024", http.StatusBadRequest)
			return
		}
	}
	params := model.GameLogParams{
		PlayerBRID: chi.URLParam(r, "playerID"),
		SeasonYear: seasonYear,
		SeasonType: defaultStr(q.Get("season_type"), "regular"),
	}

	if v := q.Get("fg_pct_lt"); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			http.Error(w, "fg_pct must be a float", http.StatusBadRequest)
			return
		}
		params.FgPctLt = &f
	}
	if v := q.Get("min_fga"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			http.Error(w, "min_fga must be an integer", http.StatusBadRequest)
			return
		}
		params.MinFGA = &n
	}

	logs, err := h.store.GameLogs(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if logs == nil {
		logs = []model.GameLog{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"count": len(logs),
		"logs":  logs,
	})
}

func (h *PlayerHandler) Compare(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	ids := q["player_id"]
	seasons := q["season"]
	seasonType := defaultStr(q.Get("season_type"), "regular")

	if len(ids) < 1 {
		writeError(w, http.StatusBadRequest, "provide at least 1 player_id param")
		return
	}
	if len(ids) > 10 {
		writeError(w, http.StatusBadRequest, "maximum 10 player_ids")
		return
	}

	results, err := h.store.Compare(r.Context(), ids, seasons, seasonType)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if results == nil {
		results = []model.StatLine{}
	}

	resp := map[string]any{"rows": results}

	if len(ids) == 2 && ids[0] != ids[1] {
		h2hSeason := ""
		if len(seasons) == 1 {
			h2hSeason = seasons[0]
		}

		h2h, err := h.store.HeadToHead(r.Context(), ids[0], ids[1], seasonType, h2hSeason)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		resp["head_to_head"] = h2h
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *PlayerHandler) Leaders(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	stat := q.Get("stat")
	season := q.Get("season")
	seasonType := defaultStr(q.Get("season_type"), "regular")
	limit, err := strconv.Atoi(defaultStr(q.Get("limit"), "10"))
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	if stat == "" {
		writeError(w, http.StatusBadRequest, "stat param required")
		return
	}

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

func (h *PlayerHandler) VsTeam(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	seasonStr := q.Get("season")
	var seasonYear int
	if seasonStr != "" {
		var err error
		seasonYear, err = strconv.Atoi(seasonStr)
		if err != nil {
			http.Error(w, "season must be a 4-digit year e.g. ?season=2024", http.StatusBadRequest)
			return
		}
	}

	team := q.Get("team")
	if team == "" {
		http.Error(w, "team query param is required e.g. ?team=GSW", http.StatusBadRequest)
		return
	}

	params := model.VsTeamParams{
		PlayerBRID: chi.URLParam(r, "playerID"),
		Team:       team,
		SeasonYear: seasonYear,
		SeasonType: defaultStr(q.Get("season_type"), "regular"),
	}

	logs, err := h.store.GameLogsVsTeam(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if logs == nil {
		logs = []model.GameLog{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"player_id": params.PlayerBRID,
		"team":      params.Team,
		"season":    params.SeasonYear,
		"count":     len(logs),
		"logs":      logs,
	})
}

func (h *PlayerHandler) VsTeamAllTime(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	team := q.Get("team")
	if team == "" {
		http.Error(w, "team query param is required e.g. ?team=GSW", http.StatusBadRequest)
		return
	}

	params := model.VsTeamAllTimeParams{
		PlayerBRID: chi.URLParam(r, "playerID"),
		Team:       team,
		SeasonType: defaultStr(q.Get("season_type"), "regular"),
	}

	seasons, err := h.store.GameLogsVsTeamAllTime(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if seasons == nil {
		seasons = []model.VsTeamSeason{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"player_id": params.PlayerBRID,
		"team":      params.Team,
		"seasons":   len(seasons),
		"rows":      seasons,
	})
}

func (h *PlayerHandler) Defense(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	seasonStr := q.Get("season")
	var seasonYear int
	if seasonStr != "" {
		var err error
		seasonYear, err = strconv.Atoi(seasonStr)
		if err != nil {
			http.Error(w, "season must be a 4-digit year, e.g. ?season=2024", http.StatusBadRequest)
			return
		}
	}

	params := model.DefenseParams{
		PlayerBRID: chi.URLParam(r, "playerID"),
		SeasonYear: seasonYear, // 0 means all seasons
		SeasonType: defaultStr(q.Get("seasontype"), "regular"),
	}

	log.Printf("Defense params: %+v", params)

	stats, err := h.store.DefenseStats(r.Context(), params)
	if err != nil {
		log.Fatalln(err)
		writeError(w, http.StatusInternalServerError, "failed to fetch defense stats")
		return
	}
	if stats == nil {
		stats = []model.DefenseStat{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"player_id": params.PlayerBRID,
		"season":    strconv.Itoa(seasonYear),
		"count":     len(stats),
		"stats":     stats,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON encode error: %v", err)
	}
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
