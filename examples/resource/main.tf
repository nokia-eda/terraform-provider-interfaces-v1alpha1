resource "interfaces-v1alpha1_interface" "leaf1_ethernet_1_14" {
  metadata = {
    labels = {
      "eda.nokia.com/role" = "edge"
      "terraform" : "true"
    }
    name      = "leaf1-ethernet-1-14"
    namespace = "eda"
  }
  spec = {
    description = "created by terraform"
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
