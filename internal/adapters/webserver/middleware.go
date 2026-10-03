package webserver

import (
	"net/http"
	"slices"
	"time"

	"github.com/gin-gonic/gin"

	"DemoExchange/internal/app/entities"
)

const timeFormat = "2006-01-02 15:04:05"

func authSecretMiddleware(secrets []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		secret := c.GetHeader("secret")

		if slices.Contains(secrets, secret) {
			return
		}

		c.AbortWithStatusJSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "Permission denied",
			"time":    time.Now().Format(timeFormat),
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
				"time":  time.Now().Format(timeFormat),
			})
			return
		}

		now := time.Now()
		accountUID, err := r.usecase.GetAccountUID(ctx, entities.Token(token))
		if err != nil {
			r.log.Errorf("authTokenMiddleware:GetAccountUID error: %v [token: %s, url: %v, headers: %v, remoteAddr: %s, duration: %v]", err, token, c.Request.URL, c.Request.Header, c.Request.RemoteAddr, time.Since(now))
			c.AbortWithStatusJSON(http.StatusOK, gin.H{
				"error": "Account verification error",
				"time":  time.Now().Format(timeFormat),
			})
			return
		}

		c.Set("accountUID", accountUID)

		c.Next()
	}
}
