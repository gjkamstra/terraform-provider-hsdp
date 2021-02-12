package hsdp

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/philips-software/go-hsdp-api/pki"
	"strings"
)

func resourcePKICert() *schema.Resource {
	return &schema.Resource{
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CreateContext: resourcePKICertCreate,
		ReadContext:   resourcePKICertRead,
		DeleteContext: resourcePKICertDelete,

		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"role": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"common_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"alt_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"ip_sans": {
				Type:     schema.TypeSet,
				ForceNew: true,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"uri_sans": {
				Type:     schema.TypeSet,
				ForceNew: true,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"other_sans": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"ttl": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"exclude_cn_from_sans": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"cert_pem": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ca_chain_pem": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_key_pem": {
				Type:      schema.TypeString,
				Sensitive: true,
				Computed:  true,
			},
			"expiration": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"issuing_ca_pem": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"serial_number": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourcePKICertCreate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(*Config)
	var err error
	var client *pki.Client

	client, err = config.PKIClient()
	if err != nil {
		return diag.FromErr(err)
	}
	defer client.Close()

	tenantID := d.Get("tenant_id").(string)
	logicalPath, err := pki.APIEndpoint(tenantID).LogicalPath()
	if err != nil {
		return diag.FromErr(fmt.Errorf("create PKI cert logicalPath: %w", err))
	}
	tenant, _, err := client.Tenants.Retrieve(logicalPath)
	if err != nil {
		return diag.FromErr(err)
	}
	roleName := d.Get("role").(string)
	ttl := d.Get("ttl").(string)
	ipSANS := expandStringList(d.Get("ip_sans").(*schema.Set).List())
	uriSANS := expandStringList(d.Get("uri_sans").(*schema.Set).List())
	otherSANS := expandStringList(d.Get("other_sans").(*schema.Set).List())
	commonName := d.Get("common_name").(string)
	altName := d.Get("alt_name").(string)
	excludeCNFromSANS := d.Get("exclude_cn_from_sans").(bool)
	role, ok := tenant.GetRoleOk(roleName)
	if !ok {
		return diag.FromErr(fmt.Errorf("role '%s' not found or invalid", roleName))
	}
	certRequest := pki.CertificateRequest{
		CommonName:        commonName,
		AltName:           altName,
		IPSANS:            strings.Join(ipSANS, ","),
		URISANS:           strings.Join(uriSANS, ","),
		OtherSANS:         strings.Join(otherSANS, ","),
		TTL:               ttl,
		ExcludeCNFromSANS: &excludeCNFromSANS,
		PrivateKeyFormat:  "pem",
		Format:            "pem",
	}
	cert, _, err := client.Services.IssueCertificate(logicalPath, role.Name, certRequest)
	if err != nil {
		return diag.FromErr(fmt.Errorf("issue PKI cert: %w", err))
	}
	d.SetId(cert.Data.SerialNumber)
	err = certToSchema(cert, d, m)
	if err != nil {
		d.SetId("")
		return diag.FromErr(err)
	}
	return diags
}

func certToSchema(cert *pki.IssueResponse, d *schema.ResourceData, _ interface{}) error {
	if cert.Data.PrivateKey != "" {
		_ = d.Set("private_key_pem", cert.Data.PrivateKey)
	}
	_ = d.Set("serial_number", cert.Data.SerialNumber)
	_ = d.Set("issuing_ca_pem", cert.Data.IssuingCa)
	_ = d.Set("cert_pem", cert.Data.Certificate)
	_ = d.Set("expiration", cert.Data.Expiration)
	_ = d.Set("ca_chain_pem", strings.Join(cert.Data.CaChain, "\n"))
	return nil
}

func resourcePKICertRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(*Config)
	var err error
	var client *pki.Client

	client, err = config.PKIClient()
	if err != nil {
		return diag.FromErr(err)
	}
	defer client.Close()

	tenantID := d.Get("tenant_id").(string)
	logicalPath, err := pki.APIEndpoint(tenantID).LogicalPath()
	if err != nil {
		return diag.FromErr(fmt.Errorf("read PKI cert logicalPath: %w", err))
	}
	cert, _, err := client.Services.GetCertificateBySerial(logicalPath, d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("read PKI cert: %w", err))
	}
	err = certToSchema(cert, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourcePKICertDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(*Config)
	var err error
	var client *pki.Client

	client, err = config.PKIClient()
	if err != nil {
		return diag.FromErr(err)
	}
	defer client.Close()

	tenantID := d.Get("tenant_id").(string)
	logicalPath, err := pki.APIEndpoint(tenantID).LogicalPath()
	if err != nil {
		return diag.FromErr(fmt.Errorf("delete PKI cert logicalPath: %w", err))
	}
	revoke, _, err := client.Services.RevokeCertificateBySerial(logicalPath, d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("delete PKI cert: %w", err))
	}
	if revoke.Data.RevocationTime > 0 {
		d.SetId("")
	}
	return diags
}
