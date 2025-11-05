package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/piyushsharan/rate-limiter/internal/limiter"
)

const (
	clientIDHeader = "X-Client-ID"
)

// RateLimiter returns a gin middleware that enforces limiter decisions.
func RateLimiter(l limiter.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := buildKey(c)

		decision, err := l.Allow(c.Request.Context(), key)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "rate limiter unavailable",
			})
			return
		}

		if !decision.Allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":          "rate limit exceeded",
				"retry_after_ms": decision.RetryAfter.Milliseconds(),
			})
			return
		}

		c.Header("X-RateLimit-Remaining", strconv.Itoa(decision.Remaining))
		c.Next()
	}
}

func buildKey(c *gin.Context) string {
	clientID := c.GetHeader(clientIDHeader)
	if clientID == "" {
		clientID = c.ClientIP()
	}

	route := c.FullPath()
	if route == "" {
		// For unmatched routes fall back to URL path.
		route = c.Request.URL.Path
	}

	return strings.Join([]string{clientID, route}, ":")
}
