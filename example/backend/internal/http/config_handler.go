package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/dtomschitz/headless-go-client/example/backend/internal"
	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	configService *internal.ConfigService
}

func NewConfigHandler(svc *internal.ConfigService) *ConfigHandler {
	return &ConfigHandler{
		configService: svc,
	}
}

// CreateConfig godoc
// @Summary Create a new configuration
// @Description Creates a new remote configuration with a version and properties.
// @Tags configs
// @Accept json
// @Produce json
// @Param config body config.Config true "Config object to be created"
// @Success 201 {object} map[string]string "message: Configuration created successfully"
// @Failure 400 {object} map[string]string "error: Invalid request payload"
// @Failure 409 {object} map[string]string "error: Config with this version already exists"
// @Failure 500 {object} map[string]string "error: Internal server error"
// @Router /configs [post]
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

// GetConfigByVersion godoc
// @Summary Get a configuration by version
// @Description Retrieves a remote configuration by its specific version string.
// @Tags configs
// @Produce json
// @Param version path string true "Config Version"
// @Success 200 {object} config.Config
// @Failure 404 {object} map[string]string "error: Config not found"
// @Failure 500 {object} map[string]string "error: Internal server error"
// @Router /configs/{version} [get]
func (h *ConfigHandler) GetConfigByVersion(c *gin.Context) {
	logger := internal.NewLogger(c)

	version := c.Param("version")
	if version == "" {
		c.JSON(NewProblemFromError(internal.NewInvalidRequestError(errors.New("version parameter is required"))))
		return
	}

	config, err := h.configService.GetConfigByVersion(c.Request.Context(), version)
	if err != nil {
		logger.With(err).Error("failed to retrieve config %s", version)
		c.JSON(NewProblemFromError(fmt.Errorf("failed to retrieve config %s", version)))
		return
	}

	c.JSON(http.StatusOK, config)
}

// GetLatestConfig godoc
// @Summary Get the latest configuration
// @Description Retrieves the remote configuration with the highest version number.
// @Tags configs
// @Produce json
// @Success 200 {object} internal.Config
// @Failure 404 {object} map[string]string "error: No configurations found"
// @Failure 500 {object} map[string]string "error: Internal server error"
// @Router /configs/latest [get]
func (h *ConfigHandler) GetLatestConfig(c *gin.Context) {
	logger := internal.NewLogger(c)

	config, err := h.configService.GetLatestConfig(c.Request.Context())
	if err != nil {
		logger.With(err).Error("failed to retrieve latest config")
		c.JSON(NewProblemFromError(errors.New("failed to retrieve latest config")))
		return
	}

	c.JSON(http.StatusOK, config)
}
