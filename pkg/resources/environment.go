package resources

import (
	"context"
	"fmt"

	"github.com/creiche/confluent-go/pkg/api"
	"github.com/creiche/confluent-go/pkg/client"
)

// EnvironmentManager handles environment-related operations via REST API.
type EnvironmentManager struct {
	client *client.Client
}

// NewEnvironmentManager creates a new environment manager.
func NewEnvironmentManager(c *client.Client) *EnvironmentManager {
	return &EnvironmentManager{client: c}
}

// ListEnvironments lists all environments in the organization.
// Returns all environments that the authenticated user has access to.
// Returns errors:
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsForbidden() if user lacks permissions
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (em *EnvironmentManager) ListEnvironments(ctx context.Context) ([]api.Environment, error) {
	req := client.Request{
		Method: "GET",
		Path:   "/org/v2/environments",
	}

	resp, err := em.client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list environments: %w", err)
	}

	var result struct {
		Data []api.Environment `json:"data"`
	}
	if err := resp.DecodeJSON(&result); err != nil {
		return nil, fmt.Errorf("failed to parse environment list response: %w", err)
	}

	return result.Data, nil
}

// GetEnvironment retrieves information about a specific environment.
// Returns errors:
//   - *api.Error with IsNotFound() if environment does not exist
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsForbidden() if user lacks permissions
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (em *EnvironmentManager) GetEnvironment(ctx context.Context, environmentID string) (*api.Environment, error) {
	req := client.Request{
		Method: "GET",
		Path:   fmt.Sprintf("/org/v2/environments/%s", environmentID),
	}

	resp, err := em.client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to describe environment %s: %w", environmentID, err)
	}

	var environment api.Environment
	if err := resp.DecodeJSON(&environment); err != nil {
		return nil, fmt.Errorf("failed to parse environment description: %w", err)
	}

	return &environment, nil
}

// CreateEnvironment creates a new environment with the specified name and display name.
// Environments are logical groupings for clusters and other resources.
// Returns errors:
//   - *api.Error with IsBadRequest() if parameters are invalid
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsForbidden() if user lacks permissions
//   - *api.Error with IsConflict() if environment name already exists
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (em *EnvironmentManager) CreateEnvironment(ctx context.Context, name string, displayName string) (*api.Environment, error) {
	body := map[string]interface{}{
		"display_name": displayName,
	}
	if name != "" {
		body["name"] = name
	}

	req := client.Request{
		Method: "POST",
		Path:   "/org/v2/environments",
		Body:   body,
	}

	resp, err := em.client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create environment: %w", err)
	}

	var environment api.Environment
	if err := resp.DecodeJSON(&environment); err != nil {
		return nil, fmt.Errorf("failed to parse create environment response: %w", err)
	}

	return &environment, nil
}

// DeleteEnvironment deletes an environment.
// This operation is irreversible. All clusters and resources within the environment must be deleted first.
// Returns errors:
//   - *api.Error with IsNotFound() if environment does not exist
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsForbidden() if user lacks permissions
//   - *api.Error with IsConflict() if environment contains resources
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (em *EnvironmentManager) DeleteEnvironment(ctx context.Context, environmentID string) error {
	req := client.Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/org/v2/environments/%s", environmentID),
	}

	_, err := em.client.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete environment %s: %w", environmentID, err)
	}
	return nil
}

// UpdateEnvironment updates the display name of an environment.
// Returns errors:
//   - *api.Error with IsNotFound() if environment does not exist
//   - *api.Error with IsBadRequest() if parameters are invalid
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsForbidden() if user lacks permissions
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (em *EnvironmentManager) UpdateEnvironment(ctx context.Context, environmentID string, displayName string) (*api.Environment, error) {
	body := map[string]interface{}{
		"display_name": displayName,
	}

	req := client.Request{
		Method: "PATCH",
		Path:   fmt.Sprintf("/org/v2/environments/%s", environmentID),
		Body:   body,
	}

	resp, err := em.client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update environment %s: %w", environmentID, err)
	}

	var environment api.Environment
	if err := resp.DecodeJSON(&environment); err != nil {
		return nil, fmt.Errorf("failed to parse update response: %w", err)
	}

	return &environment, nil
}
