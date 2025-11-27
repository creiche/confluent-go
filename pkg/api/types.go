// Package api defines the data types and interfaces for Confluent resources.
package api

// Cluster represents a Confluent Kafka cluster.
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

// Topic represents a Kafka topic.
type Topic struct {
	Name              string            `json:"name"`
	PartitionCount    int32             `json:"partition_count"`
	ReplicationFactor int16             `json:"replication_factor"`
	Config            map[string]string `json:"config"`
}

// TopicConfig represents topic-level configuration.
type TopicConfig struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ServiceAccount represents a Confluent service account.
type ServiceAccount struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Resource    Resource `json:"resource"`
}

// APIKey represents an API key for authentication.
type APIKey struct {
	ID          string  `json:"id"`
	Secret      string  `json:"secret"`
	Description string  `json:"description"`
	OwnerID     string  `json:"owner_id"`
	CreatedAt   string  `json:"created_at"`
	ExpiresAt   *string `json:"expires_at"`
}

// ACLBinding represents an access control list binding.
type ACLBinding struct {
	Principal    string `json:"principal"` // "User:12345" or "User:*"
	ResourceType string `json:"resource_type"`
	ResourceName string `json:"resource_name"`
	PatternType  string `json:"pattern_type"` // LITERAL, PREFIXED
	Operation    string `json:"operation"`
	Permission   string `json:"permission"` // ALLOW, DENY
}

// Environment represents a Confluent environment.
type Environment struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

// RoleBinding represents a role assignment to a principal.
type RoleBinding struct {
	ID          string `json:"id"`
	PrincipalID string `json:"principal_id"`
	RoleID      string `json:"role_id"`
	CRN         string `json:"crn"` // Confluent Resource Name
}

// Role represents a Confluent role (e.g., OrganizationAdmin, EnvironmentAdmin).
type Role struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Resource represents a Confluent resource reference.
type Resource struct {
	ID   string `json:"id"`
	Type string `json:"type"` // USER, SERVICE_ACCOUNT, etc.
}

// BrokerConfig represents broker-level configuration.
type BrokerConfig struct {
	BrokerID string `json:"broker_id"`
	Name     string `json:"name"`
	Value    string `json:"value"`
}

// PartitionInfo represents information about a partition.
type PartitionInfo struct {
	Topic     string  `json:"topic"`
	Partition int32   `json:"partition"`
	Leader    int32   `json:"leader"`
	Replicas  []int32 `json:"replicas"`
	ISR       []int32 `json:"isr"` // In-Sync Replicas
}

// SchemaSubject represents a schema subject (for Schema Registry).
type SchemaSubject struct {
	Name     string  `json:"name"`
	Versions []int32 `json:"versions"`
	Latest   *Schema `json:"latest"`
}

// Schema represents a schema registered in Schema Registry.
type Schema struct {
	ID         int32             `json:"id"`
	Subject    string            `json:"subject"`
	Version    int32             `json:"version"`
	Schema     string            `json:"schema"`
	Type       string            `json:"type"` // AVRO, JSON_SCHEMA, PROTOBUF
	References []SchemaReference `json:"references"`
}

// SchemaReference represents a reference to another schema.
type SchemaReference struct {
	Name    string `json:"name"`
	Subject string `json:"subject"`
	Version int32  `json:"version"`
}

// ConnectorConfig represents a Kafka Connect connector.
type ConnectorConfig struct {
	Name   string            `json:"name"`
	Config map[string]string `json:"config"`
	Type   string            `json:"type"` // SOURCE or SINK
	Tasks  int32             `json:"tasks"`
	Status ConnectorStatus   `json:"status"`
	Topics []string          `json:"topics"`
}

// ConnectorStatus represents the status of a connector.
type ConnectorStatus struct {
	State  string           `json:"state"`
	Tasks  []TaskStatus     `json:"tasks"`
	Errors []ConnectorError `json:"errors"`
}

// TaskStatus represents the status of a connector task.
type TaskStatus struct {
	ID     int32  `json:"id"`
	State  string `json:"state"`
	Worker string `json:"worker"`
	Error  string `json:"error"`
}

// ConnectorError represents an error for a connector.
type ConnectorError struct {
	Message string `json:"message"`
}
