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
	"strings"
	"sync"
	"time"

	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/models"
)

const (
	DefaultTimeout = 30 * time.Second
	APIVersion     = "v2"
	// Red Hat SSO endpoint for token refresh
	TokenEndpoint = "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token"
	ClientID      = "cloud-services"
)

// TokenResponse represents the OAuth2 token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type Client struct {
	httpClient   *http.Client
	baseURL      string
	offlineToken string
	accessToken  string
	tokenExpiry  time.Time
	tokenMutex   sync.RWMutex
}

type ClientConfig struct {
	BaseURL      string
	OfflineToken string // Changed from Token to OfflineToken
	HTTPClient   *http.Client
	Timeout      time.Duration
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
		httpClient:   config.HTTPClient,
		baseURL:      baseURL,
		offlineToken: config.OfflineToken,
	}
}

// refreshAccessToken exchanges the offline token for a new access token
func (c *Client) refreshAccessToken(ctx context.Context) error {
	if c.offlineToken == "" {
		return fmt.Errorf("no offline token provided")
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", ClientID)
	data.Set("refresh_token", c.offlineToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, TokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create token refresh request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	c.tokenMutex.Lock()
	c.accessToken = tokenResp.AccessToken
	// Set expiry with a 5-minute buffer to avoid edge cases
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn)*time.Second - 5*time.Minute)
	c.tokenMutex.Unlock()

	return nil
}

// getAccessToken returns a valid access token, refreshing if necessary
func (c *Client) getAccessToken(ctx context.Context) (string, error) {
	// For testing purposes, if offline token starts with "test-", use it directly
	if strings.HasPrefix(c.offlineToken, "test-") {
		return c.offlineToken, nil
	}

	c.tokenMutex.RLock()
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
		token := c.accessToken
		c.tokenMutex.RUnlock()
		return token, nil
	}
	c.tokenMutex.RUnlock()

	// Token is expired or doesn't exist, refresh it
	if err := c.refreshAccessToken(ctx); err != nil {
		return "", err
	}

	c.tokenMutex.RLock()
	token := c.accessToken
	c.tokenMutex.RUnlock()

	return token, nil
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

	// Get access token (will refresh if needed)
	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
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
		defer func() {
			_ = resp.Body.Close()
		}()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}

func (c *Client) unmarshalResponse(resp *http.Response, target interface{}) error {
	defer func() {
		_ = resp.Body.Close()
	}()

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

	// Get access token (will refresh if needed)
	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

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

// DownloadManifestContent downloads the content of a specific manifest file
func (c *Client) DownloadManifestContent(ctx context.Context, clusterID, fileName, folder string) (string, error) {
	if folder == "" {
		folder = "manifests"
	}

	u, _ := url.Parse(c.buildURL(fmt.Sprintf("clusters/%s/manifests/files", clusterID)))
	params := url.Values{}
	params.Add("file_name", fileName)
	params.Add("folder", folder)
	u.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	// Add authentication header
	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return "", fmt.Errorf("error getting access token: %w", err)
	}
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	req.Header.Set("Accept", "application/octet-stream")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error response %d: %s", resp.StatusCode, string(body))
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	return string(content), nil
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

	// Get access token (will refresh if needed)
	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer func() {
			_ = resp.Body.Close()
		}()
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

func (c *Client) UpdateHost(ctx context.Context, infraEnvID, hostID string, params models.HostUpdateParams) (*models.Host, error) {
	resp, err := c.doRequest(ctx, "PATCH", fmt.Sprintf("infra-envs/%s/hosts/%s", infraEnvID, hostID), params)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var host models.Host
	if err := json.NewDecoder(resp.Body).Decode(&host); err != nil {
		return nil, fmt.Errorf("failed to decode host response: %w", err)
	}
	return &host, nil
}

// Operator bundles
func (c *Client) GetOperatorBundles(ctx context.Context) (*models.Bundles, error) {
	resp, err := c.doRequest(ctx, "GET", "operators/bundles", nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var bundles models.Bundles
	if err := json.NewDecoder(resp.Body).Decode(&bundles); err != nil {
		return nil, fmt.Errorf("failed to decode bundles response: %w", err)
	}

	return &bundles, nil
}

func (c *Client) GetOperatorBundle(ctx context.Context, bundleID string) (*models.Bundle, error) {
	endpoint := fmt.Sprintf("operators/bundles/%s", bundleID)
	resp, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var bundle models.Bundle
	if err := json.NewDecoder(resp.Body).Decode(&bundle); err != nil {
		return nil, fmt.Errorf("failed to decode bundle response: %w", err)
	}

	return &bundle, nil
}

// Support levels
func (c *Client) GetSupportedFeatures(ctx context.Context, openshiftVersion, cpuArchitecture, platformType string) (*models.SupportedFeatures, error) {
	u, _ := url.Parse(c.buildURL("support-levels/features"))
	params := url.Values{}
	params.Add("openshift_version", openshiftVersion)
	if cpuArchitecture != "" {
		params.Add("cpu_architecture", cpuArchitecture)
	}
	if platformType != "" {
		params.Add("platform_type", platformType)
	}
	u.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Get access token (will refresh if needed)
	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var response models.SupportedFeaturesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode supported features response: %w", err)
	}

	return &response.Features, nil
}

func (c *Client) GetSupportedArchitectures(ctx context.Context, openshiftVersion string) (*models.SupportedArchitectures, error) {
	u, _ := url.Parse(c.buildURL("support-levels/architectures"))
	params := url.Values{}
	params.Add("openshift_version", openshiftVersion)
	u.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Get access token (will refresh if needed)
	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var response models.SupportedArchitecturesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode supported architectures response: %w", err)
	}

	return &response.Architectures, nil
}

// GetDetailedSupportedFeatures fetches detailed feature support information
func (c *Client) GetDetailedSupportedFeatures(ctx context.Context, openshiftVersion, cpuArchitecture, platformType string) (*models.DetailedSupportedFeatures, error) {
	u, _ := url.Parse(c.buildURL("support-levels/features/detailed"))
	params := url.Values{}
	params.Add("openshift_version", openshiftVersion)
	if cpuArchitecture != "" {
		params.Add("cpu_architecture", cpuArchitecture)
	}
	if platformType != "" {
		params.Add("platform_type", platformType)
	}
	u.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Get access token (will refresh if needed)
	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// The detailed endpoint returns a different structure based on swagger:
	// { "features": [...], "operators": [...] }
	// But we need to handle the "features" part as DetailedSupportedFeatures
	type DetailedResponse struct {
		Features []struct {
			FeatureSupportLevelID string                 `json:"feature-support-level-id"`
			SupportLevel          string                 `json:"support_level"`
			Incompatibilities     []string               `json:"incompatibilities,omitempty"`
			Dependencies          []string               `json:"dependencies,omitempty"`
			Properties            map[string]interface{} `json:"properties,omitempty"`
		} `json:"features"`
	}

	var detailedResp DetailedResponse
	if err := json.NewDecoder(resp.Body).Decode(&detailedResp); err != nil {
		return nil, fmt.Errorf("failed to decode detailed supported features response: %w", err)
	}

	// Convert to our model format
	features := make(models.DetailedSupportedFeatures)
	for _, feature := range detailedResp.Features {
		features[feature.FeatureSupportLevelID] = models.DetailedFeature{
			SupportLevel:      feature.SupportLevel,
			Incompatibilities: feature.Incompatibilities,
			Dependencies:      feature.Dependencies,
			Properties:        feature.Properties,
		}
	}

	return &features, nil
}

// GetClusterCredentials retrieves admin credentials for an installed cluster
func (c *Client) GetClusterCredentials(ctx context.Context, clusterID string) (*models.Credentials, error) {
	url := fmt.Sprintf("%s/%s/clusters/%s/credentials", c.baseURL, APIVersion, clusterID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Get access token (will refresh if needed)
	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var credentials models.Credentials
	if err := json.NewDecoder(resp.Body).Decode(&credentials); err != nil {
		return nil, fmt.Errorf("failed to decode credentials response: %w", err)
	}

	return &credentials, nil
}

// GetClusterEvents retrieves events for a cluster with optional filtering
func (c *Client) GetClusterEvents(ctx context.Context, clusterID string, params map[string]string) (*models.EventsResponse, error) {
	baseURL := fmt.Sprintf("%s/%s/events", c.baseURL, APIVersion)
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add cluster_id to query parameters
	query := u.Query()
	if clusterID != "" {
		query.Set("cluster_id", clusterID)
	}

	// Add optional parameters
	for key, value := range params {
		if value != "" {
			query.Set(key, value)
		}
	}

	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Get access token (will refresh if needed)
	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var events models.EventsResponse
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, fmt.Errorf("failed to decode events response: %w", err)
	}

	return &events, nil
}

// DownloadClusterCredentialFile downloads a specific credential file (kubeconfig, kubeadmin-password, etc.)
func (c *Client) DownloadClusterCredentialFile(ctx context.Context, clusterID, fileName string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s/clusters/%s/downloads/credentials?file_name=%s", c.baseURL, APIVersion, clusterID, fileName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Get access token (will refresh if needed)
	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Read the file content
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return content, nil
}

// GetClusterValidations retrieves validation information for a cluster
func (c *Client) GetClusterValidations(ctx context.Context, clusterID string) (*models.ClusterValidationResponse, error) {
	url := fmt.Sprintf("%s/%s/clusters/%s", c.baseURL, APIVersion, clusterID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Get access token (will refresh if needed)
	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse the cluster response to extract validations_info
	var clusterResp struct {
		ValidationsInfo map[string][]models.ValidationInfo `json:"validations_info"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&clusterResp); err != nil {
		return nil, fmt.Errorf("failed to decode cluster validation response: %w", err)
	}

	return &models.ClusterValidationResponse{
		ValidationsInfo: clusterResp.ValidationsInfo,
	}, nil
}

// GetHostValidations retrieves validation information for all hosts in a cluster
func (c *Client) GetHostValidations(ctx context.Context, clusterID string) (*models.HostsValidationResponse, error) {
	url := fmt.Sprintf("%s/%s/clusters/%s/hosts", c.baseURL, APIVersion, clusterID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Get access token (will refresh if needed)
	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse the hosts response to extract validations_info from each host
	var hostsResp []struct {
		ID              string                             `json:"id"`
		ValidationsInfo map[string][]models.ValidationInfo `json:"validations_info"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&hostsResp); err != nil {
		return nil, fmt.Errorf("failed to decode host validations response: %w", err)
	}

	// Convert to our response format
	hosts := make([]models.HostValidationResponse, len(hostsResp))
	for i, host := range hostsResp {
		hosts[i] = models.HostValidationResponse{
			ID:              host.ID,
			ValidationsInfo: host.ValidationsInfo,
		}
	}

	return &models.HostsValidationResponse{
		Hosts: hosts,
	}, nil
}

// GetSingleHostValidations retrieves validation information for a specific host
func (c *Client) GetSingleHostValidations(ctx context.Context, infraEnvID, hostID string) (*models.HostValidationResponse, error) {
	url := fmt.Sprintf("%s/%s/infra-envs/%s/hosts/%s", c.baseURL, APIVersion, infraEnvID, hostID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Get access token (will refresh if needed)
	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse the host response to extract validations_info
	var hostResp struct {
		ID              string                             `json:"id"`
		ValidationsInfo map[string][]models.ValidationInfo `json:"validations_info"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&hostResp); err != nil {
		return nil, fmt.Errorf("failed to decode host validation response: %w", err)
	}

	return &models.HostValidationResponse{
		ID:              hostResp.ID,
		ValidationsInfo: hostResp.ValidationsInfo,
	}, nil
}

// DownloadClusterLogs downloads cluster logs with optional filtering
func (c *Client) DownloadClusterLogs(ctx context.Context, clusterID string, params map[string]string) ([]byte, error) {
	baseURL := fmt.Sprintf("%s/%s/clusters/%s/logs", c.baseURL, APIVersion, clusterID)
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add optional parameters
	query := u.Query()
	for key, value := range params {
		if value != "" {
			query.Set(key, value)
		}
	}
	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Get access token (will refresh if needed)
	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Read the log content
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return content, nil
}

// DownloadClusterFiles downloads various cluster files (ignition configs, manifests, logs, etc.)
func (c *Client) DownloadClusterFiles(ctx context.Context, clusterID, fileName string, params map[string]string) ([]byte, error) {
	baseURL := fmt.Sprintf("%s/%s/clusters/%s/downloads/files", c.baseURL, APIVersion, clusterID)
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add file_name and optional parameters
	query := u.Query()
	query.Set("file_name", fileName)
	for key, value := range params {
		if value != "" {
			query.Set(key, value)
		}
	}
	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Get access token (will refresh if needed)
	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Read the file content
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return content, nil
}
