package main

import (
	"context"
	"fmt"
	"log"

	"github.com/creiche/confluent-go/pkg/api"
	"github.com/creiche/confluent-go/pkg/client"
	"github.com/creiche/confluent-go/pkg/resources"
)

// Example 1: List Environments and Clusters
func exampleListResources(c *client.Client, environmentID string) error {
	ctx := context.Background()

	// List environments
	envMgr := resources.NewEnvironmentManager(c)
	envs, err := envMgr.ListEnvironments(ctx)
	if err != nil {
		return fmt.Errorf("failed to list environments: %w", err)
	}

	fmt.Println("=== Environments ===")
	for _, env := range envs {
		fmt.Printf("Name: %s, ID: %s\n", env.Name, env.ID)
	}

	// List clusters in specified environment
	clusterMgr := resources.NewClusterManager(c)
	clusters, err := clusterMgr.ListClusters(ctx, environmentID)
	if err != nil {
		return fmt.Errorf("failed to list clusters: %w", err)
	}

	fmt.Println("\n=== Clusters ===")
	for _, cluster := range clusters {
		fmt.Printf("Name: %s, ID: %s, Status: %s\n", cluster.Name, cluster.ID, cluster.Status)
	}

	return nil
}

// Example 2: Manage Service Accounts and API Keys
func exampleManageServiceAccounts(c *client.Client) error {
	ctx := context.Background()
	saMgr := resources.NewServiceAccountManager(c)

	// Create a service account
	fmt.Println("=== Creating Service Account ===")
	sa, err := saMgr.CreateServiceAccount(ctx, "example-sa", "Example Service Account for operators")
	if err != nil {
		return fmt.Errorf("failed to create service account: %w", err)
	}
	fmt.Printf("Created Service Account: %s (%s)\n", sa.Name, sa.ID)

	// Create API key for the service account
	fmt.Println("\n=== Creating API Key ===")
	apiKey, err := saMgr.CreateAPIKey(ctx, sa.ID, "Key for Kubernetes operator")
	if err != nil {
		return fmt.Errorf("failed to create API key: %w", err)
	}
	fmt.Printf("Created API Key: %s\n", apiKey.ID)
	fmt.Printf("Secret: %s (save this securely!)\n", apiKey.Secret)

	// List API keys
	fmt.Println("\n=== Listing API Keys ===")
	keys, err := saMgr.ListAPIKeys(ctx, sa.ID)
	if err != nil {
		return fmt.Errorf("failed to list API keys: %w", err)
	}
	fmt.Printf("Found %d API keys\n", len(keys))

	// Cleanup: Delete API key
	fmt.Println("\n=== Cleaning up ===")
	if err := saMgr.DeleteAPIKey(ctx, apiKey.ID); err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}
	fmt.Printf("Deleted API Key: %s\n", apiKey.ID)

	// Delete service account
	if err := saMgr.DeleteServiceAccount(ctx, sa.ID); err != nil {
		return fmt.Errorf("failed to delete service account: %w", err)
	}
	fmt.Printf("Deleted Service Account: %s\n", sa.ID)

	return nil
}

// Example 3: Manage Topics
func exampleManageTopics(c *client.Client, clusterID string) error {
	ctx := context.Background()

	topicMgr := resources.NewTopicManager(c)

	// List existing topics
	fmt.Printf("=== Topics in Cluster %s ===\n", clusterID)
	topics, err := topicMgr.ListTopics(ctx, clusterID)
	if err != nil {
		return fmt.Errorf("failed to list topics: %w", err)
	}
	fmt.Printf("Found %d topics\n", len(topics))

	// Create a new topic
	fmt.Println("\n=== Creating Topic ===")
	newTopic := api.Topic{
		Name:              "example-topic",
		PartitionCount:    3,
		ReplicationFactor: 1,
		Config: map[string]string{
			"retention.ms":     "86400000", // 1 day
			"compression.type": "snappy",
		},
	}

	if err := topicMgr.CreateTopic(ctx, clusterID, newTopic); err != nil {
		fmt.Printf("Note: Topic creation might have failed (topic may already exist): %v\n", err)
	} else {
		fmt.Printf("Created topic: %s\n", newTopic.Name)
	}

	// Get topic details
	fmt.Println("\n=== Topic Details ===")
	topic, err := topicMgr.GetTopic(ctx, clusterID, "example-topic")
	if err == nil && topic != nil {
		fmt.Printf("Topic: %s, Partitions: %d, Replication Factor: %d\n",
			topic.Name, topic.PartitionCount, topic.ReplicationFactor)
	}

	// Update topic configuration
	fmt.Println("\n=== Updating Topic Config ===")
	if err := topicMgr.UpdateTopicConfig(ctx, clusterID, "example-topic", map[string]string{
		"retention.ms": "172800000", // 2 days
	}); err != nil {
		fmt.Printf("Note: Topic update might have failed: %v\n", err)
	} else {
		fmt.Printf("Updated topic configuration\n")
	}

	// Get topic configuration
	fmt.Println("\n=== Topic Configuration ===")
	configs, err := topicMgr.GetTopicConfig(ctx, clusterID, "example-topic")
	if err == nil {
		for _, cfg := range configs {
			fmt.Printf("  %s: %s\n", cfg.Name, cfg.Value)
		}
	}

	return nil
}

// Example 4: Manage ACLs
func exampleManageACLs(c *client.Client, clusterID string) error {
	ctx := context.Background()

	aclMgr := resources.NewACLManager(c)

	// List ACLs
	fmt.Printf("=== ACLs in Cluster %s ===\n", clusterID)
	acls, err := aclMgr.ListACLs(ctx, clusterID)
	if err != nil {
		fmt.Printf("Note: Failed to list ACLs (may require specific permissions): %v\n", err)
		return nil
	}
	fmt.Printf("Found %d ACLs\n", len(acls))
	for i, acl := range acls {
		fmt.Printf("  %d. Principal: %s, Operation: %s, Resource: %s/%s\n",
			i+1, acl.Principal, acl.Operation, acl.ResourceType, acl.ResourceName)
	}

	return nil
}

func main() {
	// Initialize the Confluent REST client
	cfg := client.Config{
		BaseURL:   "https://api.confluent.cloud",
		APIKey:    "YOUR_API_KEY",
		APISecret: "YOUR_API_SECRET",
	}

	c, err := client.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create Confluent client: %v", err)
	}

	// These should be set to your actual environment and cluster IDs
	environmentID := "env-abc123"
	clusterID := "lkc-xyz789"

	// Run examples
	examples := []struct {
		name string
		fn   func(*client.Client) error
	}{
		{"List Resources", func(c *client.Client) error {
			return exampleListResources(c, environmentID)
		}},
		{"Manage Service Accounts", func(c *client.Client) error {
			return exampleManageServiceAccounts(c)
		}},
		{"Manage Topics", func(c *client.Client) error {
			return exampleManageTopics(c, clusterID)
		}},
		{"Manage ACLs", func(c *client.Client) error {
			return exampleManageACLs(c, clusterID)
		}},
	}

	for _, ex := range examples {
		fmt.Printf("\n########################################\n")
		fmt.Printf("## %s\n", ex.name)
		fmt.Printf("########################################\n\n")

		if err := ex.fn(c); err != nil {
			log.Printf("Error in %s: %v", ex.name, err)
		}
	}
}
