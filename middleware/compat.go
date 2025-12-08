package middleware

import "github.com/gin-gonic/gin"

func CacheResponse() gin.HandlerFunc {
	return Cache()
}

func RateLimit() gin.HandlerFunc {
	return GlobalAPIRateLimit()
}
