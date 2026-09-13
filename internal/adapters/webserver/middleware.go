package webserver

import (
	"net/http"
	"slices"
	"time"

	"github.com/gin-gonic/gin"

	"DemoExchange/internal/app/entities"
)

func authSecretMiddleware(secrets []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		secret := c.GetHeader("secret")

		if slices.Contains(secrets, secret) {
			return
		}

		c.AbortWithStatusJSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "Permission denied",
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
	}
}

func (r *Routes) authTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		token := c.GetHeader("token")
		if token == "" {
			r.log.Errorf("authTokenMiddleware:Token is empty [url: %v, headers: %v]", c.Request.URL, c.Request.Header)
			c.AbortWithStatusJSON(http.StatusOK, gin.H{
				"error": "Invalid API-key",
				"time":  time.Now().Format("2006-01-02 15:04:05"),
			})
			return
		}

		accountUID, err := r.usecase.GetAccountUID(ctx, entities.Token(token))
		if err != nil {
			r.log.Errorf("authTokenMiddleware:GetAccountUID error: %v [url: %v, headers: %v]", err, c.Request.URL, c.Request.Header)
			c.AbortWithStatusJSON(http.StatusOK, gin.H{
				"error": "Invalid API-key",
				"time":  time.Now().Format("2006-01-02 15:04:05"),
			})
			return
		}

		c.Set("accountUID", accountUID)

		c.Next()
	}
}
