package handler

import (
	"net/http"
	"strings"
	"time"

	"multifinance-core/internal/repository"
	"multifinance-core/internal/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(authRepo repository.AuthRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" || !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization"})
			return
		}
		token := strings.TrimPrefix(h, "Bearer ")
		email, _, err := utils.ParseToken(token, 24*time.Hour)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		user, err := authRepo.FindByEmail(c.Request.Context(), email)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set("auth_user", user)
		c.Next()
	}
}
