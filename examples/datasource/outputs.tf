output "all_interfaces" {
  value = data.interfaces-v1alpha1_interface_list.all
}

output "all_interswitch_interfaces" {
  value = data.interfaces-v1alpha1_interface_list.interswitch
}

output "leaf1_ethernet_1_1" {
  value = data.interfaces-v1alpha1_interface.leaf1_ethernet_1_1
}

# output "myif_2" {
#   value = interfaces_interface.myif_2
# }
