package router

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	"github.com/yeying-community/router/internal/admin/controller/billing"
	"github.com/yeying-community/router/internal/transport/http/middleware"
)

func SetDashboardRouter(engine *gin.Engine) {
	apiRouter := engine.Group("/")
	apiRouter.Use(gzip.Gzip(gzip.DefaultCompression))
	apiRouter.Use(middleware.GlobalAPIRateLimit())
	apiRouter.Use(middleware.TokenAuth())
	{
		apiRouter.GET("/dashboard/billing/subscription", billing.GetSubscription)
		apiRouter.GET("/v1/dashboard/billing/subscription", billing.GetSubscription)
		apiRouter.GET("/dashboard/billing/usage", billing.GetUsage)
		apiRouter.GET("/v1/dashboard/billing/usage", billing.GetUsage)
	}
}
