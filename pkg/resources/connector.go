package resources

import (
	"context"
	"fmt"
	"net/url"

	"github.com/creiche/confluent-go/pkg/api"
	"github.com/creiche/confluent-go/pkg/client"
)

// ConnectorManager handles Kafka Connect connector operations via REST API.
type ConnectorManager struct {
	client *client.Client
}

// NewConnectorManager creates a new connector manager.
func NewConnectorManager(c *client.Client) *ConnectorManager {
	return &ConnectorManager{client: c}
}

// ListConnectors lists all connectors in a Kafka Connect cluster.
// Returns errors:
//   - *api.Error with IsNotFound() if connect cluster does not exist
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (cm *ConnectorManager) ListConnectors(ctx context.Context, environmentID string, clusterID string) ([]string, error) {
	req := client.Request{
		Method: "GET",
		Path:   fmt.Sprintf("/connect/v1/environments/%s/clusters/%s/connectors", environmentID, clusterID),
	}

	resp, err := cm.client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list connectors: %w", err)
	}

	var connectors []string
	if err := resp.DecodeJSON(&connectors); err != nil {
		return nil, fmt.Errorf("failed to parse connector list response: %w", err)
	}

	return connectors, nil
}

// GetConnector retrieves information about a specific connector.
// Returns errors:
//   - *api.Error with IsNotFound() if connector does not exist
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsForbidden() if user lacks permissions
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (cm *ConnectorManager) GetConnector(ctx context.Context, environmentID string, clusterID string, connectorName string) (*api.ConnectorConfig, error) {
	req := client.Request{
		Method: "GET",
		Path:   fmt.Sprintf("/connect/v1/environments/%s/clusters/%s/connectors/%s", environmentID, clusterID, connectorName),
	}

	resp, err := cm.client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get connector %s: %w", connectorName, err)
	}

	var connector api.ConnectorConfig
	if err := resp.DecodeJSON(&connector); err != nil {
		return nil, fmt.Errorf("failed to parse connector response: %w", err)
	}

	return &connector, nil
}

// CreateConnector creates a new Kafka Connect connector.
// The config map must include "connector.class" and other connector-specific settings.
// Returns errors:
//   - *api.Error with IsBadRequest() if parameters are invalid
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsConflict() if connector name already exists
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (cm *ConnectorManager) CreateConnector(ctx context.Context, environmentID string, clusterID string, name string, config map[string]string) (*api.ConnectorConfig, error) {
	body := map[string]interface{}{
		"name":   name,
		"config": config,
	}

	req := client.Request{
		Method: "POST",
		Path:   fmt.Sprintf("/connect/v1/environments/%s/clusters/%s/connectors", environmentID, clusterID),
		Body:   body,
	}

	resp, err := cm.client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create connector %s: %w", name, err)
	}

	var connector api.ConnectorConfig
	if err := resp.DecodeJSON(&connector); err != nil {
		return nil, fmt.Errorf("failed to parse create connector response: %w", err)
	}

	return &connector, nil
}

// UpdateConnector updates an existing connector's configuration.
// The new config will replace the existing configuration entirely.
// Returns errors:
//   - *api.Error with IsBadRequest() if config values are invalid
//   - *api.Error with IsNotFound() if connector does not exist
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (cm *ConnectorManager) UpdateConnector(ctx context.Context, environmentID string, clusterID string, connectorName string, config map[string]string) (*api.ConnectorConfig, error) {
	req := client.Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/connect/v1/environments/%s/clusters/%s/connectors/%s/config", environmentID, clusterID, connectorName),
		Body:   config,
	}

	resp, err := cm.client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update connector %s: %w", connectorName, err)
	}

	var connector api.ConnectorConfig
	if err := resp.DecodeJSON(&connector); err != nil {
		return nil, fmt.Errorf("failed to parse update connector response: %w", err)
	}

	return &connector, nil
}

// DeleteConnector deletes a connector and stops all its tasks.
// Returns errors:
//   - *api.Error with IsNotFound() if connector does not exist
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsForbidden() if user lacks permissions
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (cm *ConnectorManager) DeleteConnector(ctx context.Context, environmentID string, clusterID string, connectorName string) error {
	req := client.Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/connect/v1/environments/%s/clusters/%s/connectors/%s", environmentID, clusterID, connectorName),
	}

	_, err := cm.client.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete connector %s: %w", connectorName, err)
	}
	return nil
}

// GetConnectorStatus retrieves the current status of a connector and its tasks.
// Returns errors:
//   - *api.Error with IsNotFound() if connector does not exist
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (cm *ConnectorManager) GetConnectorStatus(ctx context.Context, environmentID string, clusterID string, connectorName string) (*api.ConnectorStatus, error) {
	req := client.Request{
		Method: "GET",
		Path:   fmt.Sprintf("/connect/v1/environments/%s/clusters/%s/connectors/%s/status", environmentID, clusterID, connectorName),
	}

	resp, err := cm.client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get connector status: %w", err)
	}

	var status api.ConnectorStatus
	if err := resp.DecodeJSON(&status); err != nil {
		return nil, fmt.Errorf("failed to parse connector status response: %w", err)
	}

	return &status, nil
}

// PauseConnector pauses a connector and its tasks.
// The connector will stop processing but retain its configuration.
// Returns errors:
//   - *api.Error with IsNotFound() if connector does not exist
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsForbidden() if user lacks permissions
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (cm *ConnectorManager) PauseConnector(ctx context.Context, environmentID string, clusterID string, connectorName string) error {
	req := client.Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/connect/v1/environments/%s/clusters/%s/connectors/%s/pause", environmentID, clusterID, connectorName),
	}

	_, err := cm.client.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to pause connector %s: %w", connectorName, err)
	}
	return nil
}

// ResumeConnector resumes a paused connector.
// The connector will restart processing from where it left off.
// Returns errors:
//   - *api.Error with IsNotFound() if connector does not exist
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsForbidden() if user lacks permissions
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (cm *ConnectorManager) ResumeConnector(ctx context.Context, environmentID string, clusterID string, connectorName string) error {
	req := client.Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/connect/v1/environments/%s/clusters/%s/connectors/%s/resume", environmentID, clusterID, connectorName),
	}

	_, err := cm.client.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to resume connector %s: %w", connectorName, err)
	}
	return nil
}

// RestartConnector restarts a connector and its tasks.
// This can be useful to recover from transient failures.
// Returns errors:
//   - *api.Error with IsNotFound() if connector does not exist
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsForbidden() if user lacks permissions
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (cm *ConnectorManager) RestartConnector(ctx context.Context, environmentID string, clusterID string, connectorName string) error {
	req := client.Request{
		Method: "POST",
		Path:   fmt.Sprintf("/connect/v1/environments/%s/clusters/%s/connectors/%s/restart", environmentID, clusterID, connectorName),
	}

	_, err := cm.client.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to restart connector %s: %w", connectorName, err)
	}
	return nil
}

// RestartTask restarts a specific task for a connector.
// Returns errors:
//   - *api.Error with IsNotFound() if connector or task does not exist
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsForbidden() if user lacks permissions
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (cm *ConnectorManager) RestartTask(ctx context.Context, environmentID string, clusterID string, connectorName string, taskID int32) error {
	req := client.Request{
		Method: "POST",
		Path:   fmt.Sprintf("/connect/v1/environments/%s/clusters/%s/connectors/%s/tasks/%d/restart", environmentID, clusterID, connectorName, taskID),
	}

	_, err := cm.client.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to restart task %d for connector %s: %w", taskID, connectorName, err)
	}
	return nil
}

// GetConnectorConfig retrieves the configuration for a connector.
// Returns errors:
//   - *api.Error with IsNotFound() if connector does not exist
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (cm *ConnectorManager) GetConnectorConfig(ctx context.Context, environmentID string, clusterID string, connectorName string) (map[string]string, error) {
	req := client.Request{
		Method: "GET",
		Path:   fmt.Sprintf("/connect/v1/environments/%s/clusters/%s/connectors/%s/config", environmentID, clusterID, connectorName),
	}

	resp, err := cm.client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get connector config: %w", err)
	}

	var config map[string]string
	if err := resp.DecodeJSON(&config); err != nil {
		return nil, fmt.Errorf("failed to parse connector config response: %w", err)
	}

	return config, nil
}

// ListConnectorPlugins lists all available connector plugins in the Connect cluster.
// Returns errors:
//   - *api.Error with IsNotFound() if connect cluster does not exist
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (cm *ConnectorManager) ListConnectorPlugins(ctx context.Context, environmentID string, clusterID string) ([]api.ConnectorPlugin, error) {
	req := client.Request{
		Method: "GET",
		Path:   fmt.Sprintf("/connect/v1/environments/%s/clusters/%s/connector-plugins", environmentID, clusterID),
	}

	resp, err := cm.client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list connector plugins: %w", err)
	}

	var plugins []api.ConnectorPlugin
	if err := resp.DecodeJSON(&plugins); err != nil {
		return nil, fmt.Errorf("failed to parse connector plugins response: %w", err)
	}

	return plugins, nil
}

// ValidateConnectorConfig validates a connector configuration without creating it.
// Returns validation errors and suggested values.
// Returns errors:
//   - *api.Error with IsBadRequest() if config is invalid
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (cm *ConnectorManager) ValidateConnectorConfig(ctx context.Context, environmentID string, clusterID string, connectorClass string, config map[string]string) (*api.ConnectorValidation, error) {
	if _, exists := config["connector.class"]; exists {
		return nil, fmt.Errorf("connector.class should not be included in config map, use connectorClass parameter instead")
	}

	body := map[string]interface{}{
		"connector.class": connectorClass,
	}
	for k, v := range config {
		body[k] = v
	}

	req := client.Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/connect/v1/environments/%s/clusters/%s/connector-plugins/%s/config/validate", environmentID, clusterID, url.PathEscape(connectorClass)),
		Body:   body,
	}

	resp, err := cm.client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate connector config: %w", err)
	}

	var validation api.ConnectorValidation
	if err := resp.DecodeJSON(&validation); err != nil {
		return nil, fmt.Errorf("failed to parse validation response: %w", err)
	}

	return &validation, nil
}

// GetConnectorTasks retrieves the list of tasks for a connector.
// Returns errors:
//   - *api.Error with IsNotFound() if connector does not exist
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (cm *ConnectorManager) GetConnectorTasks(ctx context.Context, environmentID string, clusterID string, connectorName string) ([]api.ConnectorTask, error) {
	req := client.Request{
		Method: "GET",
		Path:   fmt.Sprintf("/connect/v1/environments/%s/clusters/%s/connectors/%s/tasks", environmentID, clusterID, connectorName),
	}

	resp, err := cm.client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get connector tasks: %w", err)
	}

	var tasks []api.ConnectorTask
	if err := resp.DecodeJSON(&tasks); err != nil {
		return nil, fmt.Errorf("failed to parse connector tasks response: %w", err)
	}

	return tasks, nil
}

// GetTaskStatus retrieves the status of a specific connector task.
// Returns errors:
//   - *api.Error with IsNotFound() if connector or task does not exist
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (cm *ConnectorManager) GetTaskStatus(ctx context.Context, environmentID string, clusterID string, connectorName string, taskID int32) (*api.TaskStatus, error) {
	req := client.Request{
		Method: "GET",
		Path:   fmt.Sprintf("/connect/v1/environments/%s/clusters/%s/connectors/%s/tasks/%d/status", environmentID, clusterID, connectorName, taskID),
	}

	resp, err := cm.client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get task status: %w", err)
	}

	var status api.TaskStatus
	if err := resp.DecodeJSON(&status); err != nil {
		return nil, fmt.Errorf("failed to parse task status response: %w", err)
	}

	return &status, nil
}
