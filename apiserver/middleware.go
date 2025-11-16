package apiserver

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/oauth2/v4"
)

// OAuth2Middleware validates OAuth2 access tokens
func OAuth2Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if oauth2Server == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "OAuth2 server not initialized"})
			c.Abort()
			return
		}

		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format. Expected: Bearer <token>"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token
		tokenInfo, err := oauth2Server.Manager.LoadAccessToken(c.Request.Context(), tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Store token info in context for use in handlers
		c.Set("oauth_token", tokenInfo)
		c.Set("user_id", tokenInfo.GetUserID())
		c.Set("client_id", tokenInfo.GetClientID())

		c.Next()
	}
}

// GetTokenInfo retrieves OAuth2 token information from context
func GetTokenInfo(c *gin.Context) oauth2.TokenInfo {
	if tokenInfo, exists := c.Get("oauth_token"); exists {
		return tokenInfo.(oauth2.TokenInfo)
	}
	return nil
}

// GetUserID retrieves the authenticated user ID from context
func GetUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		return userID.(string)
	}
	return ""
}

// GetClientID retrieves the authenticated client ID from context
func GetClientID(c *gin.Context) string {
	if clientID, exists := c.Get("client_id"); exists {
		return clientID.(string)
	}
	return ""
}
