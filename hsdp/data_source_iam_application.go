package hsdp

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/philips-software/go-hsdp-api/iam"
)

func dataSourceIAMApplication() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIAMApplicationRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"proposition_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"global_reference_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}

}

func dataSourceIAMApplicationRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	var diags diag.Diagnostics

	client, err := config.IAMClient()
	if err != nil {
		return diag.FromErr(err)
	}
	propID := d.Get("proposition_id").(string)
	name := d.Get("name").(string)

	apps, _, err := client.Applications.GetApplications(&iam.GetApplicationsOptions{
		PropositionID: &propID,
		Name:          &name,
	})
	if err != nil {
		return diag.FromErr(err)
	}
	if len(apps) == 0 {
		return diag.FromErr(ErrResourceNotFound)
	}

	d.SetId(apps[0].ID)
	_ = d.Set("description", apps[0].Description)
	_ = d.Set("global_reference_id", apps[0].GlobalReferenceID)
	return diags
}
