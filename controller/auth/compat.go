package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func WechatOAuth(c *gin.Context) {
	WeChatAuth(c)
}

func BindWechat(c *gin.Context) {
	WeChatBind(c)
}

func BindEmail(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"message": "email binding is not supported",
	})
}
