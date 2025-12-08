package auth

import "github.com/gin-gonic/gin"

// WechatOAuth is kept for backward compatibility with older routing logic.
// It delegates to the current WeChatAuth handler.
func WechatOAuth(c *gin.Context) {
	WeChatAuth(c)
}

// BindWechat preserves the previous handler name for binding requests.
func BindWechat(c *gin.Context) {
	WeChatBind(c)
}

// BindEmail is currently not implemented; it returns 404 to indicate the
// endpoint is unavailable.
func BindEmail(c *gin.Context) {
	c.Status(404)
}
