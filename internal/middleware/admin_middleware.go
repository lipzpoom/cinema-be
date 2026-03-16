package middleware

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

// RequireRoles checks if the current user has one of the allowed roles
func RequireRoles(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		userRole := role.(string)
		hasRole := slices.Contains(allowedRoles, userRole)

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "You do not have permission to perform this action",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
