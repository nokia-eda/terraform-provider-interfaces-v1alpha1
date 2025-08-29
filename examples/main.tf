# Read all interfaces.
data "interfaces-v1alpha1_interface_list" "all" {
  namespace     = "eda"
  labelselector = "eda.nokia.com/role=edge"
}

# Read a specific interface
data "interfaces-v1alpha1_interface" "leaf_1_if_1" {
  namespace = "eda"
  name      = "leaf-1-ethernet-1-1"
}

# Use `terraform plan -generate-config-out=interfaces.tf` 
# to automatically generate configs for the resources to be imported.
# Review the generated resources and move them to appropriate config file(s) if needed.
# Modify the resources if needed and start managing them using `terraform apply`.
#
# Optionally you can add this snippet inside the resource definitions to
# ignore changes to certain attributes and prevent accidental deletion, e.g:
#   lifecycle {
#     ignore_changes = [status]
#     prevent_destroy = true
#   }

import {
  to = interfaces-v1alpha1_interface.myif_1
  id = "eda/leaf-1-ethernet-1-1"
}

# import {
#  to = interfaces_interface.myif_2
#  id = "eda/leaf-1-ethernet-1-2"
# }

resource "interfaces-v1alpha1_interface" "myif_1" {
  api_version = "interfaces.eda.nokia.com/v1alpha1"
  kind        = "Interface"
  metadata = {
    labels = {
      "eda.nokia.com/role" = "interSwitch"
      "fabric_name" : "fabric-profile1-1"
    }
    name      = "leaf-1-ethernet-1-1"
    namespace = "eda"
  }
  spec = {
    description = "generated from terraform"
    enabled     = true
    lldp        = true
    members = [
      {
        enabled            = true
        interface          = "ethernet-1-1"
        lacp_port_priority = 32768
        node               = "leaf-1"
      },
    ]
    type = "interface"
  }
  lifecycle {
    ignore_changes = [status]
  }
}
