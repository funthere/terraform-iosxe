# This is required for Terraform 0.13+
terraform {
  required_providers {
    iosxe = {
      version = ">= 0.0.2"
      source  = "ductus/local/dd"
    }
  }
}

provider "iosxe" {
  address = "http://localhost"
  port    = "3001"
  token   = "mautidur"
}


resource "iosxe_interface_ethernet" "netsim_local" {
  host = "localhost:9992"
  type = "GigabitEthernet"
  number = "1"
  description = "local netsim device"
  mtu = 401
  shutdown = true
  ipv4_address = "75.75.75.1"
  ipv4_address_mask = "255.255.255.255"
  service_policy_input = "IN_asdasd"
  service_policy_output = "OUT_ert324sdf"
}
