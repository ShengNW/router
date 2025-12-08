package router

import (
	"bytes"
	"embed"
	"io"
	"io/fs"
	"net/http"

	"github.com/gin-contrib/sse"
	"github.com/gin-gonic/gin"

	"github.com/yeying-community/router/controller"
	"github.com/yeying-community/router/controller/auth"
	"github.com/yeying-community/router/middleware"
)

func SetRouter(server *gin.Engine, buildFS embed.FS) {
	api := server.Group("/api")
	{
		api.GET("/status", controller.GetStatus)
		api.GET("/notice", controller.GetNotice)
		api.GET("/about", controller.GetAbout)
		api.GET("/home_page_content", controller.GetHomePageContent)
		api.GET("/verification", controller.SendEmailVerification)
		api.POST("/reset_password", controller.SendPasswordResetEmail)
		api.POST("/user/reset", controller.ResetPassword)

		api.POST("/user/register", controller.Register)
		api.POST("/user/login", controller.Login)
		api.GET("/user/logout", controller.Logout)

		api.GET("/oauth/state", auth.GenerateOAuthCode)
		api.GET("/oauth/github", auth.GitHubOAuth)
		api.GET("/oauth/lark", auth.LarkOAuth)
		api.GET("/oauth/wechat", auth.WechatOAuth)
		api.POST("/oauth/wechat/bind", auth.BindWechat)
		api.POST("/oauth/email/bind", auth.BindEmail)
	}

	user := api.Group("/user")
	user.Use(middleware.UserAuth())
	{
		user.GET("/self", controller.GetSelf)
		user.PUT("/self", controller.UpdateSelf)
		user.GET("/dashboard", controller.GetUserDashboard)
		user.GET("/aff", controller.GetAffCode)
		user.POST("/token", controller.GenerateAccessToken)
		user.GET("/token", controller.GetAccessToken)
		user.POST("/topup", controller.TopUp)
		user.GET("/available_models", controller.GetUserAvailableModels)
	}

	admin := api.Group("")
	admin.Use(middleware.AdminAuth())
	{
		admin.GET("/log/", controller.GetAllLogs)
		admin.GET("/log/stat", controller.GetLogsStat)
		admin.GET("/log/search", controller.SearchAllLogs)

		admin.GET("/channel/", controller.GetAllChannels)
		admin.GET("/channel/search", controller.SearchChannels)
		admin.GET("/channel/:id", controller.GetChannel)
		admin.POST("/channel/", controller.AddChannel)
		admin.PUT("/channel/", controller.UpdateChannel)
		admin.DELETE("/channel/:id", controller.DeleteChannel)
		admin.DELETE("/channel/disabled", controller.DeleteDisabledChannel)
		admin.POST("/channel/test/:id", controller.TestChannel)
		admin.POST("/channel/update_balance/:id", controller.UpdateChannelBalance)
		admin.GET("/channel/models", controller.GetChannelModels)

		admin.GET("/token/", controller.GetAllTokens)
		admin.GET("/token/search", controller.SearchTokens)
		admin.GET("/token/:id", controller.GetToken)
		admin.POST("/token/", controller.AddToken)
		admin.PUT("/token/", controller.UpdateToken)
		admin.DELETE("/token/:id", controller.DeleteToken)

		admin.GET("/user/", controller.GetAllUsers)
		admin.GET("/user/search", controller.SearchUsers)
		admin.GET("/user/manage", controller.GetUserStatus)
		admin.GET("/user/:id", controller.GetUser)
		admin.POST("/user/", controller.AddUser)
		admin.PUT("/user/", controller.UpdateUser)
		admin.DELETE("/user/:id", controller.DeleteUser)

		admin.GET("/redemption/", controller.GetAllRedemptions)
		admin.GET("/redemption/search", controller.SearchRedemptions)
		admin.GET("/redemption/:id", controller.GetRedemption)
		admin.POST("/redemption/", controller.AddRedemption)
		admin.PUT("/redemption/", controller.UpdateRedemption)
		admin.DELETE("/redemption/:id", controller.DeleteRedemption)

		admin.GET("/option/", controller.GetOptions)
		admin.POST("/option/", controller.UpdateOption)

		admin.GET("/group/", controller.GetGroups)
		admin.GET("/models", controller.DashboardListModels)
	}

	token := api.Group("")
	token.Use(middleware.TokenAuth())
	{
		token.GET("/models", controller.ListModels)
		token.GET("/models/:model", controller.RetrieveModel)
	}

	selfLog := api.Group("/log")
	selfLog.Use(middleware.UserAuth())
	{
		selfLog.GET("/self/", controller.GetUserLogs)
		selfLog.GET("/self/search", controller.SearchUserLogs)
		selfLog.GET("/self/stat", controller.GetUserLogsDashBoard)
	}

	api.GET("/openapi", controller.GetOpenAPI)

	relay := server.Group("")
	relay.Use(middleware.CacheResponse(), middleware.Distribute())
	relay.Use(middleware.RateLimit(), middleware.TokenAuth())
	relay.Use(middleware.TurnstileCheck())
	{
		relay.Any("/v1/chat/completions", controller.Relay)
		relay.Any("/v1/completions", controller.Relay)
		relay.Any("/v1/edits", controller.Relay)
		relay.Any("/v1/audio/speech", controller.Relay)
		relay.Any("/v1/audio/transcriptions", controller.Relay)
		relay.Any("/v1/audio/translations", controller.Relay)
		relay.Any("/v1/embeddings", controller.Relay)
		relay.Any("/v1/moderations", controller.Relay)
		relay.Any("/v1/images/generations", controller.Relay)
		relay.Any("/v1/oneapi/proxy/:channelid/*path", controller.Relay)
		relay.Any("/v1/oneapi/proxy/*path", controller.Relay)
		relay.Any("/dashboard/stream/model/list", controller.DashboardListModels)
	}

	if web, err := fs.Sub(buildFS, "web/build"); err == nil {
		server.StaticFS("/", http.FS(web))
		server.NoRoute(func(c *gin.Context) {
			file, err := web.Open("index.html")
			if err != nil {
				c.Status(http.StatusNotFound)
				return
			}
			defer file.Close()
			info, _ := file.Stat()
			content, err := io.ReadAll(file)
			if err != nil {
				c.Status(http.StatusInternalServerError)
				return
			}
			http.ServeContent(c.Writer, c.Request, "index.html", info.ModTime(), bytes.NewReader(content))
		})
	}

	server.GET("/sse/notice", func(c *gin.Context) {
		c.Render(http.StatusOK, sse.Event{Event: "notice", Data: gin.H{"message": "ok"}})
	})
}
