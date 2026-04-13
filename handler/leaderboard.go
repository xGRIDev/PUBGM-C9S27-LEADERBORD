package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xGRIDev/pubgm-leaderboard-tppsquad/models"
	"github.com/xGRIDev/pubgm-leaderboard-tppsquad/redis"
)

func GetGlobalLBRank(c *gin.Context) {
	pubgmlb := c.Query("LBWeapon")
	if pubgmlb == "" {
		pubgmlb = "pubgm-weapon-global-smg"
	}

	period := c.DefaultQuery("period", "C9S27")
	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit > 1000 {
		limit = 100
	}

	//user-id
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	//key := "Leaderboard:" + pubgmlb + ":" + period + "user:" + userID
	key := "Leaderboard:" + pubgmlb + ":" + period
	results, err := redis.Client.ZRevRangeWithScores(redis.Ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch LB-Rank"})
		return
	}

	leaderboard := make([]models.LBEntryRank, len(results))
	for i, result := range results {
		userID := result.Member.(string)

		leaderboard[i] = models.LBEntryRank{
			Rank:   i + 1,
			UserID: userID,
			Score:  int(result.Score),
			// Assuming Kills and Damage are stored separately, you would need to fetch them as well
			Kills:  100,    // Placeholder, replace with actual value
			Damage: 1000.0, // Placeholder, replace with actual value
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"game":         pubgmlb,
		"period":       period,
		"players":      leaderboard,
		"generated_at": time.Now().Format(time.RFC3339),
		"requested_by": getUserName(userID),
	})
}

func GetRankUser(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	game := c.Query("game")
	if game == "" {
		game = "smg-vector" // Default game mode
	}

	period := c.DefaultQuery("period", "C9S7")

	// Make sure key matches how you store in SubmitScore function
	key := "score:" + userID + ":" + game + ":" + time.Now().Format("20060102150405")

	// Debug: Check if key exists
	exists, err := redis.Client.Exists(redis.Ctx, key).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database error",
			"details": err.Error(),
		})
		return
	}

	if exists == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "Leaderboard is empty or not found",
			"game":    game,
			"period":  period,
			"rank":    nil,
			"score":   nil,
		})
		return
	}

	// Get user rank (0-based index, higher score = lower number)
	rank, err := redis.Client.ZRevRank(redis.Ctx, key, userID).Result()
	if err != nil {
		// User not found in leaderboard
		c.JSON(http.StatusOK, gin.H{
			"message": "User not found in leaderboard",
			"game":    game,
			"period":  period,
			"rank":    nil,
			"score":   nil,
		})
		return
	}

	// Get user score
	score, err := redis.Client.ZScore(redis.Ctx, key, userID).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Could not fetch score",
			"details": err.Error(),
		})
		return
	}

	// Get total players for percentile calculation
	totalPlayers, err := redis.Client.ZCard(redis.Ctx, key).Result()
	if err != nil {
		totalPlayers = 0
	}

	// Calculate percentile (if in top)
	percentile := 100.0
	if totalPlayers > 0 {
		percentile = float64(rank+1) / float64(totalPlayers) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"rank":          rank + 1, // Convert to 1-based ranking
		"score":         score,
		"game":          game,
		"period":        period,
		"user_id":       userID,
		"total_players": totalPlayers,
		"percentile":    fmt.Sprintf("%.1f%%", percentile),
		"top_percent":   fmt.Sprintf("Top %.1f%%", percentile),
	})
}

func GetTopReportPlayers(c *gin.Context) {
	var req models.TopPlayerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ssperiod := req.Period
	if ssperiod == "" {
		ssperiod = "C9S7"
	}

	key := "Leaderboard:" + req.Game + ":" + ssperiod
	results, err := redis.Client.ZRevRangeWithScores(redis.Ctx, key, 0, int64(req.Limit-1)).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch top players"})
		return
	}

	topPlyrs := make([]gin.H, len(results))
	for i, result := range results {
		userID := result.Member.(string)
		username := getUserName(userID)
		topPlyrs[i] = gin.H{
			"rank":     i + 1,
			"user_id":  userID,
			"username": username,
			"score":    result.Score,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"game":         req.Game,
		"period":       req.Period,
		"top_players":  topPlyrs,
		"limit":        req.Limit,
		"generated_at": time.Now().Format(time.RFC3339),
	})
}

func getUserName(userID string) string {
	return "Player_" + userID[:8]
}

/* func getPlayerStats(userID, game string) (kills int, damage float64, survivalTime float64) {
	kills = 0
	damage = 0.0
	survivalTime = 0.0
} */
