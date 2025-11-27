package resources

import (
	"context"
	"fmt"

	"github.com/creiche/confluent-go/pkg/api"
	"github.com/creiche/confluent-go/pkg/client"
)

// TopicManager handles topic-related operations via REST API.
type TopicManager struct {
	client *client.Client
}

// NewTopicManager creates a new topic manager.
func NewTopicManager(c *client.Client) *TopicManager {
	return &TopicManager{client: c}
}

// ListTopics lists all topics in a cluster.
// Returns errors:
//   - *api.Error with IsNotFound() if cluster does not exist
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (tm *TopicManager) ListTopics(ctx context.Context, clusterID string) ([]api.Topic, error) {
	req := client.Request{
		Method: "GET",
		Path:   fmt.Sprintf("/kafka/v3/clusters/%s/topics", clusterID),
	}

	resp, err := tm.client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list topics: %w", err)
	}

	var result struct {
		Data []api.Topic `json:"data"`
	}
	if err := resp.DecodeJSON(&result); err != nil {
		return nil, fmt.Errorf("failed to parse topic list response: %w", err)
	}

	return result.Data, nil
}

// GetTopic retrieves information about a specific topic.
// Returns errors:
//   - *api.Error with IsNotFound() if topic does not exist
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsForbidden() if user lacks permissions
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (tm *TopicManager) GetTopic(ctx context.Context, clusterID string, topicName string) (*api.Topic, error) {
	req := client.Request{
		Method: "GET",
		Path:   fmt.Sprintf("/kafka/v3/clusters/%s/topics/%s", clusterID, topicName),
	}

	resp, err := tm.client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to describe topic %s: %w", topicName, err)
	}

	var topic api.Topic
	if err := resp.DecodeJSON(&topic); err != nil {
		return nil, fmt.Errorf("failed to parse topic description: %w", err)
	}

	return &topic, nil
}

// CreateTopic creates a new topic.
// Returns errors:
//   - *api.Error with IsBadRequest() if parameters are invalid
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsConflict() if topic name already exists
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (tm *TopicManager) CreateTopic(ctx context.Context, clusterID string, topic api.Topic) error {
	body := map[string]interface{}{
		"topic_name":         topic.Name,
		"partitions_count":   topic.PartitionCount,
		"replication_factor": topic.ReplicationFactor,
		"configs":            topicConfigsToArray(topic.Config),
	}

	req := client.Request{
		Method: "POST",
		Path:   fmt.Sprintf("/kafka/v3/clusters/%s/topics", clusterID),
		Body:   body,
	}

	_, err := tm.client.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create topic %s: %w", topic.Name, err)
	}

	return nil
}

// DeleteTopic deletes a topic.
// Returns errors:
//   - *api.Error with IsNotFound() if topic does not exist
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsForbidden() if user lacks permissions
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (tm *TopicManager) DeleteTopic(ctx context.Context, clusterID string, topicName string) error {
	req := client.Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/kafka/v3/clusters/%s/topics/%s", clusterID, topicName),
	}

	_, err := tm.client.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete topic %s: %w", topicName, err)
	}
	return nil
}

// UpdateTopicConfig updates topic configuration.
// Returns errors:
//   - *api.Error with IsBadRequest() if config values are invalid
//   - *api.Error with IsNotFound() if topic does not exist
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (tm *TopicManager) UpdateTopicConfig(ctx context.Context, clusterID string, topicName string, configs map[string]string) error {
	configArray := topicConfigsToArray(configs)
	body := map[string]interface{}{
		"configs": configArray,
	}

	req := client.Request{
		Method: "PATCH",
		Path:   fmt.Sprintf("/kafka/v3/clusters/%s/topics/%s", clusterID, topicName),
		Body:   body,
	}

	_, err := tm.client.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update topic config %s: %w", topicName, err)
	}

	return nil
}

// GetTopicConfig retrieves topic configurations.
// Returns errors:
//   - *api.Error with IsNotFound() if topic does not exist
//   - *api.Error with IsUnauthorized() for authentication failures
//   - *api.Error with IsRateLimited() if rate limit is exceeded
func (tm *TopicManager) GetTopicConfig(ctx context.Context, clusterID string, topicName string) ([]api.TopicConfig, error) {
	req := client.Request{
		Method: "GET",
		Path:   fmt.Sprintf("/kafka/v3/clusters/%s/topics/%s/configs", clusterID, topicName),
	}

	resp, err := tm.client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get topic config: %w", err)
	}

	var result struct {
		Data []api.TopicConfig `json:"data"`
	}
	if err := resp.DecodeJSON(&result); err != nil {
		return nil, fmt.Errorf("failed to parse topic config response: %w", err)
	}

	return result.Data, nil
}

// Helper function to convert map to array format for API
func topicConfigsToArray(configs map[string]string) []map[string]string {
	configArray := make([]map[string]string, 0, len(configs))
	for key, value := range configs {
		configArray = append(configArray, map[string]string{
			"name":  key,
			"value": value,
		})
	}
	return configArray
}
