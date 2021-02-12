# HSDP Provider

The HSDP provider can provision and manage a select set of HSDP resources. This includes amongst others many IAM entities, Container Host instances and even some Clinical Data Repository (CDR) resources.

## Configuring the provider

```hcl
provider "hsdp" {
  region             = "us-east"
  environment        = "client-test"
  oauth2_client_id   = var.oauth2_client_id
  oauth2_password    = var.oauth2_password
  org_admin_username = var.org_admin_username
  org_admin_password = var.org_admin_password
}
```

## Argument Reference

The following arguments are supported:

* `region` - (Optional) The HSDP region to use [us-east, eu-west, sa1, ...]

* `environment` - (Optional) The HSDP environment to use within region [client-test, prod]

* `iam_url` - (Optional) IAM API endpoint (e.g. https://iam-client-test.us-east.philips-healthsuite.com). Auto-discovered when region and environment are specified.

* `idm_url` - (Optioanl) IDM API endpoint (e.g. https://idm-client-test.us-east.philips-healthsuite.com). Auto-discovered when region and environment are specified.

* `s3creds_url` - (Optional) S3 Credenials API endpoint (e.g. https://s3creds-client-test.us-east.philips-healthsuite.com). Auto-discovered when region and environment are specified.

* `oauth2_client_id` - (Required) The OAuth2 client ID as provided by HSDP

* `oauth2_password` - (Required) The OAuth2 password as provided by HSDP

* `service_id` - (Optional) The service ID to use for IAM org admin operations (conflicts with: `org_admin_username`)

* `service_private_key` - (Optional) The service private key to use for IAM org admin operations (conflicts with: `org_admin_password`)

* `org_admin_username` - (Optional) Your IAM admin username.

* `org_admin_password` - (Optional) Your IAM admin password.

* `uaa_username` - (Optional) The HSDP CF UAA username.

* `uaa_password` - (Optional) The HSDP CF UAA password.

* `uaa_url` - (Optional) The URL of the UAA authentication service

* `org_id` - **Deprecated** Your IAM root ORG id as provided by HSDP

* `shared_key` - (Optional) The shared key as provided by HSDP. Actions which require API signing will not work if this value is missing.

* `secret_key` - (Optional) The secret key as provided by HSDP. Actions which require API signing will not work if this value is missing.

* `cartel_host` - (Optional) The cartel host as provided by HSDP. Auto-discovered when region and environment are specified.

* `cartel_token` - (Optional) The cartel token as provided by HSDP.

* `cartel_secret` - (Optional) The cartel secret as provided by HSDP.

* `retry_max` - (Optional) Integer, when > 0 will use a retry-able HTTP client and retry requests when applicable.

* `debug` - **deprecated** If set to true, outputs details on API calls. Deprecated, just setting `debug_log` is sufficient.

* `debug_log` - (Optional) If set to a path, when debug is enabled outputs details to this file

