package group

import (
	"net/http"

	"github.com/gin-gonic/gin"
	groupsvc "github.com/yeying-community/router/internal/admin/service/group"
)

// GetGroups godoc
// @Summary List groups (admin)
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} docs.StandardResponse
// @Failure 401 {object} docs.ErrorResponse
// @Router /api/v1/admin/group [get]
func GetGroups(c *gin.Context) {
	groupNames := groupsvc.List()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    groupNames,
	})
}
