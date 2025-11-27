package main

import (
	"context"
	"fmt"
	"log"

	"github.com/creiche/confluent-go/pkg/api"
	"github.com/creiche/confluent-go/pkg/client"
	"github.com/creiche/confluent-go/pkg/resources"
)

// This file demonstrates a pattern for using confluent-go in a Kubernetes operator
// This would typically be in your operator's pkg/reconcilers/ directory

type OperatorConfig struct {
	BaseURL            string
	APIKey             string
	APISecret          string
	DefaultCluster     string
	DefaultEnvironment string
}

// OperatorReconciler represents a reconciler that manages Confluent resources via REST API
type OperatorReconciler struct {
	confluentClient *client.Client
	config          OperatorConfig
}

// NewOperatorReconciler creates a new reconciler for managing Confluent resources
func NewOperatorReconciler(config OperatorConfig) (*OperatorReconciler, error) {
	clientConfig := client.Config{
		BaseURL:   config.BaseURL,
		APIKey:    config.APIKey,
		APISecret: config.APISecret,
	}

	if clientConfig.BaseURL == "" {
		clientConfig.BaseURL = "https://api.confluent.cloud"
	}

	c, err := client.NewClient(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Confluent client: %w", err)
	}

	return &OperatorReconciler{
		confluentClient: c,
		config:          config,
	}, nil
}

// ReconcileTopic represents a simple example of reconciling a Kafka topic
// This would be called when a Topic custom resource is created/updated
func (r *OperatorReconciler) ReconcileTopic(ctx context.Context, topicName string, partitions int32, replicationFactor int16) error {
	topicMgr := resources.NewTopicManager(r.confluentClient)

	// Check if topic exists
	topic, err := topicMgr.GetTopic(ctx, r.config.DefaultCluster, topicName)
	if err != nil {
		// Topic doesn't exist, create it
		log.Printf("Topic %s not found, creating...\n", topicName)
		newTopic := createTopicFromSpec(topicName, partitions, replicationFactor)
		if err := topicMgr.CreateTopic(ctx, r.config.DefaultCluster, newTopic); err != nil {
			return fmt.Errorf("failed to create topic: %w", err)
		}
		return nil
	}

	// Topic exists, check if it needs updates
	if topic.PartitionCount != partitions {
		log.Printf("Topic %s partition count mismatch (expected %d, got %d)\n",
			topicName, partitions, topic.PartitionCount)
		// Note: Partition count updates require special handling in Confluent
	}

	return nil
}

// ReconcileServiceAccount ensures a service account and its API keys exist
func (r *OperatorReconciler) ReconcileServiceAccount(ctx context.Context, saName string) (*client.Config, error) {
	saMgr := resources.NewServiceAccountManager(r.confluentClient)

	// List existing service accounts to find ours
	accounts, err := saMgr.ListServiceAccounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list service accounts: %w", err)
	}

	var existingSA *api.ServiceAccount
	for i := range accounts {
		if accounts[i].Name == saName {
			// Found existing service account
			existingSA = &accounts[i]
			break
		}
	}

	var serviceAccount *api.ServiceAccount
	if existingSA == nil {
		// Create new service account
		log.Printf("Service account %s not found, creating...\n", saName)
		created, err := saMgr.CreateServiceAccount(ctx, saName, fmt.Sprintf("Service account for %s", saName))
		if err != nil {
			return nil, fmt.Errorf("failed to create service account: %w", err)
		}
		serviceAccount = created
	} else {
		serviceAccount = existingSA
	}

	// Ensure API key exists
	keys, err := saMgr.ListAPIKeys(ctx, serviceAccount.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}

	var apiKey *api.APIKey
	if len(keys) == 0 {
		// Create API key
		log.Printf("No API keys found for service account %s, creating...\n", saName)
		created, err := saMgr.CreateAPIKey(ctx, serviceAccount.ID, fmt.Sprintf("Key for %s", saName))
		if err != nil {
			return nil, fmt.Errorf("failed to create API key: %w", err)
		}
		apiKey = created
	} else {
		// Use first available key
		apiKey = &keys[0]
	}

	// Return configuration that could be used by Kubernetes secret
	return &client.Config{
		APIKey:    apiKey.ID,
		APISecret: apiKey.Secret,
	}, nil
}

// ReconcileACLs ensures proper access controls are in place
func (r *OperatorReconciler) ReconcileACLs(ctx context.Context, principal string, permissions map[string][]string) error {
	aclMgr := resources.NewACLManager(r.confluentClient)

	// permissions map: resource_type -> []operations
	// Example: {"Topic": ["Read", "Write"], "ConsumerGroup": ["Read"]}

	for resourceType, operations := range permissions {
		for _, operation := range operations {
			acl := api.ACLBinding{
				Principal:    principal,
				Operation:    operation,
				ResourceType: resourceType,
				ResourceName: "*", // Allow all resources of this type
				PatternType:  "PREFIXED",
				Permission:   "ALLOW",
			}

			if err := aclMgr.CreateACL(ctx, r.config.DefaultCluster, acl); err != nil {
				log.Printf("Failed to create ACL (may already exist): %v\n", err)
				// Continue, as the ACL might already exist
			}
		}
	}

	return nil
}

// Helper function to create a topic from spec
func createTopicFromSpec(name string, partitions int32, replicationFactor int16) api.Topic {
	return api.Topic{
		Name:              name,
		PartitionCount:    partitions,
		ReplicationFactor: replicationFactor,
		Config: map[string]string{
			"retention.ms": "604800000", // 7 days default
		},
	}
}

// Usage example in an operator's Setup function
func ExampleOperatorSetup() error {
	config := OperatorConfig{
		BaseURL:            "https://api.confluent.cloud",
		APIKey:             "your-api-key-id",
		APISecret:          "your-api-key-secret",
		DefaultCluster:     "lkc-xyz123", // Set to your cluster ID
		DefaultEnvironment: "env-abc123", // Set to your environment ID
	}

	reconciler, err := NewOperatorReconciler(config)
	if err != nil {
		return fmt.Errorf("failed to create reconciler: %w", err)
	}

	ctx := context.Background()

	// Example 1: Reconcile a topic
	if err := reconciler.ReconcileTopic(ctx, "my-app-topic", 3, 1); err != nil {
		log.Printf("Failed to reconcile topic: %v", err)
	}

	// Example 2: Reconcile service account and get credentials
	saConfig, err := reconciler.ReconcileServiceAccount(ctx, "my-app-sa")
	if err != nil {
		log.Printf("Failed to reconcile service account: %v", err)
	} else {
		log.Printf("Service account credentials: API Key=%s", saConfig.APIKey)
		// Store saConfig in a Kubernetes Secret
	}

	// Example 3: Reconcile ACLs
	permissions := map[string][]string{
		"Topic":         {"Read", "Write"},
		"ConsumerGroup": {"Read"},
	}
	if err := reconciler.ReconcileACLs(ctx, "User:sa-12345", permissions); err != nil {
		log.Printf("Failed to reconcile ACLs: %v", err)
	}

	return nil
}
