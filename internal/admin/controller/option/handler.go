package option

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yeying-community/router/common/i18n"
	"github.com/yeying-community/router/internal/admin/model"
	optionsvc "github.com/yeying-community/router/internal/admin/service/option"
)

func GetOptions(c *gin.Context) {
	options := optionsvc.GetOptions()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    options,
	})
	return
}

func UpdateOption(c *gin.Context) {
	var option model.Option
	err := json.NewDecoder(c.Request.Body).Decode(&option)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_parameter"),
		})
		return
	}
	// No special validation for options here.
	err = optionsvc.UpdateOption(option.Key, option.Value)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}
