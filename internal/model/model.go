package model

import (
	"errors"
	"time"
)

var ErrInvalidStat = errors.New("invalid stat column")

type Player struct {
	PlayerID string `json:"player_id"`
	FullName string `json:"full_name"`
}

type SeasonLine struct {
	Season       string   `json:"season"`
	Team         string   `json:"team"`
	Pos          *string  `json:"pos"`
	Age          *int     `json:"age"`
	Games        *int     `json:"games"`
	GamesStarted *int     `json:"games_started"`
	MinutesPG    *float64 `json:"minutes_pg"`
	Pts          *float64 `json:"pts"`
	Reb          *float64 `json:"reb"`
	Ast          *float64 `json:"ast"`
	Stl          *float64 `json:"stl"`
	Blk          *float64 `json:"blk"`
	Tov          *float64 `json:"tov"`
	Fgm          *float64 `json:"fgm"`
	Fga          *float64 `json:"fga"`
	FgPct        *float64 `json:"fg_pct"`
	Fg3m         *float64 `json:"fg3m"`
	Fg3a         *float64 `json:"fg3a"`
	Fg3Pct       *float64 `json:"fg3_pct"`
	Ftm          *float64 `json:"ftm"`
	Fta          *float64 `json:"fta"`
	FtPct        *float64 `json:"ft_pct"`
	EfgPct       *float64 `json:"efg_pct"`
	Per          *float64 `json:"per"`
	TsPct        *float64 `json:"ts_pct"`
	UsgPct       *float64 `json:"usg_pct"`
	Bpm          *float64 `json:"bpm"`
	Vorp         *float64 `json:"vorp"`
	Ws           *float64 `json:"ws"`
}

type GameLog struct {
	GameID    string    `json:"game_id"`
	GameDate  time.Time `json:"game_date"`
	Matchup   string    `json:"matchup"`
	WL        string    `json:"wl"`
	Min       *float64  `json:"min"`
	Pts       *float64  `json:"pts"`
	Reb       *float64  `json:"reb"`
	Ast       *float64  `json:"ast"`
	Stl       *float64  `json:"stl"`
	Blk       *float64  `json:"blk"`
	Tov       *float64  `json:"tov"`
	Fgm       *float64  `json:"fgm"`
	Fga       *float64  `json:"fga"`
	FgPct     *float64  `json:"fg_pct"`
	Fg3m      *float64  `json:"fg3m"`
	Fg3a      *float64  `json:"fg3a"`
	Fg3Pct    *float64  `json:"fg3_pct"`
	Ftm       *float64  `json:"ftm"`
	Fta       *float64  `json:"fta"`
	FtPct     *float64  `json:"ft_pct"`
	PlusMinus *float64  `json:"plus_minus"`
}

type GameLogParams struct {
	PlayerBRID string
	SeasonYear int // season_end_year: 2024 = 2023-24 season
	SeasonType string
	FgPctLt    *float64
	MinFGA     *int
}

type StatLine struct {
	FullName string   `json:"full_name"`
	PlayerID string   `json:"player_id"`
	Season   string   `json:"season"`
	Team     string   `json:"team"`
	Pts      *float64 `json:"pts"`
	Reb      *float64 `json:"reb"`
	Ast      *float64 `json:"ast"`
	Stl      *float64 `json:"stl"`
	Blk      *float64 `json:"blk"`
	FgPct    *float64 `json:"fg_pct"`
	Fg3Pct   *float64 `json:"fg3_pct"`
	TsPct    *float64 `json:"ts_pct"`
	Per      *float64 `json:"per"`
	Bpm      *float64 `json:"bpm"`
	Vorp     *float64 `json:"vorp"`
	Ws       *float64 `json:"ws"`
}

type LeaderRow struct {
	Stat     string   `json:"stat"`
	FullName string   `json:"full_name"`
	PlayerID string   `json:"player_id"`
	Season   string   `json:"season"`
	Team     string   `json:"team"`
	Value    *float64 `json:"value"`
}

type VsTeamSeason struct {
	Season    string  `json:"season"`
	Games     int     `json:"games"`
	Pts       float64 `json:"pts"`
	Reb       float64 `json:"reb"`
	Ast       float64 `json:"ast"`
	Stl       float64 `json:"stl"`
	Blk       float64 `json:"blk"`
	Tov       float64 `json:"tov"`
	FgPct     float64 `json:"fgpct"`
	Fg3Pct    float64 `json:"fg3pct"`
	PlusMinus float64 `json:"plusminus"`
}

type VsTeamParams struct {
	PlayerBRID string
	Team       string
	SeasonYear int
	SeasonType string
}

type VsTeamAllTimeParams struct {
	PlayerBRID string
	Team       string
	SeasonType string
}

type HeadToHead struct {
	PlayerA     string `json:"player_a"`
	PlayerB     string `json:"player_b"`
	GamesPlayed int    `json:"games_played"`
	AWins       int    `json:"a_wins"`
	BWins       int    `json:"b_wins"`
}
