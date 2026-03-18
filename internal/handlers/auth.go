package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	apiKey string
}

func NewAuthMiddleware(apiKey string) *AuthMiddleware {
	return &AuthMiddleware{apiKey: apiKey}
}

func (m *AuthMiddleware) RequireAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.apiKey == "" {
			c.Next()
			return
		}

		if key := strings.TrimSpace(c.GetHeader("X-API-Key")); key != "" {
			if key == m.apiKey {
				c.Next()
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
			return
		}

		auth := strings.TrimSpace(c.GetHeader("Authorization"))
		if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			token := strings.TrimSpace(auth[len("bearer "):])
			if token == m.apiKey {
				c.Next()
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid bearer token"})
			return
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing credentials"})
	}
}
