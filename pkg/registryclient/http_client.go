package registryclient

import (
	"context"
	"encoding/json"

	"github.com/cockroachdb/errors"
	"github.com/go-resty/resty/v2"

	"github.com/izuc/asset-registry/pkg/registry"
	"github.com/izuc/asset-registry/pkg/registry/registryhttp"
)

type HTTPClient struct {
	client *resty.Client
}

func NewHTTPClient(restyClient *resty.Client) *HTTPClient {
	return &HTTPClient{client: restyClient}
}

func (c *HTTPClient) SaveAsset(ctx context.Context, network string, asset *registry.Asset) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(asset).
		SetError(&registryhttp.ErrorResponse{}).
		Post(registryhttp.RegistriesEndpoint + "/" + network + registryhttp.AssetsEndpoint)
	if err != nil {
		return errors.Wrap(err, "failed to execute saveAssets HTTP call")
	}
	if resp.IsSuccess() {
		return nil
	}
	errorResp := resp.Error().(*registryhttp.ErrorResponse)
	return errors.Newf("saveAssets HTTP call returns an error: %s", errorResp.Error)
}

func (c *HTTPClient) LoadAssets(ctx context.Context, network string, assets ...string) error {
	assetID := ""
	if len(assets) > 0 {
		assetID = "/" + assets[0]
	}
	resp, err := c.client.R().
		SetContext(ctx).
		SetError(&registryhttp.ErrorResponse{}).
		Get(registryhttp.RegistriesEndpoint + "/" + network + registryhttp.AssetsEndpoint + assetID)
	if err != nil {
		return errors.Wrap(err, "failed to execute loadAssets HTTP call")
	}
	if resp.IsSuccess() {
		return nil
	}
	errorResp := resp.Error().(*registryhttp.ErrorResponse)
	return errors.Newf("loadAssets HTTP call returns an error: %s", errorResp.Error)
}

func (c *HTTPClient) LoadAsset(ctx context.Context, network string, asset string) (*registry.Asset, error) {
	assetID := "/" + asset
	resp, err := c.client.R().
		SetContext(ctx).
		SetError(&registryhttp.ErrorResponse{}).
		Get(registryhttp.RegistriesEndpoint + "/" + network + registryhttp.AssetsEndpoint + assetID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute loadAssets HTTP call")
	}
	assetStruct :=  &registry.Asset{}
	if resp.IsSuccess() {
		parseErr := json.Unmarshal(resp.Body(), assetStruct)
		if parseErr != nil {
			return nil, errors.Errorf("failed to parse asset in response body: %w", parseErr)
		}
		return assetStruct, nil
	}
	errorResp := resp.Error().(*registryhttp.ErrorResponse)
	return nil, errors.Newf("loadAsset HTTP call returns an error: %s", errorResp.Error)
}
