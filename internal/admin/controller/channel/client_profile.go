package channel

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yeying-community/router/internal/admin/model"
)

// GetClientProfiles godoc
// @Summary Get client profiles (admin)
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} docs.StandardResponse
// @Failure 401 {object} docs.ErrorResponse
// @Router /api/v1/admin/channel/client_profiles [get]
func GetClientProfiles(c *gin.Context) {
	profiles, err := model.ListEnabledClientProfilesWithDB(model.DB)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "读取客户端画像失败: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    profiles,
	})
}
