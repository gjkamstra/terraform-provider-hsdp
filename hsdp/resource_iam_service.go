package hsdp

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/philips-software/go-hsdp-api/iam"
	"net/http"
)

func resourceIAMService() *schema.Resource {
	return &schema.Resource{
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CreateContext: resourceIAMServiceCreate,
		ReadContext:   resourceIAMServiceRead,
		UpdateContext: resourceIAMServiceUpdate,
		DeleteContext: resourceIAMServiceDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: suppressCaseDiffs,
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
			},
			"application_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"validity": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  12,
				ForceNew: true,
				// TODO
				ValidateFunc: validation.IntBetween(1, 600),
			},
			"private_key": {
				Type:      schema.TypeString,
				Sensitive: true,
				Computed:  true,
			},
			"service_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"organization_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expires_on": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"scopes": {
				Type:     schema.TypeSet,
				MaxItems: 100,
				MinItems: 1, // openid
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"default_scopes": {
				Type:     schema.TypeSet,
				MaxItems: 100,
				MinItems: 1, // openid
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceIAMServiceCreate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)

	var diags diag.Diagnostics

	client, err := config.IAMClient()
	if err != nil {
		return diag.FromErr(err)
	}

	var s iam.Service
	s.Description = d.Get("description").(string)
	s.Name = d.Get("name").(string)
	s.ApplicationID = d.Get("application_id").(string)
	s.Validity = d.Get("validity").(int)
	scopes := expandStringList(d.Get("scopes").(*schema.Set).List())
	defaultScopes := expandStringList(d.Get("default_scopes").(*schema.Set).List())

	createdService, _, err := client.Services.CreateService(s)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(createdService.ID)
	_ = d.Set("expires_on", createdService.ExpiresOn)
	_ = d.Set("scopes", createdService.Scopes)
	_ = d.Set("default_scopes", createdService.DefaultScopes)
	_ = d.Set("private_key", createdService.PrivateKey)
	_ = d.Set("service_id", createdService.ServiceID)
	_ = d.Set("organization_id", createdService.OrganizationID)
	_ = d.Set("description", s.Description) // RITM0021326

	// Set scopes and default_scopes
	_, _, err = client.Services.AddScopes(*createdService, scopes, defaultScopes)
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	return diags
}

func resourceIAMServiceRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)

	var diags diag.Diagnostics

	client, err := config.IAMClient()
	if err != nil {
		return diag.FromErr(err)
	}

	id := d.Id()
	s, resp, err := client.Services.GetServiceByID(id)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}
	// Until RITM0021326 is implemented, this will always clear the field
	// d.Set("description", s.Description)

	_ = d.Set("name", s.Name)
	_ = d.Set("application_id", s.ApplicationID)
	_ = d.Set("expires_on", s.ExpiresOn)
	_ = d.Set("organization_id", s.OrganizationID)
	_ = d.Set("service_id", s.ServiceID)
	_ = d.Set("scopes", s.Scopes)
	_ = d.Set("expires_on", s.ExpiresOn)
	_ = d.Set("default_scopes", s.DefaultScopes)
	// The private key is only returned on create
	return diags
}

func resourceIAMServiceUpdate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)

	var diags diag.Diagnostics

	client, err := config.IAMClient()
	if err != nil {
		return diag.FromErr(err)
	}

	var s iam.Service
	s.ID = d.Id()

	if d.HasChange("scopes") {
		o, n := d.GetChange("scopes")
		old := expandStringList(o.(*schema.Set).List())
		newList := expandStringList(n.(*schema.Set).List())
		toAdd := difference(newList, old)
		toRemove := difference(old, newList)
		if len(toRemove) > 0 {
			_, _, err := client.Services.RemoveScopes(s, toRemove, []string{})
			if err != nil {
				return diag.FromErr(err)
			}
		}
		if len(toAdd) > 0 {
			_, _, _ = client.Services.AddScopes(s, toAdd, []string{})
		}
	}
	if d.HasChange("default_scopes") {
		o, n := d.GetChange("default_scopes")
		old := expandStringList(o.(*schema.Set).List())
		newList := expandStringList(n.(*schema.Set).List())
		toAdd := difference(newList, old)
		toRemove := difference(old, newList)
		if len(toRemove) > 0 {
			_, _, err := client.Services.RemoveScopes(s, []string{}, toRemove)
			if err != nil {
				return diag.FromErr(err)
			}
		}
		if len(toAdd) > 0 {
			_, _, _ = client.Services.AddScopes(s, []string{}, toAdd)
		}
	}
	return diags
}

func resourceIAMServiceDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)

	var diags diag.Diagnostics

	client, err := config.IAMClient()
	if err != nil {
		return diag.FromErr(err)
	}

	var s iam.Service
	s.ID = d.Id()
	ok, _, err := client.Services.DeleteService(s)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ok {
		return diag.FromErr(ErrDeleteServiceFailed)
	}
	d.SetId("")
	return diags
}
