# hsdp_stl_device
Represents a STL device

## Example usage

```hcl
data "hsdp_stl_device" "sme100" {
  serial_number = "S4439394855830303"
}

output "sme100_status" {
  value = data.hsdp_stl_device.sme100.status
}
```


## Argument reference
* `serial_number` - (Required) the serial number of the device

## Attribute reference
* `id` - The device ID
* `region` - The region to which this device is connected to
* `state` - State of the device
* `primary_interface_ip` - The IP of the primary interface