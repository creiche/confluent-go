// Package api defines the data types and interfaces for Confluent resources.
package api

// Cluster represents a Confluent Kafka cluster with its configuration and status.
// Clusters can be BASIC, STANDARD, or DEDICATED types.
type Cluster struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	ProviderType     string `json:"provider_type"`
	ProviderRegion   string `json:"provider_region"`
	ProviderCloud    string `json:"provider_cloud"`
	Status           string `json:"status"`
	BootstrapServers string `json:"bootstrap_servers"`
	Type             string `json:"type"` // BASIC, STANDARD, DEDICATED
}

// Topic represents a Kafka topic with its partition and replication configuration.
type Topic struct {
	Name              string            `json:"name"`
	PartitionCount    int32             `json:"partition_count"`
	ReplicationFactor int16             `json:"replication_factor"`
	Config            map[string]string `json:"config"`
}

// TopicConfig represents a single topic-level configuration key-value pair.
// Examples include retention.ms, cleanup.policy, compression.type, etc.
type TopicConfig struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ServiceAccount represents a Confluent service account used for programmatic access.
// Service accounts can own API keys and be granted permissions via ACLs.
type ServiceAccount struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Resource    Resource `json:"resource"`
}

// APIKey represents an API key for authentication to Confluent Cloud and Platform.
// The Secret field is only populated during creation and cannot be retrieved later.
type APIKey struct {
	ID          string  `json:"id"`
	Secret      string  `json:"secret"`
	Description string  `json:"description"`
	OwnerID     string  `json:"owner_id"`
	CreatedAt   string  `json:"created_at"`
	ExpiresAt   *string `json:"expires_at"`
}

// ACLBinding represents an access control list entry that grants or denies permissions.
// ACLs control access to Kafka resources like topics, consumer groups, and clusters.
type ACLBinding struct {
	Principal    string `json:"principal"` // "User:12345" or "User:*"
	ResourceType string `json:"resource_type"`
	ResourceName string `json:"resource_name"`
	PatternType  string `json:"pattern_type"` // LITERAL, PREFIXED
	Operation    string `json:"operation"`
	Permission   string `json:"permission"` // ALLOW, DENY
}

// Environment represents a Confluent environment, which is a logical grouping for clusters and resources.
// Environments provide isolation and organization for multi-tenant deployments.
type Environment struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

// RoleBinding represents a role assignment to a principal (user or service account).
// Role bindings grant permissions at the organization, environment, or cluster level.
type RoleBinding struct {
	ID          string `json:"id"`
	PrincipalID string `json:"principal_id"`
	RoleID      string `json:"role_id"`
	CRN         string `json:"crn"` // Confluent Resource Name
}

// Role represents a Confluent role that defines a set of permissions.
// Common roles include OrganizationAdmin, EnvironmentAdmin, CloudClusterAdmin, etc.
type Role struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Resource represents a reference to a Confluent resource.
// Used to link entities like service accounts, users, and API keys.
type Resource struct {
	ID   string `json:"id"`
	Type string `json:"type"` // USER, SERVICE_ACCOUNT, etc.
}

// BrokerConfig represents a broker-level configuration setting.
// Broker configs apply to individual Kafka brokers within a cluster.
type BrokerConfig struct {
	BrokerID string `json:"broker_id"`
	Name     string `json:"name"`
	Value    string `json:"value"`
}

// PartitionInfo represents metadata about a Kafka topic partition.
// Includes leader, replicas, and in-sync replica information.
type PartitionInfo struct {
	Topic     string  `json:"topic"`
	Partition int32   `json:"partition"`
	Leader    int32   `json:"leader"`
	Replicas  []int32 `json:"replicas"`
	ISR       []int32 `json:"isr"` // In-Sync Replicas
}

// SchemaSubject represents a subject in Confluent Schema Registry.
// A subject typically corresponds to a topic and contains multiple schema versions.
type SchemaSubject struct {
	Name     string  `json:"name"`
	Versions []int32 `json:"versions"`
	Latest   *Schema `json:"latest"`
}

// Schema represents a versioned schema in Confluent Schema Registry.
// Schemas can be AVRO, JSON_SCHEMA, or PROTOBUF format.
type Schema struct {
	ID         int32             `json:"id"`
	Subject    string            `json:"subject"`
	Version    int32             `json:"version"`
	Schema     string            `json:"schema"`
	Type       string            `json:"type"` // AVRO, JSON_SCHEMA, PROTOBUF
	References []SchemaReference `json:"references"`
}

// SchemaReference represents a reference from one schema to another.
// Used to model schema dependencies and composition.
type SchemaReference struct {
	Name    string `json:"name"`
	Subject string `json:"subject"`
	Version int32  `json:"version"`
}

// ConnectorConfig represents a Kafka Connect connector configuration.
// Connectors can be SOURCE (producing to Kafka) or SINK (consuming from Kafka).
type ConnectorConfig struct {
	Name   string            `json:"name"`
	Config map[string]string `json:"config"`
	Type   string            `json:"type"` // SOURCE or SINK
	Tasks  int32             `json:"tasks"`
	Status ConnectorStatus   `json:"status"`
	Topics []string          `json:"topics"`
}

// ConnectorStatus represents the current state of a Kafka Connect connector.
// Includes the overall state and status of individual tasks.
type ConnectorStatus struct {
	State  string           `json:"state"`
	Tasks  []TaskStatus     `json:"tasks"`
	Errors []ConnectorError `json:"errors"`
}

// TaskStatus represents the status of an individual connector task.
// Each connector can have multiple tasks running in parallel.
type TaskStatus struct {
	ID     int32  `json:"id"`
	State  string `json:"state"`
	Worker string `json:"worker"`
	Error  string `json:"error"`
}

// ConnectorError represents an error that occurred in a connector or task.
type ConnectorError struct {
	Message string `json:"message"`
}

// ConnectorPlugin represents a Kafka Connect connector plugin available in the cluster.
type ConnectorPlugin struct {
	Class   string `json:"class"`
	Type    string `json:"type"` // SOURCE or SINK
	Version string `json:"version"`
}

// ConnectorValidation represents the validation result for a connector configuration.
type ConnectorValidation struct {
	Name       string                      `json:"name"`
	ErrorCount int32                       `json:"error_count"`
	Groups     []string                    `json:"groups"`
	Configs    []ConnectorConfigValidation `json:"configs"`
}

// ConnectorConfigValidation represents validation information for a single config property.
type ConnectorConfigValidation struct {
	Definition ConfigDefinition `json:"definition"`
	Value      ConfigValue      `json:"value"`
}

// ConfigDefinition describes a connector configuration property.
type ConfigDefinition struct {
	Name          string   `json:"name"`
	Type          string   `json:"type"`
	Required      bool     `json:"required"`
	DefaultValue  string   `json:"default_value"`
	Importance    string   `json:"importance"` // HIGH, MEDIUM, LOW
	Documentation string   `json:"documentation"`
	Group         string   `json:"group"`
	Width         string   `json:"width"`
	DisplayName   string   `json:"display_name"`
	Dependents    []string `json:"dependents"`
	OrderInGroup  int32    `json:"order"`
}

// ConfigValue represents a configuration value and its validation status.
type ConfigValue struct {
	Name              string   `json:"name"`
	Value             string   `json:"value"`
	RecommendedValues []string `json:"recommended_values"`
	Errors            []string `json:"errors"`
	Visible           bool     `json:"visible"`
}

// ConnectorTask represents a task instance for a connector.
type ConnectorTask struct {
	ID        int32             `json:"id"`
	Config    map[string]string `json:"config"`
	Connector string            `json:"connector"`
}
