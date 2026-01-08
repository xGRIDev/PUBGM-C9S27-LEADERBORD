package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/xGRIDev/pubgm-leaderboard-tppsquad/handler"
	"github.com/xGRIDev/pubgm-leaderboard-tppsquad/middleware"
	"github.com/xGRIDev/pubgm-leaderboard-tppsquad/redis"
)

func main() {
	if err := redis.InitRedis(); err != nil {
		log.Fatalf("failed to connect redis: %v", err)
	}
	defer redis.CloseRedis()

	router := gin.Default()

	//Setup auth routes
	auth := router.Group("/auth")
	{
		auth.POST("/register", handler.LB_Regist)
		auth.POST("/login", handler.LB_Login)
	}

	//set protecting routes
	api := router.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		//For Submission Score
		api.POST("/scores", handler.SubmitScore)
		api.GET("/leaderboard", handler.GetGlobalLBRank)
		api.GET("/leaderboard/my-rank", handler.GetRankUser)
		api.POST("/leaderboard/top-players", handler.GetTopReportPlayers)
	}

	//Health
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	log.Println("Server Starting on :8090")
	if err := router.Run(":8090"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
