package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/models"
)

func TestClient_UpdateManifest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" || r.URL.Path != "/v2/clusters/cluster-id/manifests" {
			t.Errorf("Expected PUT /v2/clusters/cluster-id/manifests, got %s %s", r.Method, r.URL.Path)
		}

		var params models.UpdateManifestParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		if params.FileName != "updated.yaml" {
			t.Errorf("Expected filename 'updated.yaml', got %s", params.FileName)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		BaseURL: server.URL,
		OfflineToken: "test-token",
	})

	params := models.UpdateManifestParams{
		FileName: "updated.yaml",
		Content:  "updated-base64-content",
	}

	err := client.UpdateManifest(context.Background(), "cluster-id", params)
	if err != nil {
		t.Fatalf("UpdateManifest() error = %v", err)
	}
}

func TestClient_GetHost(t *testing.T) {
	expectedHost := &models.Host{
		ID:         "host-id",
		InfraEnvID: "infra-env-id",
		Status:     "known",
		StatusInfo: "Host is ready",
		RequestedHostname: "worker-1",
		Role:       "worker",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/v2/infra-envs/infra-env-id/hosts/host-id" {
			t.Errorf("Expected GET /v2/infra-envs/infra-env-id/hosts/host-id, got %s %s", r.Method, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedHost)
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		BaseURL: server.URL,
		OfflineToken: "test-token",
	})

	host, err := client.GetHost(context.Background(), "infra-env-id", "host-id")
	if err != nil {
		t.Fatalf("GetHost() error = %v", err)
	}

	if host.ID != expectedHost.ID {
		t.Errorf("GetHost() ID = %v, want %v", host.ID, expectedHost.ID)
	}

	if host.Status != expectedHost.Status {
		t.Errorf("GetHost() Status = %v, want %v", host.Status, expectedHost.Status)
	}

	if host.Role != expectedHost.Role {
		t.Errorf("GetHost() Role = %v, want %v", host.Role, expectedHost.Role)
	}
}

func TestClient_UnbindHost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/v2/infra-envs/infra-env-id/hosts/host-id/actions/unbind" {
			t.Errorf("Expected POST /v2/infra-envs/infra-env-id/hosts/host-id/actions/unbind, got %s %s", r.Method, r.URL.Path)
		}

		// Unbind doesn't have a request body
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		BaseURL: server.URL,
		OfflineToken: "test-token",
	})

	err := client.UnbindHost(context.Background(), "infra-env-id", "host-id")
	if err != nil {
		t.Fatalf("UnbindHost() error = %v", err)
	}
}

func TestClient_UpdateInfraEnv(t *testing.T) {
	expectedInfraEnv := &models.InfraEnv{
		ID:               "infra-env-id",
		Name:             "updated-infra-env",
		OpenshiftVersion: "4.15.20",
		Type:             "full-iso",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" || r.URL.Path != "/v2/infra-envs/infra-env-id" {
			t.Errorf("Expected PATCH /v2/infra-envs/infra-env-id, got %s %s", r.Method, r.URL.Path)
		}

		var params models.InfraEnvUpdateParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedInfraEnv)
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		BaseURL: server.URL,
		OfflineToken: "test-token",
	})

	name := "updated-infra-env"
	params := models.InfraEnvUpdateParams{
		Name: &name,
	}

	infraEnv, err := client.UpdateInfraEnv(context.Background(), "infra-env-id", params)
	if err != nil {
		t.Fatalf("UpdateInfraEnv() error = %v", err)
	}

	if infraEnv.ID != expectedInfraEnv.ID {
		t.Errorf("UpdateInfraEnv() ID = %v, want %v", infraEnv.ID, expectedInfraEnv.ID)
	}

	if infraEnv.Name != expectedInfraEnv.Name {
		t.Errorf("UpdateInfraEnv() Name = %v, want %v", infraEnv.Name, expectedInfraEnv.Name)
	}
}

func TestClient_ListInfraEnvs(t *testing.T) {
	expectedInfraEnvs := []models.InfraEnv{
		{
			ID:   "infra-env-1",
			Name: "infra-env-1",
		},
		{
			ID:   "infra-env-2",
			Name: "infra-env-2",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/v2/infra-envs" {
			t.Errorf("Expected GET /v2/infra-envs, got %s %s", r.Method, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedInfraEnvs)
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		BaseURL: server.URL,
		OfflineToken: "test-token",
	})

	infraEnvs, err := client.ListInfraEnvs(context.Background())
	if err != nil {
		t.Fatalf("ListInfraEnvs() error = %v", err)
	}

	if len(infraEnvs) != len(expectedInfraEnvs) {
		t.Errorf("ListInfraEnvs() returned %d infra-envs, want %d", len(infraEnvs), len(expectedInfraEnvs))
	}

	for i, infraEnv := range infraEnvs {
		if infraEnv.ID != expectedInfraEnvs[i].ID {
			t.Errorf("ListInfraEnvs()[%d].ID = %v, want %v", i, infraEnv.ID, expectedInfraEnvs[i].ID)
		}
	}
}