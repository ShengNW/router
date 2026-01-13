package group

import (
	"net/http"

	"github.com/gin-gonic/gin"
	groupsvc "github.com/yeying-community/router/internal/admin/service/group"
)

func GetGroups(c *gin.Context) {
	groupNames := groupsvc.List()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    groupNames,
	})
}
