package store

import (
	"context"
	"fmt"
	"log"
	"stathead/internal/model"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PlayerStore struct {
	db *pgxpool.Pool
}

func NewPlayerStore(db *pgxpool.Pool) *PlayerStore {
	return &PlayerStore{db: db}
}

func (s *PlayerStore) Search(ctx context.Context, q string) ([]model.Player, error) {
	if len(strings.TrimSpace(q)) < 2 {
		return nil, fmt.Errorf("query too short")
	}
	rows, err := s.db.Query(ctx, `
		SELECT player_id, full_name
		FROM players
		WHERE unaccent(lower(full_name)) LIKE unaccent(lower($1))
		ORDER BY full_name
		LIMIT 10
	`, "%"+q+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []model.Player
	for rows.Next() {
		var p model.Player
		if err := rows.Scan(&p.PlayerID, &p.FullName); err != nil {
			return nil, err
		}
		players = append(players, p)
	}
	return players, rows.Err()
}

func (s *PlayerStore) Seasons(ctx context.Context, playerID, seasonType string) ([]model.SeasonLine, error) {
	rows, err := s.db.Query(ctx, `
		SELECT season, team, pos, age, games, games_started, minutes_pg,
		       pts, reb, ast, stl, blk, tov,
		       fgm, fga, fg_pct, fg3m, fg3a, fg3_pct,
		       ftm, fta, ft_pct, efg_pct,
		       per, ts_pct, usg_pct, bpm, vorp, ws
		FROM player_season_stats
		WHERE player_id = $1 AND season_type = $2
		ORDER BY season ASC
	`, playerID, seasonType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seasons []model.SeasonLine
	for rows.Next() {
		var sl model.SeasonLine
		if err := rows.Scan(
			&sl.Season, &sl.Team, &sl.Pos, &sl.Age, &sl.Games, &sl.GamesStarted, &sl.MinutesPG,
			&sl.Pts, &sl.Reb, &sl.Ast, &sl.Stl, &sl.Blk, &sl.Tov,
			&sl.Fgm, &sl.Fga, &sl.FgPct, &sl.Fg3m, &sl.Fg3a, &sl.Fg3Pct,
			&sl.Ftm, &sl.Fta, &sl.FtPct, &sl.EfgPct,
			&sl.Per, &sl.TsPct, &sl.UsgPct, &sl.Bpm, &sl.Vorp, &sl.Ws,
		); err != nil {
			return nil, err
		}
		seasons = append(seasons, sl)
	}
	return seasons, rows.Err()
}

func (s *PlayerStore) GameLogs(ctx context.Context, params model.GameLogParams) ([]model.GameLog, error) {
	query := `
		SELECT game_id, game_date, matchup, wl, min,
		       pts, reb, ast, stl, blk, tov,
		       fgm, fga, fg_pct, fg3m, fg3a, fg3_pct,
		       ftm, fta, ft_pct, plus_minus
		FROM player_game_logs
		WHERE player_id_nba = (
		    SELECT player_id_nba FROM player_id_map
		    WHERE player_id_br = $1
		)
		AND season_type = $2
	`
	args := []any{params.PlayerBRID, params.SeasonType}
	i := 3

	if params.SeasonYear != 0 {
		query += " AND season = $" + itoa(i)
		args = append(args, fmt.Sprintf("%d-%02d", params.SeasonYear-1, params.SeasonYear%100))
		i++
	}
	if params.FgPctLt != nil {
		query += " AND fg_pct < $" + itoa(i)
		args = append(args, *params.FgPctLt)
		i++
	}
	if params.MinFGA != nil {
		query += " AND fga >= $" + itoa(i)
		args = append(args, *params.MinFGA)
		i++
	}
	query += " ORDER BY game_date ASC"

	log.Printf("query: %s", query)
	log.Printf("args: %v", args)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []model.GameLog
	for rows.Next() {
		var g model.GameLog
		if err := rows.Scan(
			&g.GameID, &g.GameDate, &g.Matchup, &g.WL, &g.Min,
			&g.Pts, &g.Reb, &g.Ast, &g.Stl, &g.Blk, &g.Tov,
			&g.Fgm, &g.Fga, &g.FgPct, &g.Fg3m, &g.Fg3a, &g.Fg3Pct,
			&g.Ftm, &g.Fta, &g.FtPct, &g.PlusMinus,
		); err != nil {
			return nil, err
		}
		logs = append(logs, g)
	}
	log.Printf("GameLogs returned %d rows", len(logs))
	return logs, rows.Err()
}

func (s *PlayerStore) Compare(ctx context.Context, playerIDs []string, season, seasonType string) ([]model.StatLine, error) {
	query := `
		SELECT p.full_name, s.player_id, s.season, s.team,
		       s.pts, s.reb, s.ast, s.stl, s.blk,
		       s.fg_pct, s.fg3_pct, s.ts_pct, s.per, s.bpm, s.vorp, s.ws
		FROM player_season_stats s
		JOIN players p ON p.player_id = s.player_id
		WHERE s.player_id = ANY($1) AND s.season_type = $2
	`
	args := []any{playerIDs, seasonType}
	i := 3
	if season != "" {
		query += " AND s.season = $" + itoa(i)
		args = append(args, season)
		i++
	}
	query += " ORDER BY s.player_id, s.season"

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []model.StatLine
	for rows.Next() {
		var sl model.StatLine
		if err := rows.Scan(
			&sl.FullName, &sl.PlayerID, &sl.Season, &sl.Team,
			&sl.Pts, &sl.Reb, &sl.Ast, &sl.Stl, &sl.Blk,
			&sl.FgPct, &sl.Fg3Pct, &sl.TsPct, &sl.Per, &sl.Bpm, &sl.Vorp, &sl.Ws,
		); err != nil {
			return nil, err
		}
		results = append(results, sl)
	}
	return results, rows.Err()
}

var allowedStatCols = map[string]bool{
	"pts": true, "reb": true, "ast": true, "stl": true, "blk": true,
	"fg_pct": true, "fg3_pct": true, "ts_pct": true, "per": true,
	"bpm": true, "vorp": true, "ws": true, "usg_pct": true,
}

func (s *PlayerStore) Leaders(ctx context.Context, stat, season, seasonType string, limit int) ([]model.LeaderRow, error) {
	if !allowedStatCols[stat] {
		return nil, model.ErrInvalidStat
	}

	query := `
		SELECT p.full_name, s.player_id, s.season, s.team, s.` + stat + `
		FROM player_season_stats s
		JOIN players p ON p.player_id = s.player_id
		WHERE s.season_type = $1
		  AND s.` + stat + ` IS NOT NULL
		  AND s.games >= 58
	`
	args := []any{seasonType}
	i := 2

	if len(season) == 4 {
		y, err := strconv.Atoi(season)
		if err != nil {
			return nil, fmt.Errorf("invalid season year %q: %w", season, err)
		}
		season = fmt.Sprintf("%d-%02d", y-1, y%100)
	}
	if season != "" {
		query += " AND s.season = $" + itoa(i)
		args = append(args, season)
		i++
	}
	query += " ORDER BY s." + stat + " DESC LIMIT $" + itoa(i)
	args = append(args, limit)
	log.Printf("DEBUG Leaders | query: %s", query)
	log.Printf("DEBUG Leaders | args: %+v", args)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	leaders := []model.LeaderRow{}
	for rows.Next() {
		var l model.LeaderRow
		l.Stat = stat
		if err := rows.Scan(&l.FullName, &l.PlayerID, &l.Season, &l.Team, &l.Value); err != nil {
			return nil, err
		}
		leaders = append(leaders, l)
	}
	return leaders, rows.Err()
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
