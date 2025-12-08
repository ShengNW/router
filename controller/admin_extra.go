package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yeying-community/router/model"
)

// GetChannelModels provides the available model list for each channel type.
// Currently it returns an empty map to keep the API surface compatible.
func GetChannelModels(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    map[string][]string{},
	})
}

// GetUserStatus currently reuses GetAllUsers to provide user information for the admin panel.
func GetUserStatus(c *gin.Context) {
	GetAllUsers(c)
}

// AddUser allows administrators to create a new user directly.
func AddUser(c *gin.Context) {
	ctx := c.Request.Context()
	user := model.User{}
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if user.Username == "" || user.Password == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户名或密码不能为空",
		})
		return
	}
	if user.DisplayName == "" {
		user.DisplayName = user.Username
	}
	if user.Role == 0 {
		user.Role = model.RoleCommonUser
	}
	if user.Status == 0 {
		user.Status = model.UserStatusEnabled
	}
	if err := user.Insert(ctx, 0); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    user,
	})
}
