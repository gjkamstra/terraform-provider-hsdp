package hsdp

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/philips-software/go-hsdp-api/iam"
)

func dataSourceService() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceServiceRead,
		Schema: map[string]*schema.Schema{
			"service_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"organization_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"application_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expires_on": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_scopes": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"scopes": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}

}

func dataSourceServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	var diags diag.Diagnostics

	client, err := config.IAMClient()
	if err != nil {
		return diag.FromErr(err)
	}

	serviceID := d.Get("service_id").(string)

	service, _, err := client.Services.GetService(&iam.GetServiceOptions{
		ServiceID: &serviceID,
	})

	if err != nil {
		return diag.FromErr(err)
	}
	_ = d.Set("application_id", service.ApplicationID)
	_ = d.Set("expires_on", service.ExpiresOn)
	_ = d.Set("organization_id", service.OrganizationID)
	_ = d.Set("description", service.Description)
	_ = d.Set("name", service.Name)
	_ = d.Set("default_scopes", service.DefaultScopes)
	_ = d.Set("scopes", service.Scopes)
	_ = d.Set("uuid", service.ID)
	d.SetId(service.ID)
	return diags
}
