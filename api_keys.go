package openrouter

import (
	"context"
	"net/http"
)

const (
	apiKeysSuffix = "/keys"
	apiKeySuffix  = "/key"
)

type KeyLimitReset string

const (
	KeyLimitResetDaily   KeyLimitReset = "daily"
	KeyLimitResetWeekly  KeyLimitReset = "weekly"
	KeyLimitResetMonthly KeyLimitReset = "monthly"
)

type APIRateLimit struct {
	Requests int    `json:"requests"`
	Interval string `json:"interval"`
	Note     string `json:"note,omitempty"`
}

type APIKey struct {
	Hash               string        `json:"hash,omitempty"`
	Name               string        `json:"name,omitempty"`
	Label              string        `json:"label,omitempty"`
	Disabled           bool          `json:"disabled,omitempty"`
	Limit              float64       `json:"limit,omitempty"`
	LimitRemaining     float64       `json:"limit_remaining,omitempty"`
	LimitReset         KeyLimitReset `json:"limit_reset,omitempty"`
	IncludeByokInLimit bool          `json:"include_byok_in_limit,omitempty"`

	Usage        float64 `json:"usage,omitempty"`
	UsageDaily   float64 `json:"usage_daily,omitempty"`
	UsageWeekly  float64 `json:"usage_weekly,omitempty"`
	UsageMonthly float64 `json:"usage_monthly,omitempty"`

	ByokUsage        float64 `json:"byok_usage,omitempty"`
	ByokUsageDaily   float64 `json:"byok_usage_daily,omitempty"`
	ByokUsageWeekly  float64 `json:"byok_usage_weekly,omitempty"`
	ByokUsageMonthly float64 `json:"byok_usage_monthly,omitempty"`

	CreatedAt string  `json:"created_at,omitempty"`
	UpdatedAt *string `json:"updated_at,omitempty"`
	ExpiresAt *string `json:"expires_at,omitempty"`
}

type APIKeyCurrent struct {
	Label              string        `json:"label,omitempty"`
	Limit              float64       `json:"limit,omitempty"`
	Usage              float64       `json:"usage,omitempty"`
	UsageDaily         float64       `json:"usage_daily,omitempty"`
	UsageWeekly        float64       `json:"usage_weekly,omitempty"`
	UsageMonthly       float64       `json:"usage_monthly,omitempty"`
	ByokUsage          float64       `json:"byok_usage,omitempty"`
	ByokUsageDaily     float64       `json:"byok_usage_daily,omitempty"`
	ByokUsageWeekly    float64       `json:"byok_usage_weekly,omitempty"`
	ByokUsageMonthly   float64       `json:"byok_usage_monthly,omitempty"`
	IsFreeTier         bool          `json:"is_free_tier,omitempty"`
	IsProvisioningKey  bool          `json:"is_provisioning_key,omitempty"`
	LimitRemaining     float64       `json:"limit_remaining,omitempty"`
	LimitReset         KeyLimitReset `json:"limit_reset,omitempty"`
	IncludeByokInLimit bool          `json:"include_byok_in_limit,omitempty"`
	RateLimit          *APIRateLimit `json:"rate_limit,omitempty"`
	ExpiresAt          *string       `json:"expires_at,omitempty"`
}

type APIKeysListResponse struct {
	Data []APIKey `json:"data"`
}

type APIKeyResponse struct {
	Data APIKey `json:"data"`
}

type APIKeyCreateResponse struct {
	Data APIKey `json:"data"`
	Key  string `json:"key,omitempty"`
}

type APIKeyCurrentResponse struct {
	Data APIKeyCurrent `json:"data"`
}

type APIKeyDeleteResponse struct {
	Deleted bool `json:"deleted"`
}

type APIKeyCreateRequest struct {
	Name               string        `json:"name,omitempty"`
	Limit              float64       `json:"limit,omitempty"`
	LimitReset         KeyLimitReset `json:"limit_reset,omitempty"`
	IncludeByokInLimit *bool         `json:"include_byok_in_limit,omitempty"`
	ExpiresAt          *string       `json:"expires_at,omitempty"`
}

type APIKeyUpdateRequest struct {
	Name               *string        `json:"name,omitempty"`
	Disabled           *bool          `json:"disabled,omitempty"`
	Limit              *float64       `json:"limit,omitempty"`
	LimitReset         *KeyLimitReset `json:"limit_reset,omitempty"`
	IncludeByokInLimit *bool          `json:"include_byok_in_limit,omitempty"`
	ExpiresAt          *string        `json:"expires_at,omitempty"`
}

// ListAPIKeys lists all API keys for the current account.
func (c *Client) ListAPIKeys(ctx context.Context) (APIKeysListResponse, error) {
	var res APIKeysListResponse

	req, err := c.newRequest(
		ctx,
		http.MethodGet,
		c.fullURL(apiKeysSuffix),
	)
	if err != nil {
		return res, err
	}

	err = c.sendRequest(req, &res)
	return res, err
}

// CreateAPIKey creates a new API key.
func (c *Client) CreateAPIKey(
	ctx context.Context,
	request APIKeyCreateRequest,
) (APIKeyCreateResponse, error) {
	var res APIKeyCreateResponse

	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(apiKeysSuffix),
		withBody(request),
	)
	if err != nil {
		return res, err
	}

	err = c.sendRequest(req, &res)
	return res, err
}

// GetAPIKey retrieves a single API key by hash.
func (c *Client) GetAPIKey(ctx context.Context, hash string) (APIKeyResponse, error) {
	var res APIKeyResponse

	req, err := c.newRequest(
		ctx,
		http.MethodGet,
		c.fullURL(apiKeysSuffix+"/"+hash),
	)
	if err != nil {
		return res, err
	}

	err = c.sendRequest(req, &res)
	return res, err
}

// DeleteAPIKey deletes a single API key by hash.
func (c *Client) DeleteAPIKey(ctx context.Context, hash string) (APIKeyDeleteResponse, error) {
	var res APIKeyDeleteResponse

	req, err := c.newRequest(
		ctx,
		http.MethodDelete,
		c.fullURL(apiKeysSuffix+"/"+hash),
	)
	if err != nil {
		return res, err
	}

	err = c.sendRequest(req, &res)
	return res, err
}

// UpdateAPIKey updates an API key by hash.
func (c *Client) UpdateAPIKey(
	ctx context.Context,
	hash string,
	request APIKeyUpdateRequest,
) (APIKeyResponse, error) {
	var res APIKeyResponse

	req, err := c.newRequest(
		ctx,
		http.MethodPatch,
		c.fullURL(apiKeysSuffix+"/"+hash),
		withBody(request),
	)
	if err != nil {
		return res, err
	}

	err = c.sendRequest(req, &res)
	return res, err
}

// GetCurrentAPIKey returns information about the API key used for this request.
func (c *Client) GetCurrentAPIKey(ctx context.Context) (APIKeyCurrentResponse, error) {
	var res APIKeyCurrentResponse

	req, err := c.newRequest(
		ctx,
		http.MethodGet,
		c.fullURL(apiKeySuffix),
	)
	if err != nil {
		return res, err
	}

	err = c.sendRequest(req, &res)
	return res, err
}
