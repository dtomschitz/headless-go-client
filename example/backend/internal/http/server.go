package http

import (
	"github.com/dtomschitz/headless-go-client/example/backend/internal"
	"github.com/gin-gonic/gin"
)

func StartServer(configService *internal.ConfigService) error {
	router := gin.Default()

	configHandler := NewConfigHandler(configService)
	clientUpdateHandler := NewClientUpdateHandler()

	// API Group
	api := router.Group("/api/v1")
	{
		configs := api.Group("/configs")
		{
			configs.POST("", configHandler.CreateConfig)
			configs.GET("/:version/properties", configHandler.GetConfigByVersion)
			configs.GET("/manifest", configHandler.GetLatestManifest)
		}

		clientUpdate := api.Group("/client")
		{
			clientUpdate.GET("/:version/binary", clientUpdateHandler.GetBinaryByVersion)
			clientUpdate.GET("/manifest", clientUpdateHandler.GetLatestManifest)
		}
	}

	return router.Run()
}
