package resources

import (
	"context"
	"fmt"

	"github.com/creiche/confluent-go/pkg/api"
	"github.com/creiche/confluent-go/pkg/client"
)

// ClusterManager handles cluster-related operations via REST API.
type ClusterManager struct {
	client *client.Client
}

// NewClusterManager creates a new cluster manager.
func NewClusterManager(c *client.Client) *ClusterManager {
	return &ClusterManager{client: c}
}

// ListClusters lists all Kafka clusters in the environment.
// Returns errors:
//   - *api.Error with IsNotFound() for invalid environment ID
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsRateLimited() if rate limit is exceeded
//   - *api.Error with IsInternalServerError() for server-side errors
func (cm *ClusterManager) ListClusters(ctx context.Context, environmentID string) ([]api.Cluster, error) {
	req := client.Request{
		Method: "GET",
		Path:   fmt.Sprintf("/cmk/v2/clusters?environment=%s", environmentID),
	}

	resp, err := cm.client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list clusters: %w", err)
	}

	var result struct {
		Data []api.Cluster `json:"data"`
	}
	if err := resp.DecodeJSON(&result); err != nil {
		return nil, fmt.Errorf("failed to parse cluster list response: %w", err)
	}

	return result.Data, nil
}

// GetCluster retrieves information about a specific cluster.
// Returns errors:
//   - *api.Error with IsNotFound() if cluster does not exist
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsForbidden() if user lacks permissions
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (cm *ClusterManager) GetCluster(ctx context.Context, clusterID string) (*api.Cluster, error) {
	req := client.Request{
		Method: "GET",
		Path:   fmt.Sprintf("/cmk/v2/clusters/%s", clusterID),
	}

	resp, err := cm.client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to describe cluster %s: %w", clusterID, err)
	}

	var cluster api.Cluster
	if err := resp.DecodeJSON(&cluster); err != nil {
		return nil, fmt.Errorf("failed to parse cluster description: %w", err)
	}

	return &cluster, nil
}

// CreateCluster creates a new Kafka cluster.
// Returns errors:
//   - *api.Error with IsBadRequest() if parameters are invalid
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsForbidden() if user lacks permissions
//   - *api.Error with IsConflict() if cluster name already exists
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (cm *ClusterManager) CreateCluster(ctx context.Context, environmentID string, name string, clusterType string, cloud string, region string) (*api.Cluster, error) {
	body := map[string]interface{}{
		"display_name": name,
		"spec": map[string]interface{}{
			"kafka_cluster": map[string]interface{}{
				"type": clusterType,
			},
			"environment": map[string]string{
				"id": environmentID,
			},
			"network": map[string]interface{}{
				"cloud":  cloud,
				"region": region,
			},
		},
	}

	req := client.Request{
		Method: "POST",
		Path:   "/cmk/v2/clusters",
		Body:   body,
	}

	resp, err := cm.client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster: %w", err)
	}

	var cluster api.Cluster
	if err := resp.DecodeJSON(&cluster); err != nil {
		return nil, fmt.Errorf("failed to parse create cluster response: %w", err)
	}

	return &cluster, nil
}

// DeleteCluster deletes a Kafka cluster.
// Returns errors:
//   - *api.Error with IsNotFound() if cluster does not exist
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsForbidden() if user lacks permissions
//   - *api.Error with IsConflict() if cluster is not in a deletable state
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (cm *ClusterManager) DeleteCluster(ctx context.Context, clusterID string) error {
	req := client.Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/cmk/v2/clusters/%s", clusterID),
	}

	_, err := cm.client.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete cluster %s: %w", clusterID, err)
	}
	return nil
}

// UpdateCluster updates cluster configuration (e.g., name).
// Returns errors:
//   - *api.Error with IsNotFound() if cluster does not exist
//   - *api.Error with IsBadRequest() if parameters are invalid
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsForbidden() if user lacks permissions
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (cm *ClusterManager) UpdateCluster(ctx context.Context, clusterID string, displayName string) (*api.Cluster, error) {
	body := map[string]interface{}{
		"display_name": displayName,
	}

	req := client.Request{
		Method: "PATCH",
		Path:   fmt.Sprintf("/cmk/v2/clusters/%s", clusterID),
		Body:   body,
	}

	resp, err := cm.client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update cluster %s: %w", clusterID, err)
	}

	var cluster api.Cluster
	if err := resp.DecodeJSON(&cluster); err != nil {
		return nil, fmt.Errorf("failed to parse update response: %w", err)
	}

	return &cluster, nil
}
