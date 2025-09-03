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

func TestClient_UpdateCluster(t *testing.T) {
	expectedCluster := &models.Cluster{
		ID:               "test-cluster-id",
		Name:             "updated-cluster",
		OpenshiftVersion: "4.15.20",
		Status:           "ready",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("Expected PATCH request, got %s", r.Method)
		}
		
		if r.URL.Path != "/v2/clusters/test-cluster-id" {
			t.Errorf("Expected path /v2/clusters/test-cluster-id, got %s", r.URL.Path)
		}

		var params models.ClusterUpdateParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedCluster)
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		BaseURL: server.URL,
		OfflineToken: "test-token",
	})

	name := "updated-cluster"
	params := models.ClusterUpdateParams{
		Name: &name,
	}

	cluster, err := client.UpdateCluster(context.Background(), "test-cluster-id", params)
	if err != nil {
		t.Fatalf("UpdateCluster() error = %v", err)
	}

	if cluster.Name != expectedCluster.Name {
		t.Errorf("UpdateCluster() Name = %v, want %v", cluster.Name, expectedCluster.Name)
	}
}

func TestClient_InstallCluster(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		
		if r.URL.Path != "/v2/clusters/test-cluster-id/actions/install" {
			t.Errorf("Expected path /v2/clusters/test-cluster-id/actions/install, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		BaseURL: server.URL,
		OfflineToken: "test-token",
	})

	err := client.InstallCluster(context.Background(), "test-cluster-id")
	if err != nil {
		t.Fatalf("InstallCluster() error = %v", err)
	}
}

func TestClient_InfraEnvOperations(t *testing.T) {
	t.Run("CreateInfraEnv", func(t *testing.T) {
		expectedInfraEnv := &models.InfraEnv{
			ID:               "infra-env-id",
			Name:             "test-infra-env",
			OpenshiftVersion: "4.15.20",
			Type:             "full-iso",
			DownloadURL:      "https://example.com/download/iso",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" || r.URL.Path != "/v2/infra-envs" {
				t.Errorf("Expected POST /v2/infra-envs, got %s %s", r.Method, r.URL.Path)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(expectedInfraEnv)
		}))
		defer server.Close()

		client := NewClient(ClientConfig{
			BaseURL: server.URL,
			OfflineToken: "test-token",
		})

		params := models.InfraEnvCreateParams{
			Name:       "test-infra-env",
			PullSecret: "fake-pull-secret",
		}

		infraEnv, err := client.CreateInfraEnv(context.Background(), params)
		if err != nil {
			t.Fatalf("CreateInfraEnv() error = %v", err)
		}

		if infraEnv.ID != expectedInfraEnv.ID {
			t.Errorf("CreateInfraEnv() ID = %v, want %v", infraEnv.ID, expectedInfraEnv.ID)
		}
	})

	t.Run("GetInfraEnv", func(t *testing.T) {
		expectedInfraEnv := &models.InfraEnv{
			ID:   "infra-env-id",
			Name: "test-infra-env",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" || r.URL.Path != "/v2/infra-envs/infra-env-id" {
				t.Errorf("Expected GET /v2/infra-envs/infra-env-id, got %s %s", r.Method, r.URL.Path)
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(expectedInfraEnv)
		}))
		defer server.Close()

		client := NewClient(ClientConfig{
			BaseURL: server.URL,
			OfflineToken: "test-token",
		})

		infraEnv, err := client.GetInfraEnv(context.Background(), "infra-env-id")
		if err != nil {
			t.Fatalf("GetInfraEnv() error = %v", err)
		}

		if infraEnv.ID != expectedInfraEnv.ID {
			t.Errorf("GetInfraEnv() ID = %v, want %v", infraEnv.ID, expectedInfraEnv.ID)
		}
	})

	t.Run("DeleteInfraEnv", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "DELETE" || r.URL.Path != "/v2/infra-envs/infra-env-id" {
				t.Errorf("Expected DELETE /v2/infra-envs/infra-env-id, got %s %s", r.Method, r.URL.Path)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client := NewClient(ClientConfig{
			BaseURL: server.URL,
			OfflineToken: "test-token",
		})

		err := client.DeleteInfraEnv(context.Background(), "infra-env-id")
		if err != nil {
			t.Fatalf("DeleteInfraEnv() error = %v", err)
		}
	})
}

func TestClient_ManifestOperations(t *testing.T) {
	t.Run("CreateManifest", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" || r.URL.Path != "/v2/clusters/cluster-id/manifests" {
				t.Errorf("Expected POST /v2/clusters/cluster-id/manifests, got %s %s", r.Method, r.URL.Path)
			}

			var params models.CreateManifestParams
			if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			if params.FileName != "test.yaml" {
				t.Errorf("Expected filename 'test.yaml', got %s", params.FileName)
			}

			w.WriteHeader(http.StatusCreated)
		}))
		defer server.Close()

		client := NewClient(ClientConfig{
			BaseURL: server.URL,
			OfflineToken: "test-token",
		})

		params := models.CreateManifestParams{
			Folder:   "manifests",
			FileName: "test.yaml",
			Content:  "base64content",
		}

		err := client.CreateManifest(context.Background(), "cluster-id", params)
		if err != nil {
			t.Fatalf("CreateManifest() error = %v", err)
		}
	})

	t.Run("ListManifests", func(t *testing.T) {
		expectedManifests := []models.Manifest{
			{
				Folder:   "manifests",
				FileName: "config1.yaml",
			},
			{
				Folder:   "openshift",
				FileName: "config2.yaml",
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" || r.URL.Path != "/v2/clusters/cluster-id/manifests" {
				t.Errorf("Expected GET /v2/clusters/cluster-id/manifests, got %s %s", r.Method, r.URL.Path)
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(expectedManifests)
		}))
		defer server.Close()

		client := NewClient(ClientConfig{
			BaseURL: server.URL,
			OfflineToken: "test-token",
		})

		manifests, err := client.ListManifests(context.Background(), "cluster-id")
		if err != nil {
			t.Fatalf("ListManifests() error = %v", err)
		}

		if len(manifests) != len(expectedManifests) {
			t.Errorf("ListManifests() returned %d manifests, want %d", len(manifests), len(expectedManifests))
		}
	})

	t.Run("DeleteManifest", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "DELETE" {
				t.Errorf("Expected DELETE request, got %s", r.Method)
			}

			// Check query parameters
			folder := r.URL.Query().Get("folder")
			fileName := r.URL.Query().Get("file_name")

			if folder != "manifests" {
				t.Errorf("Expected folder 'manifests', got %s", folder)
			}
			if fileName != "test.yaml" {
				t.Errorf("Expected file_name 'test.yaml', got %s", fileName)
			}

			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client := NewClient(ClientConfig{
			BaseURL: server.URL,
			OfflineToken: "test-token",
		})

		err := client.DeleteManifest(context.Background(), "cluster-id", "manifests", "test.yaml")
		if err != nil {
			t.Fatalf("DeleteManifest() error = %v", err)
		}
	})
}

func TestClient_VersionsAndOperators(t *testing.T) {
	t.Run("GetOpenShiftVersions", func(t *testing.T) {
		expectedVersions := models.OpenshiftVersions{
			"4.15.20": models.OpenshiftVersion{
				DisplayName:  "OpenShift 4.15.20",
				SupportLevel: "production",
				Default:      true,
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("Expected GET request, got %s", r.Method)
			}

			// Check query parameters
			version := r.URL.Query().Get("version")
			onlyLatest := r.URL.Query().Get("only_latest")

			if version != "" && version != "4.15" {
				t.Errorf("Unexpected version parameter: %s", version)
			}
			if onlyLatest != "" && onlyLatest != "true" {
				t.Errorf("Unexpected only_latest parameter: %s", onlyLatest)
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(expectedVersions)
		}))
		defer server.Close()

		client := NewClient(ClientConfig{
			BaseURL: server.URL,
			OfflineToken: "test-token",
		})

		versions, err := client.GetOpenShiftVersions(context.Background(), "4.15", true)
		if err != nil {
			t.Fatalf("GetOpenShiftVersions() error = %v", err)
		}

		if len(*versions) != 1 {
			t.Errorf("GetOpenShiftVersions() returned %d versions, want 1", len(*versions))
		}
	})

	t.Run("GetSupportedOperators", func(t *testing.T) {
		expectedOperators := []string{"odf", "cnv", "lso", "mce"}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" || r.URL.Path != "/v2/supported-operators" {
				t.Errorf("Expected GET /v2/supported-operators, got %s %s", r.Method, r.URL.Path)
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(expectedOperators)
		}))
		defer server.Close()

		client := NewClient(ClientConfig{
			BaseURL: server.URL,
			OfflineToken: "test-token",
		})

		operators, err := client.GetSupportedOperators(context.Background())
		if err != nil {
			t.Fatalf("GetSupportedOperators() error = %v", err)
		}

		if len(operators) != len(expectedOperators) {
			t.Errorf("GetSupportedOperators() returned %d operators, want %d", len(operators), len(expectedOperators))
		}
	})
}

func TestClient_HostOperations(t *testing.T) {
	t.Run("ListHosts", func(t *testing.T) {
		expectedHosts := []models.Host{
			{
				ID:         "host-1",
				InfraEnvID: "infra-env-id",
				Status:     "known",
			},
			{
				ID:         "host-2",
				InfraEnvID: "infra-env-id",
				Status:     "discovering",
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" || r.URL.Path != "/v2/infra-envs/infra-env-id/hosts" {
				t.Errorf("Expected GET /v2/infra-envs/infra-env-id/hosts, got %s %s", r.Method, r.URL.Path)
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(expectedHosts)
		}))
		defer server.Close()

		client := NewClient(ClientConfig{
			BaseURL: server.URL,
			OfflineToken: "test-token",
		})

		hosts, err := client.ListHosts(context.Background(), "infra-env-id")
		if err != nil {
			t.Fatalf("ListHosts() error = %v", err)
		}

		if len(hosts) != len(expectedHosts) {
			t.Errorf("ListHosts() returned %d hosts, want %d", len(hosts), len(expectedHosts))
		}
	})

	t.Run("BindHost", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" || r.URL.Path != "/v2/infra-envs/infra-env-id/hosts/host-id/actions/bind" {
				t.Errorf("Expected POST /v2/infra-envs/infra-env-id/hosts/host-id/actions/bind, got %s %s", r.Method, r.URL.Path)
			}

			var params models.BindHostParams
			if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			if params.ClusterID != "cluster-id" {
				t.Errorf("Expected cluster_id 'cluster-id', got %s", params.ClusterID)
			}

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClient(ClientConfig{
			BaseURL: server.URL,
			OfflineToken: "test-token",
		})

		params := models.BindHostParams{
			ClusterID: "cluster-id",
		}

		err := client.BindHost(context.Background(), "infra-env-id", "host-id", params)
		if err != nil {
			t.Fatalf("BindHost() error = %v", err)
		}
	})
}

func TestClient_ContextCancellation(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		BaseURL: server.URL,
		OfflineToken: "test-token",
	})

	// Create a context that will be cancelled immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.GetCluster(ctx, "test-cluster-id")
	if err == nil {
		t.Error("Expected error due to cancelled context, got none")
	}
}

func TestClient_Timeout(t *testing.T) {
	// Create a server that delays response longer than timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		BaseURL: server.URL,
		OfflineToken: "test-token",
		Timeout: 50 * time.Millisecond,
	})

	_, err := client.GetCluster(context.Background(), "test-cluster-id")
	if err == nil {
		t.Error("Expected timeout error, got none")
	}
}