package models

import "time"

type Scores struct {
	UserID    int64     `json:"user_id"`
	Game      string    `json:"game" binding:"required"`
	Kills     int       `json:"kills" binding:"required"`
	Damage    float64   `json:"damage" binding:"required"`
	Survival  float64   `json:"survival_time" binding:"required"`
	Score     int       `json:"score"`
	Submitted time.Time `json:"submitted_at"`
}

type LBEntryRank struct {
	Rank   int     `json:"rank"`
	UserID string  `json:"user_id"`
	Score  int     `json:"score"`
	Kills  int     `json:"kills"`
	Damage float64 `json:"damage"`
}

type TopPlayerReq struct {
	Game   string `json:"game" binding:"required"`
	Period string `json:"period"`
	Limit  int    `json:"limit" binding:"min=1,max=100"`
}
