package resources

import (
	"context"
	"fmt"
	"net/url"

	"github.com/creiche/confluent-go/pkg/api"
	"github.com/creiche/confluent-go/pkg/client"
)

// ServiceAccountManager handles service account operations via REST API.
type ServiceAccountManager struct {
	client *client.Client
}

// NewServiceAccountManager creates a new service account manager.
func NewServiceAccountManager(c *client.Client) *ServiceAccountManager {
	return &ServiceAccountManager{client: c}
}

// ListServiceAccounts lists all service accounts in the organization.
// Returns all service accounts that the authenticated user has access to.
// Returns errors:
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsForbidden() if user lacks permissions
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (sam *ServiceAccountManager) ListServiceAccounts(ctx context.Context) ([]api.ServiceAccount, error) {
	req := client.Request{
		Method: "GET",
		Path:   "/iam/v2/service-accounts",
	}

	resp, err := sam.client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list service accounts: %w", err)
	}

	var result struct {
		Data []api.ServiceAccount `json:"data"`
	}
	if err := resp.DecodeJSON(&result); err != nil {
		return nil, fmt.Errorf("failed to parse service account list response: %w", err)
	}

	return result.Data, nil
}

// GetServiceAccount retrieves information about a specific service account.
// Returns errors:
//   - *api.Error with IsNotFound() if service account does not exist
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsForbidden() if user lacks permissions
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (sam *ServiceAccountManager) GetServiceAccount(ctx context.Context, serviceAccountID string) (*api.ServiceAccount, error) {
	req := client.Request{
		Method: "GET",
		Path:   fmt.Sprintf("/iam/v2/service-accounts/%s", serviceAccountID),
	}

	resp, err := sam.client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to describe service account %s: %w", serviceAccountID, err)
	}

	var account api.ServiceAccount
	if err := resp.DecodeJSON(&account); err != nil {
		return nil, fmt.Errorf("failed to parse service account description: %w", err)
	}

	return &account, nil
}

// CreateServiceAccount creates a new service account with the specified name and description.
// The service account can be used to authenticate applications and services.
// Returns errors:
//   - *api.Error with IsBadRequest() if parameters are invalid
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsForbidden() if user lacks permissions
//   - *api.Error with IsConflict() if service account name already exists
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (sam *ServiceAccountManager) CreateServiceAccount(ctx context.Context, name string, description string) (*api.ServiceAccount, error) {
	body := map[string]interface{}{
		"display_name": name,
		"description":  description,
	}

	req := client.Request{
		Method: "POST",
		Path:   "/iam/v2/service-accounts",
		Body:   body,
	}

	resp, err := sam.client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create service account: %w", err)
	}

	var account api.ServiceAccount
	if err := resp.DecodeJSON(&account); err != nil {
		return nil, fmt.Errorf("failed to parse create service account response: %w", err)
	}

	return &account, nil
}

// DeleteServiceAccount deletes a service account and all associated API keys.
// This operation is irreversible. All API keys for this service account will be invalidated.
// Returns errors:
//   - *api.Error with IsNotFound() if service account does not exist
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsForbidden() if user lacks permissions
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (sam *ServiceAccountManager) DeleteServiceAccount(ctx context.Context, serviceAccountID string) error {
	req := client.Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/iam/v2/service-accounts/%s", serviceAccountID),
	}

	_, err := sam.client.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete service account %s: %w", serviceAccountID, err)
	}
	return nil
}

// UpdateServiceAccount updates the display name and description of a service account.
// Returns errors:
//   - *api.Error with IsNotFound() if service account does not exist
//   - *api.Error with IsBadRequest() if parameters are invalid
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsForbidden() if user lacks permissions
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (sam *ServiceAccountManager) UpdateServiceAccount(ctx context.Context, serviceAccountID string, displayName string, description string) (*api.ServiceAccount, error) {
	body := map[string]interface{}{
		"display_name": displayName,
		"description":  description,
	}

	req := client.Request{
		Method: "PATCH",
		Path:   fmt.Sprintf("/iam/v2/service-accounts/%s", serviceAccountID),
		Body:   body,
	}

	resp, err := sam.client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update service account %s: %w", serviceAccountID, err)
	}

	var account api.ServiceAccount
	if err := resp.DecodeJSON(&account); err != nil {
		return nil, fmt.Errorf("failed to parse update response: %w", err)
	}

	return &account, nil
}

// CreateAPIKey creates a new API key for a service account.
// The API key secret is only returned once and cannot be retrieved later.
// Store it securely immediately after creation.
// Returns errors:
//   - *api.Error with IsNotFound() if service account does not exist
//   - *api.Error with IsBadRequest() if parameters are invalid
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsForbidden() if user lacks permissions
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (sam *ServiceAccountManager) CreateAPIKey(ctx context.Context, serviceAccountID string, description string) (*api.APIKey, error) {
	body := map[string]interface{}{
		"spec": map[string]interface{}{
			"owner": map[string]string{
				"id": serviceAccountID,
			},
			"description": description,
		},
	}

	req := client.Request{
		Method: "POST",
		Path:   "/iam/v2/api-keys",
		Body:   body,
	}

	resp, err := sam.client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	var apiKey api.APIKey
	if err := resp.DecodeJSON(&apiKey); err != nil {
		return nil, fmt.Errorf("failed to parse create API key response: %w", err)
	}

	return &apiKey, nil
}

// ListAPIKeys lists all API keys for a service account.
// Note: API key secrets are not included in the response.
// Returns errors:
//   - *api.Error with IsNotFound() if service account does not exist
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsForbidden() if user lacks permissions
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (sam *ServiceAccountManager) ListAPIKeys(ctx context.Context, serviceAccountID string) ([]api.APIKey, error) {
	req := client.Request{
		Method: "GET",
		Path:   fmt.Sprintf("/iam/v2/api-keys?owner=%s", url.QueryEscape(serviceAccountID)),
	}

	resp, err := sam.client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}

	var result struct {
		Data []api.APIKey `json:"data"`
	}
	if err := resp.DecodeJSON(&result); err != nil {
		return nil, fmt.Errorf("failed to parse API key list response: %w", err)
	}

	return result.Data, nil
}

// DeleteAPIKey deletes an API key.
// This operation is irreversible. The API key will be immediately invalidated.
// Returns errors:
//   - *api.Error with IsNotFound() if API key does not exist
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsForbidden() if user lacks permissions
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (sam *ServiceAccountManager) DeleteAPIKey(ctx context.Context, apiKeyID string) error {
	req := client.Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/iam/v2/api-keys/%s", apiKeyID),
	}

	_, err := sam.client.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete API key %s: %w", apiKeyID, err)
	}
	return nil
}
