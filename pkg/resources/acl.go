package resources

import (
	"context"
	"fmt"

	"github.com/creiche/confluent-go/pkg/api"
	"github.com/creiche/confluent-go/pkg/client"
)

// ACLManager handles ACL-related operations via REST API.
type ACLManager struct {
	client *client.Client
}

// NewACLManager creates a new ACL manager.
func NewACLManager(c *client.Client) *ACLManager {
	return &ACLManager{client: c}
}

// ListACLs lists all ACLs in a cluster.
func (am *ACLManager) ListACLs(ctx context.Context, clusterID string) ([]api.ACLBinding, error) {
	req := client.Request{
		Method: "GET",
		Path:   fmt.Sprintf("/kafka/v3/clusters/%s/acls", clusterID),
	}

	resp, err := am.client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list ACLs: %w", err)
	}

	var result struct {
		Data []api.ACLBinding `json:"data"`
	}
	if err := resp.DecodeJSON(&result); err != nil {
		return nil, fmt.Errorf("failed to parse ACL list response: %w", err)
	}

	return result.Data, nil
}

// CreateACL creates a new ACL binding.
func (am *ACLManager) CreateACL(ctx context.Context, clusterID string, acl api.ACLBinding) error {
	body := map[string]interface{}{
		"resource_type": acl.ResourceType,
		"resource_name": acl.ResourceName,
		"pattern_type":  acl.PatternType,
		"principal":     acl.Principal,
		"operation":     acl.Operation,
		"permission":    acl.Permission,
	}

	req := client.Request{
		Method: "POST",
		Path:   fmt.Sprintf("/kafka/v3/clusters/%s/acls", clusterID),
		Body:   body,
	}

	_, err := am.client.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create ACL: %w", err)
	}

	return nil
}

// DeleteACL deletes an ACL binding.
func (am *ACLManager) DeleteACL(ctx context.Context, clusterID string, principal string, operation string, resourceType string, resourceName string) error {
	req := client.Request{
		Method: "DELETE",
		Path: fmt.Sprintf("/kafka/v3/clusters/%s/acls?principal=%s&operation=%s&resource_type=%s&resource_name=%s",
			clusterID, principal, operation, resourceType, resourceName),
	}

	_, err := am.client.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete ACL: %w", err)
	}

	return nil
}
