# Get all interfaces.
data "interfaces-v1alpha1_interface_list" "all" {
  namespace = "eda"
}

# Get all interfaces with label selector
data "interfaces-v1alpha1_interface_list" "interswitch" {
  namespace     = "eda"
  labelselector = "eda.nokia.com/role=interSwitch"
}

# Get a single interface by name
data "interfaces-v1alpha1_interface" "leaf1_ethernet_1_1" {
  namespace = "eda"
  name      = "leaf1-ethernet-1-1"
}

