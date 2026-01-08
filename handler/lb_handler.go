package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xGRIDev/pubgm-leaderboard-tppsquad/middleware"
	"github.com/xGRIDev/pubgm-leaderboard-tppsquad/models"
	"github.com/xGRIDev/pubgm-leaderboard-tppsquad/redis"
	"golang.org/x/crypto/bcrypt"
)

func LB_Regist(c *gin.Context) {
	var user models.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	exists, err := redis.Client.HExists(redis.Ctx, "users", user.Username).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists"})
		return
	}

	// Hashing password
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash password"})
		return
	}

	user.ID = GenerateID()
	user.Password = string(hashedPass)
	user.CreatedAt = time.Now()

	// Convert user to JSON for storage
	userJSON, err := json.Marshal(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not marshal user data"})
		return
	}

	// Store user as JSON string in Redis hash
	err = redis.Client.HSet(redis.Ctx, "users", user.Username, userJSON).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save user"})
		return
	}

	// Also store email to username mapping for uniqueness
	err = redis.Client.HSet(redis.Ctx, "user_emails", user.Email, user.Username).Err()
	if err != nil {
		fmt.Printf("Warning: Could not save email mapping: %v\n", err)
	}

	// generating for token
	token, err := middleware.GenerateTKN(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	user.Password = ""
	c.JSON(http.StatusCreated, models.AuthResponse{
		Token: token,
		User:  user,
	})
}

func LB_Login(c *gin.Context) {
	var req models.Login
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user from Redis
	userJSON, err := redis.Client.HGet(redis.Ctx, "users", req.Username).Result()
	if err != nil || userJSON == "" {
		// Try alternative lookup method
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	var user models.User
	err = json.Unmarshal([]byte(userJSON), &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not parse user data"})
		return
	}

	// Compare buat  password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate token
	token, err := middleware.GenerateTKN(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	// Don't return password
	user.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user":    user,
	})
}

func GenerateID() string {
	return time.Now().Format("20060102150405") + "-" + string(rune(time.Now().Nanosecond()))
}
