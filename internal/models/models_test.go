package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestCluster_JSONMarshal(t *testing.T) {
	cluster := &Cluster{
		Kind:             "Cluster",
		ID:               "test-id",
		Name:             "test-cluster",
		OpenshiftVersion: "4.15.20",
		BaseDNSDomain:    "example.com",
		APIVips:          []APIVip{{IP: "192.168.1.100"}},
		IngressVips:      []IngressVip{{IP: "192.168.1.101"}},
		Status:           "ready",
		StatusInfo:       "Ready to install",
		CreatedAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(cluster)
	if err != nil {
		t.Fatalf("Failed to marshal cluster: %v", err)
	}

	var unmarshaled Cluster
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal cluster: %v", err)
	}

	if unmarshaled.ID != cluster.ID {
		t.Errorf("ID mismatch: got %s, want %s", unmarshaled.ID, cluster.ID)
	}

	if unmarshaled.Name != cluster.Name {
		t.Errorf("Name mismatch: got %s, want %s", unmarshaled.Name, cluster.Name)
	}

	if len(unmarshaled.APIVips) != len(cluster.APIVips) {
		t.Errorf("APIVips length mismatch: got %d, want %d", len(unmarshaled.APIVips), len(cluster.APIVips))
	}

	if len(cluster.APIVips) > 0 && unmarshaled.APIVips[0].IP != cluster.APIVips[0].IP {
		t.Errorf("APIVips[0].IP mismatch: got %s, want %s", unmarshaled.APIVips[0].IP, cluster.APIVips[0].IP)
	}

	if len(unmarshaled.IngressVips) != len(cluster.IngressVips) {
		t.Errorf("IngressVips length mismatch: got %d, want %d", len(unmarshaled.IngressVips), len(cluster.IngressVips))
	}

	if len(cluster.IngressVips) > 0 && unmarshaled.IngressVips[0].IP != cluster.IngressVips[0].IP {
		t.Errorf("IngressVips[0].IP mismatch: got %s, want %s", unmarshaled.IngressVips[0].IP, cluster.IngressVips[0].IP)
	}
}

func TestCluster_NewFields(t *testing.T) {
	cluster := &Cluster{
		Kind:               "Cluster",
		ID:                 "test-new-fields",
		Name:               "test-cluster-new",
		OpenshiftVersion:   "4.15.20",
		OCPReleaseImage:    "quay.io/openshift-release-dev/ocp-release:4.15.20-x86_64",
		SchedulableMasters: true,
		Tags:               "test,swagger-compliant,new-fields",
		APIVips: []APIVip{
			{IP: "192.168.1.100"},
			{IP: "192.168.1.200"},
		},
		IngressVips: []IngressVip{
			{IP: "192.168.1.101"},
		},
		ClusterNetworks: []ClusterNetwork{
			{CIDR: "10.128.0.0/14", HostPrefix: 23},
		},
		ServiceNetworks: []ServiceNetwork{
			{CIDR: "172.30.0.0/16"},
		},
		MachineNetworks: []MachineNetwork{
			{CIDR: "192.168.1.0/24"},
		},
		Status:     "ready",
		StatusInfo: "Ready with new fields",
	}

	// Test JSON marshaling/unmarshaling
	data, err := json.Marshal(cluster)
	if err != nil {
		t.Fatalf("Failed to marshal cluster with new fields: %v", err)
	}

	var unmarshaled Cluster
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal cluster with new fields: %v", err)
	}

	// Validate new fields
	if unmarshaled.SchedulableMasters != cluster.SchedulableMasters {
		t.Errorf("SchedulableMasters mismatch: got %v, want %v", unmarshaled.SchedulableMasters, cluster.SchedulableMasters)
	}

	if unmarshaled.Tags != cluster.Tags {
		t.Errorf("Tags mismatch: got %s, want %s", unmarshaled.Tags, cluster.Tags)
	}

	if unmarshaled.OCPReleaseImage != cluster.OCPReleaseImage {
		t.Errorf("OCPReleaseImage mismatch: got %s, want %s", unmarshaled.OCPReleaseImage, cluster.OCPReleaseImage)
	}

	// Test multiple VIPs
	if len(unmarshaled.APIVips) != 2 {
		t.Errorf("Expected 2 API VIPs, got %d", len(unmarshaled.APIVips))
	}

	if unmarshaled.APIVips[0].IP != "192.168.1.100" {
		t.Errorf("First API VIP mismatch: got %s, want %s", unmarshaled.APIVips[0].IP, "192.168.1.100")
	}

	if unmarshaled.APIVips[1].IP != "192.168.1.200" {
		t.Errorf("Second API VIP mismatch: got %s, want %s", unmarshaled.APIVips[1].IP, "192.168.1.200")
	}

	// Test network arrays
	if len(unmarshaled.ClusterNetworks) != 1 {
		t.Errorf("Expected 1 cluster network, got %d", len(unmarshaled.ClusterNetworks))
	}

	if unmarshaled.ClusterNetworks[0].CIDR != "10.128.0.0/14" {
		t.Errorf("Cluster network CIDR mismatch: got %s, want %s", unmarshaled.ClusterNetworks[0].CIDR, "10.128.0.0/14")
	}
}

func TestClusterCreateParams_Validation(t *testing.T) {
	tests := []struct {
		name   string
		params ClusterCreateParams
		valid  bool
	}{
		{
			name: "valid minimal params",
			params: ClusterCreateParams{
				Name:             "test-cluster",
				OpenshiftVersion: "4.15.20",
				PullSecret:       "fake-secret",
				CPUArchitecture:  "x86_64",
			},
			valid: true,
		},
		{
			name: "valid complete params",
			params: ClusterCreateParams{
				Name:              "test-cluster",
				OpenshiftVersion:  "4.15.20",
				PullSecret:        "fake-secret",
				CPUArchitecture:   "x86_64",
				BaseDNSDomain:     "example.com",
				APIVips:           []APIVip{{IP: "192.168.1.100"}},
				IngressVips:       []IngressVip{{IP: "192.168.1.101"}},
				SSHPublicKey:      "ssh-rsa AAAA...",
				ControlPlaneCount: 3,
				OLMOperators: []OLMOperator{
					{Name: "local-storage-operator", Properties: "{\"version\":\"4.15\"}"},
				},
			},
			valid: true,
		},
		{
			name: "with proxy configuration",
			params: ClusterCreateParams{
				Name:             "test-cluster",
				OpenshiftVersion: "4.15.20",
				PullSecret:       "fake-secret",
				CPUArchitecture:  "x86_64",
				HTTPProxy:        "http://proxy.example.com:8080",
				HTTPSProxy:       "https://proxy.example.com:8443",
				NoProxy:          "localhost,127.0.0.1",
			},
			valid: true,
		},
		{
			name: "SNO configuration",
			params: ClusterCreateParams{
				Name:              "sno-cluster",
				OpenshiftVersion:  "4.15.20",
				PullSecret:        "fake-secret",
				CPUArchitecture:   "x86_64",
				ControlPlaneCount: 1,
			},
			valid: true,
		},
		{
			name: "multi-architecture",
			params: ClusterCreateParams{
				Name:              "multi-arch-cluster",
				OpenshiftVersion:  "4.15.20",
				PullSecret:        "fake-secret",
				CPUArchitecture:   "multi",
				ControlPlaneCount: 3,
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling
			data, err := json.Marshal(tt.params)
			if err != nil && tt.valid {
				t.Errorf("Failed to marshal valid params: %v", err)
			}

			// Test that required fields are present
			if tt.valid {
				if tt.params.Name == "" {
					t.Error("Valid params should have Name")
				}
				if tt.params.OpenshiftVersion == "" {
					t.Error("Valid params should have OpenshiftVersion")
				}
				if tt.params.PullSecret == "" {
					t.Error("Valid params should have PullSecret")
				}
			}

			// Verify JSON structure
			var jsonMap map[string]interface{}
			if err := json.Unmarshal(data, &jsonMap); err == nil {
				if _, ok := jsonMap["name"]; !ok && tt.valid {
					t.Error("JSON should include 'name' field")
				}
				if _, ok := jsonMap["openshift_version"]; !ok && tt.valid {
					t.Error("JSON should include 'openshift_version' field")
				}
			}
		})
	}
}

func TestInfraEnv_JSONMarshal(t *testing.T) {
	infraEnv := &InfraEnv{
		Kind:             "InfraEnv",
		ID:               "test-infra-id",
		Name:             "test-infra-env",
		OpenshiftVersion: "4.15.20",
		CPUArchitecture:  "x86_64",
		ClusterID:        "cluster-123",
		SSHAuthorizedKey: "ssh-rsa AAAA...",
		PullSecretSet:    true,
		Type:             "full-iso",
		DownloadURL:      "https://example.com/iso",
		CreatedAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(infraEnv)
	if err != nil {
		t.Fatalf("Failed to marshal infraEnv: %v", err)
	}

	var unmarshaled InfraEnv
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal infraEnv: %v", err)
	}

	if unmarshaled.ID != infraEnv.ID {
		t.Errorf("ID mismatch: got %s, want %s", unmarshaled.ID, infraEnv.ID)
	}

	if unmarshaled.CPUArchitecture != infraEnv.CPUArchitecture {
		t.Errorf("CPUArchitecture mismatch: got %s, want %s", unmarshaled.CPUArchitecture, infraEnv.CPUArchitecture)
	}

	if unmarshaled.PullSecretSet != infraEnv.PullSecretSet {
		t.Errorf("PullSecretSet mismatch: got %v, want %v", unmarshaled.PullSecretSet, infraEnv.PullSecretSet)
	}
}

func TestProxy_JSONMarshal(t *testing.T) {
	proxy := &Proxy{
		HTTPProxy:  "http://proxy.example.com:8080",
		HTTPSProxy: "https://proxy.example.com:8443",
		NoProxy:    "localhost,127.0.0.1,.example.com",
	}

	data, err := json.Marshal(proxy)
	if err != nil {
		t.Fatalf("Failed to marshal proxy: %v", err)
	}

	var unmarshaled Proxy
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal proxy: %v", err)
	}

	if unmarshaled.HTTPProxy != proxy.HTTPProxy {
		t.Errorf("HTTPProxy mismatch: got %s, want %s", unmarshaled.HTTPProxy, proxy.HTTPProxy)
	}

	if unmarshaled.HTTPSProxy != proxy.HTTPSProxy {
		t.Errorf("HTTPSProxy mismatch: got %s, want %s", unmarshaled.HTTPSProxy, proxy.HTTPSProxy)
	}

	if unmarshaled.NoProxy != proxy.NoProxy {
		t.Errorf("NoProxy mismatch: got %s, want %s", unmarshaled.NoProxy, proxy.NoProxy)
	}
}

func TestManifest_JSONMarshal(t *testing.T) {
	manifest := &Manifest{
		Folder:         "manifests",
		FileName:       "custom-config.yaml",
		ManifestSource: "user",
	}

	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("Failed to marshal manifest: %v", err)
	}

	var unmarshaled Manifest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal manifest: %v", err)
	}

	if unmarshaled.Folder != manifest.Folder {
		t.Errorf("Folder mismatch: got %s, want %s", unmarshaled.Folder, manifest.Folder)
	}

	if unmarshaled.FileName != manifest.FileName {
		t.Errorf("FileName mismatch: got %s, want %s", unmarshaled.FileName, manifest.FileName)
	}
}

func TestCreateManifestParams_JSONMarshal(t *testing.T) {
	params := &CreateManifestParams{
		Folder:   "manifests",
		FileName: "config.yaml",
		Content:  "YXBpVmVyc2lvbjogdjEK", // base64 encoded "apiVersion: v1\n"
	}

	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Failed to marshal CreateManifestParams: %v", err)
	}

	var unmarshaled CreateManifestParams
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal CreateManifestParams: %v", err)
	}

	if unmarshaled.Content != params.Content {
		t.Errorf("Content mismatch: got %s, want %s", unmarshaled.Content, params.Content)
	}
}

func TestOpenshiftVersions_JSONMarshal(t *testing.T) {
	versions := OpenshiftVersions{
		"4.15.20": OpenshiftVersion{
			DisplayName:      "OpenShift 4.15.20",
			SupportLevel:     "production",
			Default:          true,
			CPUArchitectures: []string{"x86_64", "aarch64"},
		},
		"4.14.15": OpenshiftVersion{
			DisplayName:      "OpenShift 4.14.15",
			SupportLevel:     "maintenance",
			Default:          false,
			CPUArchitectures: []string{"x86_64"},
		},
	}

	data, err := json.Marshal(versions)
	if err != nil {
		t.Fatalf("Failed to marshal OpenshiftVersions: %v", err)
	}

	var unmarshaled OpenshiftVersions
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal OpenshiftVersions: %v", err)
	}

	if len(unmarshaled) != len(versions) {
		t.Errorf("Version count mismatch: got %d, want %d", len(unmarshaled), len(versions))
	}

	v415, ok := unmarshaled["4.15.20"]
	if !ok {
		t.Error("Missing version 4.15.20")
	} else {
		if v415.DisplayName != "OpenShift 4.15.20" {
			t.Errorf("DisplayName mismatch: got %s, want %s", v415.DisplayName, "OpenShift 4.15.20")
		}
		if v415.SupportLevel != "production" {
			t.Errorf("SupportLevel mismatch: got %s, want %s", v415.SupportLevel, "production")
		}
		if !v415.Default {
			t.Error("Version 4.15.20 should be default")
		}
	}
}

func TestHost_JSONMarshal(t *testing.T) {
	host := &Host{
		Kind:              "Host",
		ID:                "host-123",
		ClusterID:         "cluster-456",
		InfraEnvID:        "infra-789",
		Status:            "known",
		StatusInfo:        "Host is ready",
		RequestedHostname: "worker-1",
		Role:              "worker",
		CreatedAt:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Progress: &Progress{
			CurrentStage: "Waiting for control plane",
			ProgressInfo: "Waiting for control plane to be ready",
		},
	}

	data, err := json.Marshal(host)
	if err != nil {
		t.Fatalf("Failed to marshal host: %v", err)
	}

	var unmarshaled Host
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal host: %v", err)
	}

	if unmarshaled.ID != host.ID {
		t.Errorf("ID mismatch: got %s, want %s", unmarshaled.ID, host.ID)
	}

	if unmarshaled.Status != host.Status {
		t.Errorf("Status mismatch: got %s, want %s", unmarshaled.Status, host.Status)
	}

	if unmarshaled.Progress == nil {
		t.Error("Progress should not be nil")
	} else if unmarshaled.Progress.CurrentStage != host.Progress.CurrentStage {
		t.Errorf("Progress.CurrentStage mismatch: got %s, want %s",
			unmarshaled.Progress.CurrentStage, host.Progress.CurrentStage)
	}
}

func TestPlatform_JSONMarshal(t *testing.T) {
	platform := &Platform{
		Type: "baremetal",
		Baremetal: &BaremetalPlatform{
			APIVips:     []string{"192.168.1.100"},
			IngressVips: []string{"192.168.1.101"},
		},
	}

	data, err := json.Marshal(platform)
	if err != nil {
		t.Fatalf("Failed to marshal platform: %v", err)
	}

	var unmarshaled Platform
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal platform: %v", err)
	}

	if unmarshaled.Type != platform.Type {
		t.Errorf("Type mismatch: got %s, want %s", unmarshaled.Type, platform.Type)
	}

	if unmarshaled.Baremetal == nil {
		t.Error("Baremetal platform should not be nil")
	} else {
		if len(unmarshaled.Baremetal.APIVips) != len(platform.Baremetal.APIVips) {
			t.Errorf("APIVips length mismatch: got %d, want %d",
				len(unmarshaled.Baremetal.APIVips), len(platform.Baremetal.APIVips))
		}
	}
}

func TestOLMOperator_JSONMarshal(t *testing.T) {
	operator := &OLMOperator{
		Name:       "odf",
		Properties: `{"storageClass": "gp3"}`,
	}

	data, err := json.Marshal(operator)
	if err != nil {
		t.Fatalf("Failed to marshal OLMOperator: %v", err)
	}

	var unmarshaled OLMOperator
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal OLMOperator: %v", err)
	}

	if unmarshaled.Name != operator.Name {
		t.Errorf("Name mismatch: got %s, want %s", unmarshaled.Name, operator.Name)
	}

	if unmarshaled.Properties != operator.Properties {
		t.Errorf("Properties mismatch: got %s, want %s", unmarshaled.Properties, operator.Properties)
	}
}
