package handler

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xGRIDev/pubgm-leaderboard-tppsquad/models"
	"github.com/xGRIDev/pubgm-leaderboard-tppsquad/redis"
)

func SubmitScore(c *gin.Context) {
	// Debug: Log incoming request
	log.Println("SubmitScore: Processing request...")

	// Get user_id from context (set by middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		log.Println("SubmitScore: user_id not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		log.Println("SubmitScore: user_id is not a valid string")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	log.Printf("SubmitScore: Authenticated user_id: %s\n", userIDStr)

	var score models.Scores
	if err := c.ShouldBindJSON(&score); err != nil {
		log.Printf("SubmitScore: Bind error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Calculate total score
	score.Score = ScoreCalc(score.Kills, score.Damage, score.Survival)
	score.UserID = score.UserID
	score.Submitted = time.Now()

	log.Printf("SubmitScore: Processing score for user %s: %+v\n", userIDStr, score)

	//store score in redis
	gameKey := "Leaderboard:" + score.Game + "C9S27"

	//get current / all time leaderboard
	err := redis.Client.ZIncrBy(redis.Ctx, gameKey+":C9S27", float64(score.Score), userIDStr).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update LB"})
		return
	}

	//Get weekkly LB
	weekLBKey := getWeekRankKey()
	err = redis.Client.ZIncrBy(redis.Ctx, gameKey+":weekly:"+weekLBKey, float64(score.Score), userIDStr).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update weekly LB"})
		return
	}

	//detail score
	keyScore := "score:" + userIDStr + ":" + score.Game + ":" + time.Now().Format("20060102150405")
	mapScore := map[string]interface{}{
		"user_id":   score.UserID,
		"game":      score.Game,
		"kills":     score.Kills,
		"damage":    score.Damage,
		"survival":  score.Survival,
		"score":     score.Score,
		"submitted": score.Submitted.Format(time.RFC3339),
	}

	err = redis.Client.HSet(redis.Ctx, keyScore, mapScore).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not saving score details"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		  "message": "Score submitted Successfully",
		"score":   score.Score,
	})
}

func ScoreCalc(kills int, damage float64, survive float64) int {
	return kills*100 + int(damage) + int(survive*10)
}

func getWeekRankKey() string {
	now := time.Now()
	year, week := now.ISOWeek()
	return string(rune(year)) + "-" + string(rune(week))
}
