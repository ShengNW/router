package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yeying-community/router/common/config"
)

func GetOpenAPI(c *gin.Context) {
	openapi := gin.H{
		"openapi": "3.0.0",
		"info": gin.H{
			"title":       config.SystemName,
			"version":     "1.0.0",
			"description": "OpenAPI schema metadata for the router service.",
		},
		"servers": []gin.H{
			{
				"url":         config.ServerAddress,
				"description": "Primary router server",
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    openapi,
	})
}
