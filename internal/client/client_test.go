package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/models"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		config   ClientConfig
		wantURL  string
		wantToken string
	}{
		{
			name: "default configuration",
			config: ClientConfig{
				OfflineToken: "test-token",
			},
			wantURL:  "https://api.openshift.com/api/assisted-install",
			wantToken: "test-token",
		},
		{
			name: "custom endpoint",
			config: ClientConfig{
				BaseURL: "https://custom.api.example.com",
				OfflineToken:   "custom-token",
			},
			wantURL:  "https://custom.api.example.com",
			wantToken: "custom-token",
		},
		{
			name: "custom timeout",
			config: ClientConfig{
				OfflineToken: "test-token",
				Timeout: 60 * time.Second,
			},
			wantURL:  "https://api.openshift.com/api/assisted-install",
			wantToken: "test-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.config)
			
			if client.baseURL != tt.wantURL {
				t.Errorf("NewClient() baseURL = %v, want %v", client.baseURL, tt.wantURL)
			}
			
			if client.offlineToken != tt.wantToken {
				t.Errorf("NewClient() token = %v, want %v", client.offlineToken, tt.wantToken)
			}
			
			if client.httpClient == nil {
				t.Error("NewClient() httpClient is nil")
			}
		})
	}
}

func TestClient_CreateCluster(t *testing.T) {
	expectedCluster := &models.Cluster{
		ID:               "test-cluster-id",
		Name:             "test-cluster",
		OpenshiftVersion: "4.15.20",
		Status:           "insufficient",
		StatusInfo:       "Waiting for hosts",
		Kind:             "Cluster",
		Href:             "/api/assisted-install/v2/clusters/test-cluster-id",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		
		if r.URL.Path != "/v2/clusters" {
			t.Errorf("Expected path /v2/clusters, got %s", r.URL.Path)
		}
		
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Authorization header 'Bearer test-token', got %s", r.Header.Get("Authorization"))
		}
		
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type header 'application/json', got %s", r.Header.Get("Content-Type"))
		}
		
		// Decode request body
		var params models.ClusterCreateParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}
		
		// Verify request params
		if params.Name != "test-cluster" {
			t.Errorf("Expected cluster name 'test-cluster', got %s", params.Name)
		}
		
		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(expectedCluster)
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		BaseURL: server.URL,
		OfflineToken: "test-token",
	})

	params := models.ClusterCreateParams{
		Name:             "test-cluster",
		OpenshiftVersion: "4.15.20",
		PullSecret:       "fake-pull-secret",
	}

	cluster, err := client.CreateCluster(context.Background(), params)
	if err != nil {
		t.Fatalf("CreateCluster() error = %v", err)
	}

	if cluster.ID != expectedCluster.ID {
		t.Errorf("CreateCluster() ID = %v, want %v", cluster.ID, expectedCluster.ID)
	}
	
	if cluster.Name != expectedCluster.Name {
		t.Errorf("CreateCluster() Name = %v, want %v", cluster.Name, expectedCluster.Name)
	}
}

func TestClient_GetCluster(t *testing.T) {
	expectedCluster := &models.Cluster{
		ID:               "test-cluster-id",
		Name:             "test-cluster",
		OpenshiftVersion: "4.15.20",
		Status:           "ready",
		StatusInfo:       "Ready to install",
		Kind:             "Cluster",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		
		if r.URL.Path != "/v2/clusters/test-cluster-id" {
			t.Errorf("Expected path /v2/clusters/test-cluster-id, got %s", r.URL.Path)
		}
		
		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedCluster)
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		BaseURL: server.URL,
		OfflineToken: "test-token",
	})

	cluster, err := client.GetCluster(context.Background(), "test-cluster-id")
	if err != nil {
		t.Fatalf("GetCluster() error = %v", err)
	}

	if cluster.ID != expectedCluster.ID {
		t.Errorf("GetCluster() ID = %v, want %v", cluster.ID, expectedCluster.ID)
	}
	
	if cluster.Status != expectedCluster.Status {
		t.Errorf("GetCluster() Status = %v, want %v", cluster.Status, expectedCluster.Status)
	}
}

func TestClient_DeleteCluster(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		
		if r.URL.Path != "/v2/clusters/test-cluster-id" {
			t.Errorf("Expected path /v2/clusters/test-cluster-id, got %s", r.URL.Path)
		}
		
		// Send response
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		BaseURL: server.URL,
		OfflineToken: "test-token",
	})

	err := client.DeleteCluster(context.Background(), "test-cluster-id")
	if err != nil {
		t.Fatalf("DeleteCluster() error = %v", err)
	}
}

func TestClient_ErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		wantErrorMsg   string
	}{
		{
			name:         "bad request",
			statusCode:   http.StatusBadRequest,
			responseBody: `{"message": "Invalid cluster configuration"}`,
			wantErrorMsg: "API request failed with status 400",
		},
		{
			name:         "unauthorized",
			statusCode:   http.StatusUnauthorized,
			responseBody: `{"message": "Invalid token"}`,
			wantErrorMsg: "API request failed with status 401",
		},
		{
			name:         "not found",
			statusCode:   http.StatusNotFound,
			responseBody: `{"message": "Cluster not found"}`,
			wantErrorMsg: "API request failed with status 404",
		},
		{
			name:         "internal server error",
			statusCode:   http.StatusInternalServerError,
			responseBody: `{"message": "Internal server error"}`,
			wantErrorMsg: "API request failed with status 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := NewClient(ClientConfig{
				BaseURL: server.URL,
				OfflineToken: "test-token",
			})

			_, err := client.GetCluster(context.Background(), "test-cluster-id")
			if err == nil {
				t.Fatal("Expected error but got none")
			}

			if err.Error()[:len(tt.wantErrorMsg)] != tt.wantErrorMsg {
				t.Errorf("Error message = %v, want prefix %v", err.Error(), tt.wantErrorMsg)
			}
		})
	}
}

func TestClient_ListClusters(t *testing.T) {
	expectedClusters := []models.Cluster{
		{
			ID:   "cluster-1",
			Name: "cluster-1",
		},
		{
			ID:   "cluster-2",
			Name: "cluster-2",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		
		if r.URL.Path != "/v2/clusters" {
			t.Errorf("Expected path /v2/clusters, got %s", r.URL.Path)
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedClusters)
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		BaseURL: server.URL,
		OfflineToken: "test-token",
	})

	clusters, err := client.ListClusters(context.Background())
	if err != nil {
		t.Fatalf("ListClusters() error = %v", err)
	}

	if len(clusters) != len(expectedClusters) {
		t.Errorf("ListClusters() returned %d clusters, want %d", len(clusters), len(expectedClusters))
	}

	for i, cluster := range clusters {
		if cluster.ID != expectedClusters[i].ID {
			t.Errorf("ListClusters()[%d].ID = %v, want %v", i, cluster.ID, expectedClusters[i].ID)
		}
	}
}

func TestClient_DownloadManifestContent(t *testing.T) {
	expectedContent := "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: test-manifest"
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/clusters/cluster-123/manifests/files" {
			// Verify query parameters
			if r.URL.Query().Get("file_name") != "test.yaml" {
				t.Errorf("Expected file_name=test.yaml, got %s", r.URL.Query().Get("file_name"))
			}
			if r.URL.Query().Get("folder") != "manifests" {
				t.Errorf("Expected folder=manifests, got %s", r.URL.Query().Get("folder"))
			}
			
			w.Header().Set("Content-Type", "application/octet-stream")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(expectedContent))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		BaseURL: server.URL,
		OfflineToken: "test-token",
	})

	content, err := client.DownloadManifestContent(context.Background(), "cluster-123", "test.yaml", "manifests")
	if err != nil {
		t.Fatalf("DownloadManifestContent() error = %v", err)
	}

	if content != expectedContent {
		t.Errorf("DownloadManifestContent() = %v, want %v", content, expectedContent)
	}
}