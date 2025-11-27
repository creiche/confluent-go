package resources

import (
	"context"
	"fmt"

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

// CreateServiceAccount creates a new service account.
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

// DeleteServiceAccount deletes a service account.
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

// UpdateServiceAccount updates a service account.
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
func (sam *ServiceAccountManager) ListAPIKeys(ctx context.Context, serviceAccountID string) ([]api.APIKey, error) {
	req := client.Request{
		Method: "GET",
		Path:   fmt.Sprintf("/iam/v2/api-keys?owner=%s", serviceAccountID),
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
