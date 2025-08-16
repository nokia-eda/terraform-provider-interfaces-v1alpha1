resource "interfaces-v1alpha1_interface" "leaf1_ethernet_1_14" {
  api_version = "interfaces.eda.nokia.com/v1alpha1"
  kind        = "Interface"
  metadata = {
    labels = {
      "eda.nokia.com/role" = "edge"
      "terraform" : "true"
    }
    name      = "leaf1-ethernet-1-14"
    namespace = "eda"
  }
  spec = {
    description = "generated from terraform"
    enabled     = true
    lldp        = true
    members = [
      {
        enabled   = true
        interface = "ethernet-1-14"
        node      = "leaf1"
      },
    ]
    type = "interface"
  }
}
