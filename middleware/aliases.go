package middleware

import "github.com/gin-gonic/gin"

// CacheResponse is preserved for compatibility; it now delegates to Cache.
func CacheResponse() func(c *gin.Context) {
	return Cache()
}

// RateLimit is preserved for compatibility; it now delegates to the existing
// global API rate limit middleware.
func RateLimit() func(c *gin.Context) {
	return GlobalAPIRateLimit()
}
