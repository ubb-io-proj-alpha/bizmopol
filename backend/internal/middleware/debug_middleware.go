package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func DebugLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		log.Printf("[DEBUG] %s %s %d %s",
			method, path, c.Writer.Status(), time.Since(start))
	}
}
