package http

import (
	"errors"
	"net/http"

	"github.com/dtomschitz/headless-go-client/example/backend/internal"
	"github.com/gin-gonic/gin"

	_ "embed"
)

//go:embed dummy_data/client_update_manifest.json
var dummyClientUpdateManifest []byte

//go:embed dummy_data/client
var dummyClientBinary []byte

type ClientUpdateHandler struct {
}

func NewClientUpdateHandler() *ClientUpdateHandler {
	return &ClientUpdateHandler{}
}

func (h *ClientUpdateHandler) GetBinaryByVersion(c *gin.Context) {
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

	logger.Info("fetch config")

	c.Data(http.StatusOK, "application/json", dummyClientBinary)
}

func (h *ClientUpdateHandler) GetLatestManifest(c *gin.Context) {
	logger := internal.NewLogger(c)

	/*config, err := h.configService.GetLatestConfig(c.Request.Context())
	if err != nil {
		logger.With(err).Error("failed to retrieve latest config")
		c.JSON(NewProblemFromError(errors.New("failed to retrieve latest config")))
		return
	}*/

	logger.With("manifest", string(dummyClientUpdateManifest)).Info("fetch latest manifest")

	c.Data(http.StatusOK, "application/json", dummyClientUpdateManifest)
}
