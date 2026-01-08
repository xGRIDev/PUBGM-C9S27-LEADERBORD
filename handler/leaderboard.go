package handler

import (
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

	period := c.DefaultQuery("period", "all_time")
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

	key := "Leaderboard:" + pubgmlb + ":" + period + "user:" + userID

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
		game = "SMG-Rank-VECTOR"
	}
	period := c.DefaultQuery("period", "all_time")
	key := "Leaderboard:" + game + ":" + period

	//Get user of Rank (Based 0-index)

	rank, err := redis.Client.ZRevRank(redis.Ctx, key, userID).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't fetch of rank"})
		return
	}

	//Get User of score
	score, err := redis.Client.ZScore(redis.Ctx, key, userID).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't fetch of score"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"rank":   rank + 1,
		"score":  score,
		"game":   game,
		"period": period,
		"user":   getUserName(userID),
	})
}

func GetTopReportPlayers(c *gin.Context) {
	var req models.TopPlayerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	period := req.Period
	if period == "" {
		period = "all_time"
	}

	key := "Leaderboard:" + req.Game + ":" + period

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
		"generated_at": time.Now().Format(time.RFC3339),
	})
}

func getUserName(userID string) string {
	return "Player_" + userID[:8]
}
