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

// Example 5: Manage Connectors
func exampleManageConnectors(c *client.Client, environmentID string, connectClusterID string) error {
	ctx := context.Background()

	connectorMgr := resources.NewConnectorManager(c)

	// List connectors
	fmt.Printf("=== Connectors in Connect Cluster %s ===\n", connectClusterID)
	connectors, err := connectorMgr.ListConnectors(ctx, environmentID, connectClusterID)
	if err != nil {
		fmt.Printf("Note: Failed to list connectors: %v\n", err)
		return nil
	}
	fmt.Printf("Found %d connectors\n", len(connectors))
	for i, connector := range connectors {
		fmt.Printf("  %d. %s\n", i+1, connector)
	}

	// List available connector plugins
	fmt.Println("\n=== Available Connector Plugins ===")
	plugins, err := connectorMgr.ListConnectorPlugins(ctx, environmentID, connectClusterID)
	if err == nil {
		for _, plugin := range plugins {
			fmt.Printf("  - %s (%s) v%s\n", plugin.Class, plugin.Type, plugin.Version)
		}
	}

	// Create a new connector
	fmt.Println("\n=== Creating Connector ===")
	connectorConfig := map[string]string{
		"connector.class":             "io.confluent.connect.jdbc.JdbcSourceConnector",
		"tasks.max":                   "1",
		"connection.url":              "jdbc:postgresql://localhost:5432/mydb",
		"connection.user":             "postgres",
		"connection.password":         "password",
		"mode":                        "incrementing",
		"incrementing.column.name":    "id",
		"topic.prefix":                "jdbc-",
		"poll.interval.ms":            "1000",
		"batch.max.rows":              "100",
		"table.whitelist":             "users,orders",
		"validate.non.null":           "false",
		"errors.tolerance":            "none",
		"errors.log.enable":           "true",
		"errors.log.include.messages": "true",
	}

	connector, err := connectorMgr.CreateConnector(ctx, environmentID, connectClusterID, "jdbc-source-example", connectorConfig)
	if err != nil {
		fmt.Printf("Note: Connector creation might have failed (connector may already exist): %v\n", err)
	} else {
		fmt.Printf("Created connector: %s\n", connector.Name)
	}

	// Get connector details
	fmt.Println("\n=== Connector Details ===")
	connector, err = connectorMgr.GetConnector(ctx, environmentID, connectClusterID, "jdbc-source-example")
	if err == nil && connector != nil {
		fmt.Printf("Connector: %s\n", connector.Name)
		fmt.Printf("Type: %s\n", connector.Type)
		fmt.Printf("Tasks: %d\n", connector.Tasks)
	}

	// Get connector status
	fmt.Println("\n=== Connector Status ===")
	status, err := connectorMgr.GetConnectorStatus(ctx, environmentID, connectClusterID, "jdbc-source-example")
	if err == nil && status != nil {
		fmt.Printf("State: %s\n", status.State)
		fmt.Printf("Tasks:\n")
		for _, task := range status.Tasks {
			fmt.Printf("  Task %d: %s (Worker: %s)\n", task.ID, task.State, task.Worker)
			if task.Error != "" {
				fmt.Printf("    Error: %s\n", task.Error)
			}
		}
	}

	// Pause connector
	fmt.Println("\n=== Pausing Connector ===")
	if err := connectorMgr.PauseConnector(ctx, environmentID, connectClusterID, "jdbc-source-example"); err != nil {
		fmt.Printf("Note: Failed to pause connector: %v\n", err)
	} else {
		fmt.Println("Connector paused")
	}

	// Resume connector
	fmt.Println("\n=== Resuming Connector ===")
	if err := connectorMgr.ResumeConnector(ctx, environmentID, connectClusterID, "jdbc-source-example"); err != nil {
		fmt.Printf("Note: Failed to resume connector: %v\n", err)
	} else {
		fmt.Println("Connector resumed")
	}

	// Update connector configuration
	fmt.Println("\n=== Updating Connector Config ===")
	updatedConfig := map[string]string{
		"tasks.max":        "2",
		"poll.interval.ms": "2000",
	}
	for k, v := range connectorConfig {
		if _, exists := updatedConfig[k]; !exists {
			updatedConfig[k] = v
		}
	}
	_, err = connectorMgr.UpdateConnector(ctx, environmentID, connectClusterID, "jdbc-source-example", updatedConfig)
	if err != nil {
		fmt.Printf("Note: Failed to update connector: %v\n", err)
	} else {
		fmt.Println("Connector configuration updated")
	}

	// Validate a connector configuration
	fmt.Println("\n=== Validating Connector Config ===")
	validation, err := connectorMgr.ValidateConnectorConfig(ctx, environmentID, connectClusterID,
		"io.confluent.connect.s3.S3SinkConnector",
		map[string]string{
			"topics":            "my-topic",
			"s3.bucket.name":    "my-bucket",
			"s3.region":         "us-west-2",
			"flush.size":        "1000",
			"storage.class":     "STANDARD",
			"format.class":      "io.confluent.connect.s3.format.json.JsonFormat",
			"partitioner.class": "io.confluent.connect.storage.partitioner.DefaultPartitioner",
		})
	if err == nil && validation != nil {
		fmt.Printf("Validation result: %d errors\n", validation.ErrorCount)
		if validation.ErrorCount > 0 {
			for _, cfg := range validation.Configs {
				if len(cfg.Value.Errors) > 0 {
					fmt.Printf("  Config '%s' errors:\n", cfg.Value.Name)
					for _, errMsg := range cfg.Value.Errors {
						fmt.Printf("    - %s\n", errMsg)
					}
				}
			}
		}
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
	connectClusterID := "lcc-xyz789"

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
		{"Manage Connectors", func(c *client.Client) error {
			return exampleManageConnectors(c, environmentID, connectClusterID)
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
