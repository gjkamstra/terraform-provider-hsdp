package hsdp

import (
	"context"
	"encoding/json"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	creds "github.com/philips-software/go-hsdp-api/s3creds"
)

func resourceS3CredsPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceS3CredsPolicyCreate,
		ReadContext:   resourceS3CredsPolicyRead,
		DeleteContext: resourceS3CredsPolicyDelete,

		Schema: map[string]*schema.Schema{
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateFunc:     validatePolicyJSON,
				DiffSuppressFunc: suppressEquivalentPolicyDiffs,
			},
			"product_key": {
				Type:      schema.TypeString,
				Sensitive: true,
				ForceNew:  true,
				Required:  true,
			},
		},
	}
}

func resourceS3CredsPolicyCreate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)

	var diags diag.Diagnostics

	client, err := config.S3CredsClient()
	if err != nil {
		return diag.FromErr(err)
	}

	productKey := d.Get("product_key").(string)
	policyJSON := d.Get("policy").(string)
	var policy creds.Policy

	err = json.Unmarshal([]byte(policyJSON), &policy)
	if err != nil {
		return diag.FromErr(err)
	}
	policy.ProductKey = productKey

	createdPolicy, _, err := client.Policy.CreatePolicy(policy)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(strconv.Itoa(createdPolicy.ID))
	return diags
}

func resourceS3CredsPolicyRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)

	var diags diag.Diagnostics

	client, err := config.S3CredsClient()
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	productKey := d.Get("product_key").(string)

	policies, _, err := client.Policy.GetPolicy(&creds.GetPolicyOptions{
		ID:         &id,
		ProductKey: &productKey,
	})
	if err != nil {
		return diag.FromErr(err)
	}
	if len(policies) != 1 { // Policy was deleted
		d.SetId("")
		return diags
	}
	policy := policies[0]

	d.SetId(strconv.Itoa(policy.ID))
	policy.ID = 0 // Don't marshal ID
	policyJSON, err := json.Marshal(policy)
	if err != nil {
		d.SetId("")
		return diag.FromErr(err)

	}
	_ = d.Set("policy", policyJSON)
	_ = d.Set("product_key", productKey)
	return diags
}

func resourceS3CredsPolicyDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)

	var diags diag.Diagnostics

	client, err := config.S3CredsClient()
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	productKey := d.Get("product_key").(string)
	policy := creds.Policy{
		ID:         id,
		ProductKey: productKey,
	}
	ok, _, err := client.Policy.DeletePolicy(policy)
	if err != nil {
		return diag.FromErr(err)
	}
	if ok {
		d.SetId("")
	}
	return diags
}
