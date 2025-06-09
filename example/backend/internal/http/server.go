package http

import (
	"github.com/dtomschitz/headless-go-client/example/backend/internal"
	"github.com/gin-gonic/gin"
)

func StartServer(configService *internal.ConfigService) error {
	router := gin.Default()

	configHandler := NewConfigHandler(configService)

	// API Group
	api := router.Group("/api/v1")
	{
		configs := api.Group("/configs")
		{
			configs.POST("", configHandler.CreateConfig)
			configs.GET("/:version", configHandler.GetConfigByVersion)
			configs.GET("/latest", configHandler.GetLatestConfig)
		}
	}

	return router.Run()
}
