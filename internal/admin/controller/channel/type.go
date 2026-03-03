package channel

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yeying-community/router/internal/admin/model"
)

type channelTypeOptionItem struct {
	Key         int    `json:"key"`
	Text        string `json:"text"`
	Value       int    `json:"value"`
	Name        string `json:"name,omitempty"`
	Color       string `json:"color,omitempty"`
	Description string `json:"description,omitempty"`
	Tip         string `json:"tip,omitempty"`
}

// GetChannelTypes godoc
// @Summary Get channel interface type catalog (admin)
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} docs.StandardResponse
// @Failure 401 {object} docs.ErrorResponse
// @Router /api/v1/admin/channel/types [get]
func GetChannelTypes(c *gin.Context) {
	rows := make([]model.ChannelTypeCatalog, 0)
	if err := model.DB.
		Where("enabled = ?", true).
		Order("sort_order asc, id asc").
		Find(&rows).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "读取接口类型配置失败: " + err.Error(),
		})
		return
	}

	options := make([]channelTypeOptionItem, 0, len(rows))
	for _, row := range rows {
		label := strings.TrimSpace(row.Label)
		if label == "" {
			label = strings.TrimSpace(row.Name)
		}
		if label == "" {
			label = "unknown"
		}
		options = append(options, channelTypeOptionItem{
			Key:         row.ID,
			Value:       row.ID,
			Text:        label,
			Name:        strings.TrimSpace(row.Name),
			Color:       strings.TrimSpace(row.Color),
			Description: strings.TrimSpace(row.Description),
			Tip:         strings.TrimSpace(row.Tip),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    options,
	})
}
