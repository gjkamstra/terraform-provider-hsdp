# hsdp_iam_role
Provides a resource for managing HSDP IAM roles

## Example Usage

The following example creates a role

```hcl
resource "hsdp_iam_role" "TDRALL" {
  name        = "TDRALL"
  description = "Role for TDR users with ALL access"

  permissions = [
    "DATAITEM.CREATEONBEHALF",
    "DATAITEM.READ",
    "DATAITEM.DELETEONBEHALF",
    "DATAITEM.DELETE",
    "CONTRACT.CREATE",
    "DATAITEM.READONBEHALF",
    "CONTRACT.READ",
    "DATAITEM.CREATE",
  ]

  managing_organization = hsdp_iam_org.testdev.id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the group
* `description` - (Required) The description of the group
* `permissions` - (Required) The list of permission to assign to this role
* `managing_organization` - (Required) The managing organization ID of this role
* `ticket_protection` - (Optional) Defaults to true. Set to false to remove e.g. `CLIENT.SCOPES` permission which is only addable using a HSDP support ticket. 


## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the role

## Import

An existing role can be imported using `terraform import hsdp_iam_role`, e.g.

```shell
> terraform import hsdp_iam_role.myrole a-guid
```

