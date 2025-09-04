package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccHostResource(t *testing.T) {
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
				Config: testAccHostResourceConfig("test-host-123"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHostExists("oai_host.test"),
					resource.TestCheckResourceAttr("oai_host.test", "id", "test-host-123"),
					resource.TestCheckResourceAttr("oai_host.test", "role", "auto-assign"),
					resource.TestCheckResourceAttrSet("oai_host.test", "infra_env_id"),
					resource.TestCheckResourceAttrSet("oai_host.test", "status"),
					resource.TestCheckResourceAttrSet("oai_host.test", "status_info"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "oai_host.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccHostResource_Complete(t *testing.T) {
	// Skip test if not running acceptance tests
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccHostResourceCompleteConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHostExists("oai_host.test"),
					resource.TestCheckResourceAttr("oai_host.test", "id", "test-host-complete-456"),
					resource.TestCheckResourceAttr("oai_host.test", "requested_hostname", "worker-node-01"),
					resource.TestCheckResourceAttr("oai_host.test", "host_name", "worker-node-01.example.com"),
					resource.TestCheckResourceAttr("oai_host.test", "role", "worker"),
					resource.TestCheckResourceAttr("oai_host.test", "machine_config_pool_name", "worker"),
					resource.TestCheckResourceAttr("oai_host.test", "ignition_endpoint_token", "secret-token"),
					resource.TestCheckResourceAttr("oai_host.test", "disks_selected_config.0.id", "/dev/sda"),
					resource.TestCheckResourceAttr("oai_host.test", "disks_selected_config.0.role", "install"),
					resource.TestCheckResourceAttr("oai_host.test", "disks_skip_formatting.0.disk_id", "/dev/sdb"),
					resource.TestCheckResourceAttr("oai_host.test", "node_labels.0.key", "environment"),
					resource.TestCheckResourceAttr("oai_host.test", "node_labels.0.value", "production"),
					resource.TestCheckResourceAttr("oai_host.test", "node_labels.1.key", "tier"),
					resource.TestCheckResourceAttr("oai_host.test", "node_labels.1.value", "worker"),
					resource.TestCheckResourceAttr("oai_host.test", "ignition_endpoint_http_headers.0.key", "Authorization"),
					resource.TestCheckResourceAttr("oai_host.test", "ignition_endpoint_http_headers.0.value", "Bearer token123"),
				),
			},
		},
	})
}

func testAccCheckHostExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Host ID is set")
		}

		// Here you would typically call the API to verify the host exists
		// For now, we'll just check that the ID is set
		return nil
	}
}

func testAccHostResourceConfig(hostID string) string {
	return fmt.Sprintf(`
provider "oai" {
  endpoint = "https://api.openshift.com/api/assisted-install"
  token    = "test-token"
}

# Note: In a real scenario, you'd need an actual infra env ID from a discovered host
resource "oai_host" "test" {
  id           = %[1]q
  infra_env_id = "infra-env-12345"  # This would come from a real infra env
  role         = "auto-assign"
}
`, hostID)
}

func testAccHostResourceCompleteConfig() string {
	return `
provider "oai" {
  endpoint = "https://api.openshift.com/api/assisted-install"
  token    = "test-token"
}

# Note: In a real scenario, you'd need actual IDs from discovered resources
resource "oai_host" "test" {
  id                        = "test-host-complete-456"
  infra_env_id             = "infra-env-12345"
  cluster_id               = "cluster-789"
  requested_hostname       = "worker-node-01"
  host_name                = "worker-node-01.example.com"
  role                     = "worker"
  machine_config_pool_name = "worker"
  ignition_endpoint_token  = "secret-token"
  
  disks_selected_config {
    id   = "/dev/sda"
    role = "install"
  }
  
  disks_skip_formatting {
    disk_id = "/dev/sdb"
  }
  
  node_labels {
    key   = "environment"
    value = "production"
  }
  
  node_labels {
    key   = "tier"
    value = "worker"
  }
  
  ignition_endpoint_http_headers {
    key   = "Authorization"
    value = "Bearer token123"
  }
}
`
}