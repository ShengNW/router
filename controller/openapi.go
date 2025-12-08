package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetOpenAPI returns a not-found status to indicate that no OpenAPI document
// is currently published.
func GetOpenAPI(c *gin.Context) {
	c.Status(http.StatusNotFound)
}
