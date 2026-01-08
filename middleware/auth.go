package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var secretJWT = []byte("pubgmlb")

type AuthClaims struct {
	UserId string `json:"user_id"`
	jwt.RegisteredClaims
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("AuthMiddleware: Checking authentication...")

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Println("AuthMiddleware: No Authorization header found")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		log.Printf("AuthMiddleware: Raw Authorization header: %s\n", authHeader)

		// Check if header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			log.Println("AuthMiddleware: Authorization header doesn't start with 'Bearer '")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must start with 'Bearer '"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		tokenString = strings.TrimSpace(tokenString)

		/*log.Printf("AuthMiddleware: Token string (first 50 chars): %s...\n",
		safeSubstring(tokenString, 0, 50))*/

		if tokenString == "" {
			log.Println("AuthMiddleware: Token string is empty")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is empty"})
			c.Abort()
			return
		}

		claims := &AuthClaims{}

		// parsing token
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// validasi signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return secretJWT, nil
		})

		if err != nil {
			log.Printf("AuthMiddleware: Token parsing error: %v\n", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid token",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		if !token.Valid {
			log.Println("AuthMiddleware: Token is invalid")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		log.Printf("AuthMiddleware: Token valid for user_id: %s\n", claims.UserId)
		c.Set("user_id", claims.UserId)
		c.Next()
	}

}

func GenerateTKN(userID string) (string, error) {
	claims := &AuthClaims{
		UserId:           userID,
		RegisteredClaims: jwt.RegisteredClaims{},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretJWT)
}
