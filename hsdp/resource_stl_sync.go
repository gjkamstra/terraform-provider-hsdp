package hsdp

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/philips-software/go-hsdp-api/stl"
)

func resourceSTLSync() *schema.Resource {
	return &schema.Resource{
		Description:   `The ` + "`hsdp_stl_sync`" + ` resource syncs device config to the actual device.`,
		CreateContext: resourceSTLSyncCreate,
		ReadContext:   resourceSTLSyncRead,
		DeleteContext: resourceSTLSyncDelete,

		Schema: map[string]*schema.Schema{
			"triggers": {
				Description: "A map of arbitrary strings that, when changed, will force the 'hsdp_stl_sync' resource to be replaced, re-sync conifg with the device.",
				Type:        schema.TypeMap,
				Required:    true,
				ForceNew:    true,
			},
			"serial_number": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSTLSyncDelete(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	d.SetId("")
	return diag.Diagnostics{}
}

func resourceSTLSyncRead(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return diag.Diagnostics{}
}

func resourceSTLSyncCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)
	var diags diag.Diagnostics
	var client *stl.Client
	var err error

	if endpoint, ok := d.GetOk("endpoint"); ok {
		client, err = config.STLClient(endpoint.(string))
	} else {
		client, err = config.STLClient()
	}
	if err != nil {
		return diag.FromErr(err)
	}
	serialNumber := d.Get("serial_number").(string)
	err = client.Devices.SyncDeviceConfig(ctx, serialNumber)
	if err != nil {
		return diag.FromErr(fmt.Errorf("hsdp_stl_sync: %w", err))
	}
	d.SetId(serialNumber)
	return diags
}
