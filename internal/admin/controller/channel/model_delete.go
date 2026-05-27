package channel

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yeying-community/router/internal/admin/model"
)

type deleteChannelModelRequest struct {
	Model         string `json:"model"`
	UpstreamModel string `json:"upstream_model"`
}

func DeleteChannelModel(c *gin.Context) {
	channelID := strings.TrimSpace(c.Param("id"))
	if channelID == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "渠道 ID 无效"})
		return
	}
	modelName := strings.TrimSpace(c.Query("model"))
	upstreamModel := strings.TrimSpace(c.Query("upstream_model"))
	if modelName == "" && upstreamModel == "" {
		req := deleteChannelModelRequest{}
		if err := c.ShouldBindJSON(&req); err == nil {
			modelName = strings.TrimSpace(req.Model)
			upstreamModel = strings.TrimSpace(req.UpstreamModel)
		}
	}
	if modelName == "" && upstreamModel == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "渠道模型无效"})
		return
	}
	logModelName := modelName
	if logModelName == "" {
		logModelName = upstreamModel
	}
	if err := model.DeleteChannelModelWithDB(model.DB, channelID, modelName, upstreamModel); err != nil {
		logChannelAdminWarn(c, "delete_model", stringField("channel_id", channelID), stringField("model", logModelName), stringField("reason", err.Error()))
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	logChannelAdminInfo(c, "delete_model", stringField("channel_id", channelID), stringField("model", logModelName))
	c.JSON(http.StatusOK, gin.H{"success": true, "message": ""})
}
