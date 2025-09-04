package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccInfraEnvResource(t *testing.T) {
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
				Config: testAccInfraEnvResourceConfig("test-infra-env"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInfraEnvExists("oai_infra_env.test"),
					resource.TestCheckResourceAttr("oai_infra_env.test", "name", "test-infra-env"),
					resource.TestCheckResourceAttr("oai_infra_env.test", "cpu_architecture", "x86_64"),
					resource.TestCheckResourceAttr("oai_infra_env.test", "image_type", "minimal-iso"),
					resource.TestCheckResourceAttrSet("oai_infra_env.test", "id"),
					resource.TestCheckResourceAttrSet("oai_infra_env.test", "download_url"),
					resource.TestCheckResourceAttrSet("oai_infra_env.test", "expires_at"),
					resource.TestCheckResourceAttrSet("oai_infra_env.test", "type"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "oai_infra_env.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"pull_secret"},
			},
		},
	})
}

func TestAccInfraEnvResource_Complete(t *testing.T) {
	// Skip test if not running acceptance tests
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInfraEnvResourceCompleteConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInfraEnvExists("oai_infra_env.test"),
					resource.TestCheckResourceAttr("oai_infra_env.test", "name", "complete-infra-env"),
					resource.TestCheckResourceAttr("oai_infra_env.test", "cpu_architecture", "x86_64"),
					resource.TestCheckResourceAttr("oai_infra_env.test", "image_type", "full-iso"),
					resource.TestCheckResourceAttr("oai_infra_env.test", "ssh_authorized_key", "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC..."),
					resource.TestCheckResourceAttr("oai_infra_env.test", "additional_ntp_sources", "ntp1.example.com,ntp2.example.com"),
					resource.TestCheckResourceAttr("oai_infra_env.test", "additional_trust_bundle", "-----BEGIN CERTIFICATE-----\nMIIC..."),
					resource.TestCheckResourceAttr("oai_infra_env.test", "proxy.http_proxy", "http://proxy.example.com:8080"),
					resource.TestCheckResourceAttr("oai_infra_env.test", "proxy.https_proxy", "https://proxy.example.com:8443"),
					resource.TestCheckResourceAttr("oai_infra_env.test", "proxy.no_proxy", "localhost,127.0.0.1,.example.com"),
				),
			},
		},
	})
}

func testAccCheckInfraEnvExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No InfraEnv ID is set")
		}

		// Here you would typically call the API to verify the infra env exists
		// For now, we'll just check that the ID is set
		return nil
	}
}

func testAccInfraEnvResourceConfig(name string) string {
	return fmt.Sprintf(`
provider "oai" {
  endpoint = "https://api.openshift.com/api/assisted-install"
  token    = "test-token"
}

resource "oai_infra_env" "test" {
  name             = %[1]q
  pull_secret      = "fake-pull-secret-for-testing"
  cpu_architecture = "x86_64"
  image_type       = "minimal-iso"
}
`, name)
}

func testAccInfraEnvResourceCompleteConfig() string {
	return `
provider "oai" {
  endpoint = "https://api.openshift.com/api/assisted-install"
  token    = "test-token"
}

resource "oai_infra_env" "test" {
  name                    = "complete-infra-env"
  pull_secret             = "fake-pull-secret-for-testing"
  cpu_architecture        = "x86_64"
  image_type              = "full-iso"
  ssh_authorized_key      = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC..."
  openshift_version       = "4.15.20"
  additional_ntp_sources  = "ntp1.example.com,ntp2.example.com"
  additional_trust_bundle = "-----BEGIN CERTIFICATE-----\nMIIC..."
  
  proxy {
    http_proxy  = "http://proxy.example.com:8080"
    https_proxy = "https://proxy.example.com:8443" 
    no_proxy    = "localhost,127.0.0.1,.example.com"
  }
  
  kernel_arguments {
    operation = "append"
    value     = "console=ttyS0,115200n8"
  }
  
  static_network_config {
    network_yaml = "version: 2\nethernets:\n  eth0:\n    dhcp4: true"
    mac_interface_map {
      mac_address      = "aa:bb:cc:dd:ee:ff"
      logical_nic_name = "eth0"
    }
  }
}
`
}