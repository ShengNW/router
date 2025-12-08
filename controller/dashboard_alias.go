package controller

import "github.com/gin-gonic/gin"

// GetUserLogsDashBoard is kept for backward compatibility with historical
// routing; it now delegates to GetUserDashboard.
func GetUserLogsDashBoard(c *gin.Context) {
	GetUserDashboard(c)
}
