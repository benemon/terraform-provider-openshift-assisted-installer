package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/models"
)

const (
	DefaultTimeout = 30 * time.Second
	APIVersion     = "v2"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

type ClientConfig struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
	Timeout    time.Duration
}

func NewClient(config ClientConfig) *Client {
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{
			Timeout: config.Timeout,
		}
		if config.HTTPClient.Timeout == 0 {
			config.HTTPClient.Timeout = DefaultTimeout
		}
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openshift.com/api/assisted-install"
	}

	return &Client{
		httpClient: config.HTTPClient,
		baseURL:    baseURL,
		token:      config.Token,
	}
}

func (c *Client) buildURL(endpoint string) string {
	u, _ := url.Parse(c.baseURL)
	u.Path = path.Join(u.Path, APIVersion, endpoint)
	return u.String()
}

func (c *Client) doRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.buildURL(endpoint), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}

func (c *Client) unmarshalResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()
	
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	return nil
}

// Cluster operations
func (c *Client) CreateCluster(ctx context.Context, params models.ClusterCreateParams) (*models.Cluster, error) {
	resp, err := c.doRequest(ctx, "POST", "clusters", params)
	if err != nil {
		return nil, err
	}

	var cluster models.Cluster
	if err := c.unmarshalResponse(resp, &cluster); err != nil {
		return nil, err
	}

	return &cluster, nil
}

func (c *Client) GetCluster(ctx context.Context, clusterID string) (*models.Cluster, error) {
	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("clusters/%s", clusterID), nil)
	if err != nil {
		return nil, err
	}

	var cluster models.Cluster
	if err := c.unmarshalResponse(resp, &cluster); err != nil {
		return nil, err
	}

	return &cluster, nil
}

func (c *Client) UpdateCluster(ctx context.Context, clusterID string, params models.ClusterUpdateParams) (*models.Cluster, error) {
	resp, err := c.doRequest(ctx, "PATCH", fmt.Sprintf("clusters/%s", clusterID), params)
	if err != nil {
		return nil, err
	}

	var cluster models.Cluster
	if err := c.unmarshalResponse(resp, &cluster); err != nil {
		return nil, err
	}

	return &cluster, nil
}

func (c *Client) DeleteCluster(ctx context.Context, clusterID string) error {
	_, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("clusters/%s", clusterID), nil)
	return err
}

func (c *Client) InstallCluster(ctx context.Context, clusterID string) error {
	_, err := c.doRequest(ctx, "POST", fmt.Sprintf("clusters/%s/actions/install", clusterID), nil)
	return err
}

func (c *Client) ListClusters(ctx context.Context) ([]models.Cluster, error) {
	resp, err := c.doRequest(ctx, "GET", "clusters", nil)
	if err != nil {
		return nil, err
	}

	var clusters []models.Cluster
	if err := c.unmarshalResponse(resp, &clusters); err != nil {
		return nil, err
	}

	return clusters, nil
}

// InfraEnv operations
func (c *Client) CreateInfraEnv(ctx context.Context, params models.InfraEnvCreateParams) (*models.InfraEnv, error) {
	resp, err := c.doRequest(ctx, "POST", "infra-envs", params)
	if err != nil {
		return nil, err
	}

	var infraEnv models.InfraEnv
	if err := c.unmarshalResponse(resp, &infraEnv); err != nil {
		return nil, err
	}

	return &infraEnv, nil
}

func (c *Client) GetInfraEnv(ctx context.Context, infraEnvID string) (*models.InfraEnv, error) {
	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("infra-envs/%s", infraEnvID), nil)
	if err != nil {
		return nil, err
	}

	var infraEnv models.InfraEnv
	if err := c.unmarshalResponse(resp, &infraEnv); err != nil {
		return nil, err
	}

	return &infraEnv, nil
}

func (c *Client) UpdateInfraEnv(ctx context.Context, infraEnvID string, params models.InfraEnvUpdateParams) (*models.InfraEnv, error) {
	resp, err := c.doRequest(ctx, "PATCH", fmt.Sprintf("infra-envs/%s", infraEnvID), params)
	if err != nil {
		return nil, err
	}

	var infraEnv models.InfraEnv
	if err := c.unmarshalResponse(resp, &infraEnv); err != nil {
		return nil, err
	}

	return &infraEnv, nil
}

func (c *Client) DeleteInfraEnv(ctx context.Context, infraEnvID string) error {
	_, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("infra-envs/%s", infraEnvID), nil)
	return err
}

func (c *Client) ListInfraEnvs(ctx context.Context) ([]models.InfraEnv, error) {
	resp, err := c.doRequest(ctx, "GET", "infra-envs", nil)
	if err != nil {
		return nil, err
	}

	var infraEnvs []models.InfraEnv
	if err := c.unmarshalResponse(resp, &infraEnvs); err != nil {
		return nil, err
	}

	return infraEnvs, nil
}

// Manifest operations
func (c *Client) CreateManifest(ctx context.Context, clusterID string, params models.CreateManifestParams) error {
	_, err := c.doRequest(ctx, "POST", fmt.Sprintf("clusters/%s/manifests", clusterID), params)
	return err
}

func (c *Client) UpdateManifest(ctx context.Context, clusterID string, params models.UpdateManifestParams) error {
	_, err := c.doRequest(ctx, "PUT", fmt.Sprintf("clusters/%s/manifests", clusterID), params)
	return err
}

func (c *Client) DeleteManifest(ctx context.Context, clusterID string, folder, fileName string) error {
	u, _ := url.Parse(c.buildURL(fmt.Sprintf("clusters/%s/manifests", clusterID)))
	q := u.Query()
	q.Set("folder", folder)
	q.Set("file_name", fileName)
	u.RawQuery = q.Encode()
	
	req, err := http.NewRequestWithContext(ctx, "DELETE", u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (c *Client) ListManifests(ctx context.Context, clusterID string) ([]models.Manifest, error) {
	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("clusters/%s/manifests", clusterID), nil)
	if err != nil {
		return nil, err
	}

	var manifests []models.Manifest
	if err := c.unmarshalResponse(resp, &manifests); err != nil {
		return nil, err
	}

	return manifests, nil
}

// OpenShift versions
func (c *Client) GetOpenShiftVersions(ctx context.Context, version string, onlyLatest bool) (*models.OpenshiftVersions, error) {
	u, _ := url.Parse(c.buildURL("openshift-versions"))
	params := url.Values{}
	if version != "" {
		params.Add("version", version)
	}
	if onlyLatest {
		params.Add("only_latest", "true")
	}
	if len(params) > 0 {
		u.RawQuery = params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var versions models.OpenshiftVersions
	if err := c.unmarshalResponse(resp, &versions); err != nil {
		return nil, err
	}

	return &versions, nil
}

// Supported operators
func (c *Client) GetSupportedOperators(ctx context.Context) ([]string, error) {
	resp, err := c.doRequest(ctx, "GET", "supported-operators", nil)
	if err != nil {
		return nil, err
	}

	var operators []string
	if err := c.unmarshalResponse(resp, &operators); err != nil {
		return nil, err
	}

	return operators, nil
}

// Host operations
func (c *Client) ListHosts(ctx context.Context, infraEnvID string) ([]models.Host, error) {
	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("infra-envs/%s/hosts", infraEnvID), nil)
	if err != nil {
		return nil, err
	}

	var hosts []models.Host
	if err := c.unmarshalResponse(resp, &hosts); err != nil {
		return nil, err
	}

	return hosts, nil
}

func (c *Client) GetHost(ctx context.Context, infraEnvID, hostID string) (*models.Host, error) {
	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("infra-envs/%s/hosts/%s", infraEnvID, hostID), nil)
	if err != nil {
		return nil, err
	}

	var host models.Host
	if err := c.unmarshalResponse(resp, &host); err != nil {
		return nil, err
	}

	return &host, nil
}

func (c *Client) BindHost(ctx context.Context, infraEnvID, hostID string, params models.BindHostParams) error {
	_, err := c.doRequest(ctx, "POST", fmt.Sprintf("infra-envs/%s/hosts/%s/actions/bind", infraEnvID, hostID), params)
	return err
}

func (c *Client) UnbindHost(ctx context.Context, infraEnvID, hostID string) error {
	_, err := c.doRequest(ctx, "POST", fmt.Sprintf("infra-envs/%s/hosts/%s/actions/unbind", infraEnvID, hostID), nil)
	return err
}