# hsdp_pki_tenant

Onboard tenant to PKI Service. Cloud foundry users with SpaceDeveloper role can onboard tenant

> This resource is only available when `uaa_*` (Cloud foundry) and `iam` credentials are set

## Example usage

```hcl
resource "hsdp_pki_tenant" "tenant" {
  organization_name = "client-my-org"
  space_name = "prod"
  
  iam_orgs = [
    var.iam_org_id
  ]
  
  ca {
    common_name = "Common Name Here"
  }
  
  role {
    name = "ec384"
    allow_any_name = true
    allow_ip_sans = true
    allow_subdomains = true
    allowed_domains = []
    allowed_other_sans = []
    allowed_uri_sans = []
    client_flag = true
    server_flag = true
    enforce_hostnames = false
    key_bits = 384
    key_type = "ec"
  }
}
```

## Argument reference 
The following arguments are supported:

* `organization_name` - (Required) The CF organization name to use
* `space_name` - (Required) The CF space name to verify the user is part of
* `role` - (Required) A role definition. Muliple roles are supported
* `ca` - (Required) The Certificate Authority information to use.
  * `common_name` - (Required) The common name to use
    
Each `role` definition takes the following arguments:
* `name` - (Required) The role name. This is used for lookup
* `key_type` - (Required) The key type. Values [`ec`, `rsa`]
* `key_bits` - (Required, int) Key length. Typically `384` for `ec` key types. 
* `client_flags` - (Required, bool) Allow use on clients
* `server_flags` - (Required, bool) Allow use on servers
* `allow_any_name` - (Required, bool) Allow any name
* `allow_ip_sans` - (Required, bool) Allow IP Subject Alternative Names (SAN)
* `allow_subdomains` - (Required, bool) Allow subdomains to be created
* `allow_any_name` - (Required, bool) Allow any name to be used
* `allowed_domains` - (Optional, list(string)) List of allowed domains
* `allowed_other_sans` - (Optional, list(string)) List of allowed other SANs
* `allowed_uri_sans` - (Optional, list(string)) List of allowed URI SANs
* `enforce_hostnames` - (Optional, bool) Enforce hostnames. Default: `false`

## Attribute reference
The following attributes are exported:

* `id` - The HSDP PKI `logical_path` of the tenant. The Terraform provider uses this as the Tenant ID
* `logical_path` - Same as `id`. This is for consistency.
* `private_key_pem` - The private key in PEM format