package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccClusterResource(t *testing.T) {
	// Skip test if not running acceptance tests
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccClusterResourceConfig("test-cluster"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists("oai_cluster.test"),
					resource.TestCheckResourceAttr("oai_cluster.test", "name", "test-cluster"),
					resource.TestCheckResourceAttr("oai_cluster.test", "openshift_version", "4.15.20"),
					resource.TestCheckResourceAttr("oai_cluster.test", "cpu_architecture", "x86_64"),
					resource.TestCheckResourceAttr("oai_cluster.test", "base_dns_domain", "example.com"),
					resource.TestCheckResourceAttr("oai_cluster.test", "cluster_network_cidr", "10.128.0.0/14"),
					resource.TestCheckResourceAttr("oai_cluster.test", "service_network_cidr", "172.30.0.0/16"),
					resource.TestCheckResourceAttr("oai_cluster.test", "trigger_installation", "false"),
					resource.TestCheckResourceAttrSet("oai_cluster.test", "id"),
					resource.TestCheckResourceAttrSet("oai_cluster.test", "status"),
					resource.TestCheckResourceAttrSet("oai_cluster.test", "kind"),
					resource.TestCheckResourceAttrSet("oai_cluster.test", "href"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "oai_cluster.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"pull_secret", "timeouts"},
			},
			// Update and Read testing
			{
				Config: testAccClusterResourceConfig("test-cluster-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("oai_cluster.test", "name", "test-cluster-updated"),
				),
			},
		},
	})
}

func TestAccClusterResource_Complete(t *testing.T) {
	// Skip test if not running acceptance tests
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with all optional attributes
			{
				Config: testAccClusterResourceCompleteConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists("oai_cluster.test"),
					resource.TestCheckResourceAttr("oai_cluster.test", "name", "complete-cluster"),
					resource.TestCheckResourceAttr("oai_cluster.test", "openshift_version", "4.15.20"),
					resource.TestCheckResourceAttr("oai_cluster.test", "cpu_architecture", "x86_64"),
					resource.TestCheckResourceAttr("oai_cluster.test", "base_dns_domain", "example.com"),
					resource.TestCheckResourceAttr("oai_cluster.test", "api_vips.0.ip", "192.168.1.100"),
					resource.TestCheckResourceAttr("oai_cluster.test", "ingress_vips.0.ip", "192.168.1.101"),
					resource.TestCheckResourceAttr("oai_cluster.test", "cluster_network_cidr", "10.128.0.0/14"),
					resource.TestCheckResourceAttr("oai_cluster.test", "cluster_network_host_prefix", "23"),
					resource.TestCheckResourceAttr("oai_cluster.test", "service_network_cidr", "172.30.0.0/16"),
					resource.TestCheckResourceAttr("oai_cluster.test", "vip_dhcp_allocation", "false"),
					resource.TestCheckResourceAttr("oai_cluster.test", "user_managed_networking", "false"),
					resource.TestCheckResourceAttr("oai_cluster.test", "control_plane_count", "3"),
					resource.TestCheckResourceAttr("oai_cluster.test", "schedulable_masters", "false"),
					resource.TestCheckResourceAttr("oai_cluster.test", "high_availability_mode", "Full"),
					resource.TestCheckResourceAttr("oai_cluster.test", "hyperthreading", "Enabled"),
					resource.TestCheckResourceAttr("oai_cluster.test", "network_type", "OVNKubernetes"),
					resource.TestCheckResourceAttr("oai_cluster.test", "tags", "test,cluster,complete"),
					resource.TestCheckResourceAttr("oai_cluster.test", "trigger_installation", "false"),
					resource.TestCheckResourceAttr("oai_cluster.test", "olm_operators.0.name", "local-storage-operator"),
					resource.TestCheckResourceAttr("oai_cluster.test", "olm_operators.0.properties", "{\"version\":\"4.15\"}"),
					resource.TestCheckResourceAttr("oai_cluster.test", "olm_operators.1.name", "odf-operator"),
					resource.TestCheckResourceAttr("oai_cluster.test", "olm_operators.1.properties", "{\"version\":\"4.15\",\"namespace\":\"openshift-storage\"}"),
					resource.TestCheckResourceAttrSet("oai_cluster.test", "ssh_public_key"),
					resource.TestCheckResourceAttrSet("oai_cluster.test", "additional_ntp_source"),
				),
			},
		},
	})
}

func TestAccClusterResource_SNO(t *testing.T) {
	// Skip test if not running acceptance tests
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterResourceSNOConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists("oai_cluster.test"),
					resource.TestCheckResourceAttr("oai_cluster.test", "name", "sno-cluster"),
					resource.TestCheckResourceAttr("oai_cluster.test", "control_plane_count", "1"),
					resource.TestCheckResourceAttr("oai_cluster.test", "schedulable_masters", "true"), // SNO defaults to true
					resource.TestCheckResourceAttr("oai_cluster.test", "cpu_architecture", "x86_64"),
				),
			},
		},
	})
}

func TestAccClusterResource_WithProxy(t *testing.T) {
	// Skip test if not running acceptance tests
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterResourceWithProxyConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists("oai_cluster.test"),
					resource.TestCheckResourceAttr("oai_cluster.test", "http_proxy", "http://proxy.example.com:8080"),
					resource.TestCheckResourceAttr("oai_cluster.test", "https_proxy", "https://proxy.example.com:8443"),
					resource.TestCheckResourceAttr("oai_cluster.test", "no_proxy", "localhost,127.0.0.1,.example.com"),
				),
			},
		},
	})
}

func testAccCheckClusterExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Cluster ID is set")
		}

		// Here you would typically call the API to verify the cluster exists
		// For now, we'll just check that the ID is set
		return nil
	}
}

func testAccClusterResourceConfig(name string) string {
	return fmt.Sprintf(`
provider "oai" {
  endpoint = "https://api.openshift.com/api/assisted-install"
  token    = "test-token"
}

resource "oai_cluster" "test" {
  name              = %[1]q
  openshift_version = "4.15.20"
  pull_secret       = "fake-pull-secret-for-testing"
  cpu_architecture  = "x86_64"
  
  base_dns_domain      = "example.com"
  cluster_network_cidr = "10.128.0.0/14"
  service_network_cidr = "172.30.0.0/16"
  trigger_installation = false  # Don't trigger installation in tests
  
  timeouts {
    create = "90m"
    update = "30m"
  }
}
`, name)
}

func testAccClusterResourceCompleteConfig() string {
	return `
provider "oai" {
  endpoint = "https://api.openshift.com/api/assisted-install"
  token    = "test-token"
}

resource "oai_cluster" "test" {
  name              = "complete-cluster"
  openshift_version = "4.15.20"
  pull_secret       = "fake-pull-secret-for-testing"
  cpu_architecture  = "x86_64"
  
  base_dns_domain          = "example.com"
  api_vips = [
    {
      ip = "192.168.1.100"
    }
  ]
  ingress_vips = [
    {
      ip = "192.168.1.101"  
    }
  ]
  cluster_network_cidr     = "10.128.0.0/14"
  cluster_network_host_prefix = 23
  service_network_cidr     = "172.30.0.0/16"
  
  ssh_public_key          = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC..."
  vip_dhcp_allocation     = false
  user_managed_networking = false
  additional_ntp_source   = "ntp.example.com"
  
  control_plane_count    = 3
  schedulable_masters    = false  # Don't schedule workloads on masters for multi-node
  high_availability_mode = "Full"  # Deprecated, but still supported for backward compatibility
  hyperthreading        = "Enabled"
  network_type          = "OVNKubernetes"
  tags                  = "test,cluster,complete"
  trigger_installation  = false  # Don't trigger installation in tests
  
  olm_operators {
    name       = "local-storage-operator"
    properties = "{\"version\":\"4.15\"}"
  }
  
  olm_operators {
    name       = "odf-operator"
    properties = "{\"version\":\"4.15\",\"namespace\":\"openshift-storage\"}"
  }
  
  timeouts {
    create = "90m"
    update = "30m"
  }
}
`
}

func testAccClusterResourceSNOConfig() string {
	return `
provider "oai" {
  endpoint = "https://api.openshift.com/api/assisted-install"
  token    = "test-token"
}

resource "oai_cluster" "test" {
  name              = "sno-cluster"
  openshift_version = "4.15.20"
  pull_secret       = "fake-pull-secret-for-testing"
  cpu_architecture  = "x86_64"
  
  base_dns_domain      = "example.com"
  control_plane_count  = 1           # Single node
  schedulable_masters  = true        # Allow workloads on master for SNO
  trigger_installation = false       # Don't trigger installation in tests
  
  # SNO-specific configuration
  cluster_network_cidr = "10.128.0.0/14"
  service_network_cidr = "172.30.0.0/16"
  
  # Optional: Add minimal operators for functionality  
  olm_operators {
    name = "lso"  # Local Storage Operator for storage
  }
  
  timeouts {
    create = "90m"
    update = "30m"
  }
}
`
}

func testAccClusterResourceWithProxyConfig() string {
	return `
provider "oai" {
  endpoint = "https://api.openshift.com/api/assisted-install"
  token    = "test-token"
}

resource "oai_cluster" "test" {
  name              = "proxy-cluster"
  openshift_version = "4.15.20"
  pull_secret       = "fake-pull-secret-for-testing"
  cpu_architecture  = "x86_64"
  
  base_dns_domain      = "example.com"
  trigger_installation = false  # Don't trigger installation in tests
  
  http_proxy  = "http://proxy.example.com:8080"
  https_proxy = "https://proxy.example.com:8443"
  no_proxy    = "localhost,127.0.0.1,.example.com"
  
  timeouts {
    create = "90m"
    update = "30m"
  }
}
`
}

func TestAccClusterResource_WithInstallationTrigger(t *testing.T) {
	// Skip test if not running acceptance tests
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterResourceWithInstallationTriggerConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists("oai_cluster.test"),
					resource.TestCheckResourceAttr("oai_cluster.test", "name", "trigger-test-cluster"),
					resource.TestCheckResourceAttr("oai_cluster.test", "cpu_architecture", "x86_64"),
					resource.TestCheckResourceAttr("oai_cluster.test", "control_plane_count", "1"),
					resource.TestCheckResourceAttr("oai_cluster.test", "trigger_installation", "true"),
					// In a real test, we'd check that status eventually becomes "installed"
					// but since we're not actually installing, we just verify the config
				),
			},
		},
	})
}

func TestAccClusterResource_WithOLMOperators(t *testing.T) {
	// Skip test if not running acceptance tests
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterResourceWithOLMOperatorsConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists("oai_cluster.test"),
					resource.TestCheckResourceAttr("oai_cluster.test", "name", "olm-test-cluster"),
					resource.TestCheckResourceAttr("oai_cluster.test", "cpu_architecture", "x86_64"),
					resource.TestCheckResourceAttr("oai_cluster.test", "trigger_installation", "false"),
					resource.TestCheckResourceAttr("oai_cluster.test", "olm_operators.#", "3"),
					resource.TestCheckResourceAttr("oai_cluster.test", "olm_operators.0.name", "local-storage-operator"),
					resource.TestCheckResourceAttr("oai_cluster.test", "olm_operators.0.properties", "{\"version\":\"4.15\"}"),
					resource.TestCheckResourceAttr("oai_cluster.test", "olm_operators.1.name", "odf-operator"),
					resource.TestCheckResourceAttr("oai_cluster.test", "olm_operators.2.name", "cnv-operator"),
				),
			},
		},
	})
}

func testAccClusterResourceWithInstallationTriggerConfig() string {
	return `
provider "oai" {
  endpoint = "https://api.openshift.com/api/assisted-install"
  token    = "test-token"
}

resource "oai_cluster" "test" {
  name              = "trigger-test-cluster"
  openshift_version = "4.15.20"
  pull_secret       = "fake-pull-secret-for-testing"
  cpu_architecture  = "x86_64"
  
  base_dns_domain       = "example.com"
  control_plane_count   = 1  # SNO deployment
  trigger_installation  = true  # This would trigger installation in real scenario
  
  timeouts {
    create = "90m"
    update = "30m"
  }
}
`
}

func testAccClusterResourceWithOLMOperatorsConfig() string {
	return `
provider "oai" {
  endpoint = "https://api.openshift.com/api/assisted-install"
  token    = "test-token"
}

resource "oai_cluster" "test" {
  name              = "olm-test-cluster"
  openshift_version = "4.15.20"
  pull_secret       = "fake-pull-secret-for-testing"
  cpu_architecture  = "x86_64"
  
  base_dns_domain      = "example.com"
  control_plane_count  = 3
  trigger_installation = false  # Don't trigger installation in tests
  
  olm_operators {
    name       = "local-storage-operator"
    properties = "{\"version\":\"4.15\"}"
  }
  
  olm_operators {
    name       = "odf-operator"
    properties = "{\"version\":\"4.15\",\"namespace\":\"openshift-storage\"}"
  }
  
  olm_operators {
    name       = "cnv-operator"
  }
  
  timeouts {
    create = "90m"
    update = "30m"
  }
}
`
}