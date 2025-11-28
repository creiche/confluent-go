package resources_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/creiche/confluent-go/pkg/api"
	"github.com/creiche/confluent-go/pkg/client"
	"github.com/creiche/confluent-go/pkg/resources"
)

func newTestClient(t *testing.T, baseURL string) *client.Client {
	cfg := client.Config{
		BaseURL:   baseURL,
		APIKey:    "test-key",
		APISecret: "test-secret",
	}
	c, err := client.NewClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	return c
}

// Cluster Manager Tests

func TestClusterManager_ListClusters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cmk/v2/clusters" {
			t.Errorf("Expected path /cmk/v2/clusters, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "lkc-1", "name": "cluster-1"},
				{"id": "lkc-2", "name": "cluster-2"},
			},
		}); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := newTestClient(t, server.URL)
	mgr := resources.NewClusterManager(c)

	clusters, err := mgr.ListClusters(context.Background(), "env-123")
	if err != nil {
		t.Fatalf("ListClusters failed: %v", err)
	}

	if len(clusters) != 2 {
		t.Errorf("Expected 2 clusters, got %d", len(clusters))
	}
}

func TestClusterManager_GetCluster(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "lkc-123",
			"name": "my-cluster",
		}); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := newTestClient(t, server.URL)
	mgr := resources.NewClusterManager(c)

	cluster, err := mgr.GetCluster(context.Background(), "lkc-123")
	if err != nil {
		t.Fatalf("GetCluster failed: %v", err)
	}

	if cluster.ID != "lkc-123" {
		t.Errorf("Expected ID lkc-123, got %s", cluster.ID)
	}
}

func TestClusterManager_DeleteCluster(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := newTestClient(t, server.URL)
	mgr := resources.NewClusterManager(c)

	err := mgr.DeleteCluster(context.Background(), "lkc-123")
	if err != nil {
		t.Fatalf("DeleteCluster failed: %v", err)
	}
}

// Topic Manager Tests

func TestTopicManager_ListTopics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"name": "topic-1"},
				{"name": "topic-2"},
			},
		}); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := newTestClient(t, server.URL)
	mgr := resources.NewTopicManager(c)

	topics, err := mgr.ListTopics(context.Background(), "lkc-123")
	if err != nil {
		t.Fatalf("ListTopics failed: %v", err)
	}

	if len(topics) != 2 {
		t.Errorf("Expected 2 topics, got %d", len(topics))
	}
}

func TestTopicManager_GetTopic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"name":             "my-topic",
			"partitions_count": 3,
		}); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := newTestClient(t, server.URL)
	mgr := resources.NewTopicManager(c)

	topic, err := mgr.GetTopic(context.Background(), "lkc-123", "my-topic")
	if err != nil {
		t.Fatalf("GetTopic failed: %v", err)
	}

	if topic.Name != "my-topic" {
		t.Errorf("Expected name my-topic, got %s", topic.Name)
	}
}

func TestTopicManager_DeleteTopic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := newTestClient(t, server.URL)
	mgr := resources.NewTopicManager(c)

	err := mgr.DeleteTopic(context.Background(), "lkc-123", "my-topic")
	if err != nil {
		t.Fatalf("DeleteTopic failed: %v", err)
	}
}

// Service Account Manager Tests

func TestServiceAccountManager_ListServiceAccounts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "sa-1", "display_name": "service-account-1"},
				{"id": "sa-2", "display_name": "service-account-2"},
			},
		}); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := newTestClient(t, server.URL)
	mgr := resources.NewServiceAccountManager(c)

	sas, err := mgr.ListServiceAccounts(context.Background())
	if err != nil {
		t.Fatalf("ListServiceAccounts failed: %v", err)
	}

	if len(sas) != 2 {
		t.Errorf("Expected 2 service accounts, got %d", len(sas))
	}
}

func TestServiceAccountManager_CreateServiceAccount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"id":           "sa-new",
			"display_name": "new-sa",
		}); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := newTestClient(t, server.URL)
	mgr := resources.NewServiceAccountManager(c)

	sa, err := mgr.CreateServiceAccount(context.Background(), "new-sa", "New SA")
	if err != nil {
		t.Fatalf("CreateServiceAccount failed: %v", err)
	}

	if sa.ID != "sa-new" {
		t.Errorf("Expected ID sa-new, got %s", sa.ID)
	}
}

func TestServiceAccountManager_DeleteServiceAccount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := newTestClient(t, server.URL)
	mgr := resources.NewServiceAccountManager(c)

	err := mgr.DeleteServiceAccount(context.Background(), "sa-123")
	if err != nil {
		t.Fatalf("DeleteServiceAccount failed: %v", err)
	}
}

// ACL Manager Tests

func TestACLManager_ListACLs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"principal": "User:sa-1", "operation": "Read"},
			},
		}); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := newTestClient(t, server.URL)
	mgr := resources.NewACLManager(c)

	acls, err := mgr.ListACLs(context.Background(), "lkc-123")
	if err != nil {
		t.Fatalf("ListACLs failed: %v", err)
	}

	if len(acls) != 1 {
		t.Errorf("Expected 1 ACL, got %d", len(acls))
	}
}

func TestACLManager_CreateACL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := newTestClient(t, server.URL)
	mgr := resources.NewACLManager(c)

	acl := api.ACLBinding{
		Principal:    "User:sa-1",
		Operation:    "Read",
		ResourceType: "Topic",
		ResourceName: "my-topic",
		PatternType:  "LITERAL",
		Permission:   "ALLOW",
	}

	err := mgr.CreateACL(context.Background(), "lkc-123", acl)
	if err != nil {
		t.Fatalf("CreateACL failed: %v", err)
	}
}

func TestACLManager_DeleteACL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := newTestClient(t, server.URL)
	mgr := resources.NewACLManager(c)

	err := mgr.DeleteACL(context.Background(), "lkc-123", "User:sa-1", "Read", "Topic", "my-topic")
	if err != nil {
		t.Fatalf("DeleteACL failed: %v", err)
	}
}

// Environment Manager Tests

func TestEnvironmentManager_ListEnvironments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "env-1", "name": "environment-1"},
				{"id": "env-2", "name": "environment-2"},
			},
		}); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := newTestClient(t, server.URL)
	mgr := resources.NewEnvironmentManager(c)

	envs, err := mgr.ListEnvironments(context.Background())
	if err != nil {
		t.Fatalf("ListEnvironments failed: %v", err)
	}

	if len(envs) != 2 {
		t.Errorf("Expected 2 environments, got %d", len(envs))
	}
}

func TestEnvironmentManager_GetEnvironment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "env-123",
			"name": "my-environment",
		}); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := newTestClient(t, server.URL)
	mgr := resources.NewEnvironmentManager(c)

	env, err := mgr.GetEnvironment(context.Background(), "env-123")
	if err != nil {
		t.Fatalf("GetEnvironment failed: %v", err)
	}

	if env.ID != "env-123" {
		t.Errorf("Expected ID env-123, got %s", env.ID)
	}
}

func TestEnvironmentManager_CreateEnvironment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "env-new",
			"name": "new-environment",
		}); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := newTestClient(t, server.URL)
	mgr := resources.NewEnvironmentManager(c)

	env, err := mgr.CreateEnvironment(context.Background(), "new-environment", "New Environment")
	if err != nil {
		t.Fatalf("CreateEnvironment failed: %v", err)
	}

	if env.ID != "env-new" {
		t.Errorf("Expected ID env-new, got %s", env.ID)
	}
}

func TestEnvironmentManager_DeleteEnvironment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := newTestClient(t, server.URL)
	mgr := resources.NewEnvironmentManager(c)

	err := mgr.DeleteEnvironment(context.Background(), "env-123")
	if err != nil {
		t.Fatalf("DeleteEnvironment failed: %v", err)
	}
}

// Connector Manager Tests

func TestConnectorManager_ListConnectors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/connect/v1/environments/env-123/clusters/lcc-123/connectors" {
			t.Errorf("Expected path /connect/v1/environments/env-123/clusters/lcc-123/connectors, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode([]string{
			"connector-1",
			"connector-2",
		}); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := newTestClient(t, server.URL)
	mgr := resources.NewConnectorManager(c)

	connectors, err := mgr.ListConnectors(context.Background(), "env-123", "lcc-123")
	if err != nil {
		t.Fatalf("ListConnectors failed: %v", err)
	}

	if len(connectors) != 2 {
		t.Errorf("Expected 2 connectors, got %d", len(connectors))
	}

	if connectors[0] != "connector-1" {
		t.Errorf("Expected first connector to be connector-1, got %s", connectors[0])
	}
}

func TestConnectorManager_GetConnector(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"name":   "my-connector",
			"config": map[string]string{"connector.class": "io.confluent.connect.jdbc.JdbcSourceConnector"},
			"tasks":  2,
		}); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := newTestClient(t, server.URL)
	mgr := resources.NewConnectorManager(c)

	connector, err := mgr.GetConnector(context.Background(), "env-123", "lcc-123", "my-connector")
	if err != nil {
		t.Fatalf("GetConnector failed: %v", err)
	}

	if connector.Name != "my-connector" {
		t.Errorf("Expected name my-connector, got %s", connector.Name)
	}
}

func TestConnectorManager_CreateConnector(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"name": "new-connector",
			"config": map[string]string{
				"connector.class": "io.confluent.connect.jdbc.JdbcSourceConnector",
			},
		}); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := newTestClient(t, server.URL)
	mgr := resources.NewConnectorManager(c)

	config := map[string]string{
		"connector.class": "io.confluent.connect.jdbc.JdbcSourceConnector",
		"tasks.max":       "1",
	}

	connector, err := mgr.CreateConnector(context.Background(), "env-123", "lcc-123", "new-connector", config)
	if err != nil {
		t.Fatalf("CreateConnector failed: %v", err)
	}

	if connector.Name != "new-connector" {
		t.Errorf("Expected name new-connector, got %s", connector.Name)
	}
}

func TestConnectorManager_UpdateConnector(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"name": "my-connector",
			"config": map[string]string{
				"tasks.max": "2",
			},
		}); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := newTestClient(t, server.URL)
	mgr := resources.NewConnectorManager(c)

	config := map[string]string{
		"tasks.max": "2",
	}

	connector, err := mgr.UpdateConnector(context.Background(), "env-123", "lcc-123", "my-connector", config)
	if err != nil {
		t.Fatalf("UpdateConnector failed: %v", err)
	}

	if connector.Name != "my-connector" {
		t.Errorf("Expected name my-connector, got %s", connector.Name)
	}
}

func TestConnectorManager_DeleteConnector(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := newTestClient(t, server.URL)
	mgr := resources.NewConnectorManager(c)

	err := mgr.DeleteConnector(context.Background(), "env-123", "lcc-123", "my-connector")
	if err != nil {
		t.Fatalf("DeleteConnector failed: %v", err)
	}
}

func TestConnectorManager_GetConnectorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"state": "RUNNING",
			"tasks": []map[string]interface{}{
				{"id": 0, "state": "RUNNING"},
			},
		}); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := newTestClient(t, server.URL)
	mgr := resources.NewConnectorManager(c)

	status, err := mgr.GetConnectorStatus(context.Background(), "env-123", "lcc-123", "my-connector")
	if err != nil {
		t.Fatalf("GetConnectorStatus failed: %v", err)
	}

	if status.State != "RUNNING" {
		t.Errorf("Expected state RUNNING, got %s", status.State)
	}
}

func TestConnectorManager_PauseConnector(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	c := newTestClient(t, server.URL)
	mgr := resources.NewConnectorManager(c)

	err := mgr.PauseConnector(context.Background(), "env-123", "lcc-123", "my-connector")
	if err != nil {
		t.Fatalf("PauseConnector failed: %v", err)
	}
}

func TestConnectorManager_ResumeConnector(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	c := newTestClient(t, server.URL)
	mgr := resources.NewConnectorManager(c)

	err := mgr.ResumeConnector(context.Background(), "env-123", "lcc-123", "my-connector")
	if err != nil {
		t.Fatalf("ResumeConnector failed: %v", err)
	}
}

func TestConnectorManager_RestartConnector(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := newTestClient(t, server.URL)
	mgr := resources.NewConnectorManager(c)

	err := mgr.RestartConnector(context.Background(), "env-123", "lcc-123", "my-connector")
	if err != nil {
		t.Fatalf("RestartConnector failed: %v", err)
	}
}

func TestConnectorManager_ListConnectorPlugins(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode([]map[string]interface{}{
			{
				"class":   "io.confluent.connect.jdbc.JdbcSourceConnector",
				"type":    "SOURCE",
				"version": "10.7.4",
			},
			{
				"class":   "io.confluent.connect.s3.S3SinkConnector",
				"type":    "SINK",
				"version": "10.5.4",
			},
		}); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := newTestClient(t, server.URL)
	mgr := resources.NewConnectorManager(c)

	plugins, err := mgr.ListConnectorPlugins(context.Background(), "env-123", "lcc-123")
	if err != nil {
		t.Fatalf("ListConnectorPlugins failed: %v", err)
	}

	if len(plugins) != 2 {
		t.Errorf("Expected 2 plugins, got %d", len(plugins))
	}
}
