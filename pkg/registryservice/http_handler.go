package registryservice

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/izuc/asset-registry/pkg/registry"
	"github.com/izuc/asset-registry/pkg/registry/registryhttp"
	"github.com/labstack/echo"
	"go.uber.org/zap"
)

var (
	Networks = map[string]bool{
		"pollen":   true,
		"nectar":   true,
		"internal": true,
		"test":     true,
	}
)

type HTTPHandler struct {
	service registry.Service
	logger  *zap.SugaredLogger

	filter *SwearFilter
}

func NewHTTPHandler(service registry.Service, logger *zap.SugaredLogger) *HTTPHandler {
	return &HTTPHandler{service: service, logger: logger, filter: NewSwearFilter(true, badWords...)}
}

func networkAllowed(network string) bool {
	return Networks[network]
}

func (h *HTTPHandler) SaveAsset(c echo.Context) error {
	ctx := c.Request().Context()
	network := c.Param("network")
	if !networkAllowed(network) {
		return c.JSON(http.StatusForbidden, "network not allowed")
	}
	var asset *registry.Asset
	if err := json.NewDecoder(c.Request().Body).Decode(&asset); err != nil {
		err = errors.Wrap(err, "failed to parse request body as JSON into an asset")
		h.logger.Infow("Invalid http request", "error", err)
		return c.JSON(http.StatusBadRequest, registryhttp.NewErrorResponse(err))
	}

	if len(asset.Name) > 20 {
		return c.JSON(http.StatusBadRequest, fmt.Sprintf("Asset name too long"))
	}

	if len(asset.Symbol) > 4 {
		return c.JSON(http.StatusBadRequest, fmt.Sprintf("Asset symbol too long"))
	}

	// filter
	match, _ := h.filter.Check(asset.Name)
	if len(match) > 0 {
		return c.JSON(http.StatusBadRequest, fmt.Sprintf("%s is forbidden, as contains %v", asset.Name, match))
	}
	match, _ = h.filter.Check(asset.ID)
	if len(match) > 0 {
		return c.JSON(http.StatusBadRequest, fmt.Sprintf("%s is forbidden, as contains %v", asset.ID, match))
	}
	match, _ = h.filter.Check(asset.Symbol)
	if len(match) > 0 {
		return c.JSON(http.StatusBadRequest, fmt.Sprintf("%s is forbidden, as contains %v", asset.Symbol, match))
	}

	if err := h.service.SaveAsset(ctx, network, asset); err != nil {
		return errors.Wrap(err, "service failed to save Assets")
	}
	return c.JSON(http.StatusCreated, asset)
}

func (h *HTTPHandler) LoadAsset(c echo.Context) error {
	ctx := c.Request().Context()
	network := c.Param("network")
	if !networkAllowed(network) {
		return c.JSON(http.StatusForbidden, "network not allowed")
	}
	ID := c.Param("ID")
	result, err := h.service.LoadAsset(ctx, network, ID)
	if err != nil {
		return c.JSON(http.StatusNotFound, errors.Wrap(err, "service failed to load Asset"))
	}
	return c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) LoadAssets(c echo.Context) error {
	ctx := c.Request().Context()
	network := c.Param("network")
	if !networkAllowed(network) {
		return c.JSON(http.StatusForbidden, "network not allowed")
	}
	result, err := h.service.LoadAssets(ctx, network)
	if err != nil {
		return c.JSON(http.StatusNotFound, errors.Wrap(err, "service failed to load Assets"))
	}
	return c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) DeleteAssetByID(c echo.Context) error {
	ctx := c.Request().Context()
	network := c.Param("network")
	if !networkAllowed(network) {
		return c.JSON(http.StatusForbidden, "network not allowed")
	}
	ID := c.Param("ID")
	err := h.service.DeleteAssetByID(ctx, network, ID)
	if err != nil {
		return c.JSON(http.StatusNotFound, errors.Wrap(err, "service failed to delete the Asset"))
	}
	return c.JSON(http.StatusOK, nil)
}

func (h *HTTPHandler) DeleteAssetByName(c echo.Context) error {
	ctx := c.Request().Context()
	network := c.Param("network")
	if !networkAllowed(network) {
		return c.JSON(http.StatusForbidden, "network not allowed")
	}
	name := c.Param("name")
	err := h.service.DeleteAssetByName(ctx, network, name)
	if err != nil {
		return c.JSON(http.StatusNotFound, errors.Wrap(err, "service failed to delete the Asset"))
	}
	return c.JSON(http.StatusOK, nil)
}

func (h *HTTPHandler) LoadFilter(c echo.Context) error {
	result := h.filter.Load()
	return c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) AddFilter(c echo.Context) error {
	word := c.Param("word")
	if len(word) == 0 {
		return c.JSON(http.StatusBadRequest, "invalid empty-string as filter")
	}
	h.filter.Add(word)
	return c.JSON(http.StatusOK, word)
}

func (h *HTTPHandler) DeleteFilter(c echo.Context) error {
	word := c.Param("word")
	if len(word) == 0 {
		return c.JSON(http.StatusBadRequest, "invalid empty-string as filter")
	}
	h.filter.Delete(word)
	return c.JSON(http.StatusOK, word)
}
