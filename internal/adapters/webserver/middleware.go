package webserver

import (
	"DemoExchange/internal/app/entities"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func authSecretMiddleware(secrets []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		secret := c.GetHeader("secret")

		for _, s := range secrets {
			if s == secret {
				return
			}
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
		accountUID, err := r.usecase.GetAccountUID(ctx, entities.Token(token))
		if err != nil {
			r.log.Errorf("GetAccountUID error: %v [headers: %v]", err, c.Request.Header)
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
