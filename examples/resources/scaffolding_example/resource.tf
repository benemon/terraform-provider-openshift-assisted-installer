resource "oai_cluster" "example" {
  name             = "my-openshift-cluster"
  openshift_version = "4.15.20"
  pull_secret      = var.pull_secret
  
  # Optional networking configuration
  base_dns_domain = "example.com"
  api_vips        = ["192.168.1.100"]
  ingress_vips    = ["192.168.1.101"]
  
  # SSH access
  ssh_public_key = file("~/.ssh/id_rsa.pub")
  
  # High availability mode
  high_availability_mode = "Full"
  
  # Network configuration
  cluster_network_cidr = "10.128.0.0/14"
  service_network_cidr = "172.30.0.0/16"
  
  timeouts {
    create = "90m"
    update = "30m"
  }
}

variable "pull_secret" {
  description = "Red Hat pull secret for OpenShift installation"
  type        = string
  sensitive   = true
}
