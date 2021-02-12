package hsdp

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/philips-software/go-hsdp-api/stl"
)

func dataSourceSTLDevice() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSTLDeviceRead,
		Schema: map[string]*schema.Schema{
			"endpoint": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"serial_number": {
				Type:     schema.TypeString,
				Required: true,
			},
			"environment": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hardware_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"primary_interface_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}

}

func dataSourceSTLDeviceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)
	var diags diag.Diagnostics
	var client *stl.Client
	var err error

	endpoint := d.Get("endpoint").(string)
	if endpoint != "" {
		client, err = config.STLClient(endpoint)
	} else {
		client, err = config.STLClient()
	}
	if err != nil {
		return diag.FromErr(err)
	}
	serialNumber := d.Get("serial_number").(string)
	device, err := client.Devices.GetDeviceBySerial(ctx, serialNumber)
	if err != nil {
		return diag.FromErr(fmt.Errorf("read STL device: %w", err))
	}
	_ = d.Set("region", device.Region)
	_ = d.Set("name", device.Name)
	_ = d.Set("state", device.State)
	_ = d.Set("primary_interface_ip", device.PrimaryInterface.Address)
	d.SetId(fmt.Sprintf("%d", device.ID))
	return diags
}
