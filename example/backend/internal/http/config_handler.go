package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/dtomschitz/headless-go-client/example/backend/internal"
	"github.com/gin-gonic/gin"

	_ "embed"
)

//go:embed dummy_data/manifest.json
var dummyManifest []byte

//go:embed dummy_data/config.json
var dummyConfig []byte

type ConfigHandler struct {
	configService *internal.ConfigService
}

func NewConfigHandler(svc *internal.ConfigService) *ConfigHandler {
	return &ConfigHandler{
		configService: svc,
	}
}

func (h *ConfigHandler) CreateConfig(c *gin.Context) {
	var config internal.Config
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload", "details": err.Error()})
		return
	}

	err := h.configService.CreateConfig(c.Request.Context(), &config)
	if err != nil {
		if err.Error() == fmt.Sprintf("config with version '%s' already exists", config.Version) { // Example for specific error string check
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create config", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Configuration created successfully"})
}

func (h *ConfigHandler) GetConfigByVersion(c *gin.Context) {
	logger := internal.NewLogger(c)

	version := c.Param("version")
	if version == "" {
		c.JSON(NewProblemFromError(internal.NewInvalidRequestError(errors.New("version parameter is required"))))
		return
	}

	/*config, err := h.configService.GetConfigByVersion(c.Request.Context(), version)
	if err != nil {
		logger.With(err).Error("failed to retrieve config %s", version)
		c.JSON(NewProblemFromError(fmt.Errorf("failed to retrieve config %s", version)))
		return
	}*/

	logger.With("config", string(dummyConfig)).Info("fetch config")

	c.Data(http.StatusOK, "application/json", dummyConfig)
}

func (h *ConfigHandler) GetLatestManifest(c *gin.Context) {
	logger := internal.NewLogger(c)

	/*config, err := h.configService.GetLatestConfig(c.Request.Context())
	if err != nil {
		logger.With(err).Error("failed to retrieve latest config")
		c.JSON(NewProblemFromError(errors.New("failed to retrieve latest config")))
		return
	}*/

	logger.With("manifest", string(dummyManifest)).Info("fetch latest manifest")

	c.Data(http.StatusOK, "application/json", dummyManifest)
}
