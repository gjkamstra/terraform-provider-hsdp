# hsdp_dicom_object_store
This resource manages DICOM Object stores

# Example Usage

```hcl
resource "hsdp_dicom_object_store" "store1" {
  config_url = var.dicom_base_url
  organization_id = var.iam_org_one_id
  description = "Store 1"

  static_access {
    endpoint = "https://s3-external.amazonaws.com"
    bucket_name = "xxxx-xxxx-xxxx-xxxx"
    access_key = "xxx"
    secret_key = "yyy"
  }
}

resource "hsdp_dicom_object_store" "store2" {
  config_url = var.dicom_base_url
  organization_id = var.iam_org_two_id
  description = "Store 2"
  
  s3creds_access {
    endpoint = "https://xxx.com"
    product_key = "xxxx-xxxx-xxxx-xxxx"
    bucket_name = "yyyy-yyyy-yyy-yyyy"
    folder_path = "/store1"
    service_account {
      service_id = "a@b.com"
      private_key = var.service_private_key
    }
  }
}
```

# Argument reference

* `config_url` - (Required) The base config URL of the DICOM Object store instance
* `organization_id` - (Required) the IAM organization ID to use for authorization
* `description` - (Optional) Description of the object store
* `static_access` - (Optional) Details of the CDR service account
    * `endpoint` - (Required) The S3 bucket endpoint
    * `bucket_name` - (Required) The S3 bucket name
    * `access_key` - (Required) The S3 access key
    * `secret_key` - (Required) The S3 secret key
* `s3creds_access` - (Optional) the FHIR store configuration
    * `endpoint` - (Required) The S3Creds bucket endpoint
    * `product_key` - (Required) The S3Creds product key  
    * `bucket_name` - (Required) The S3Creds bucket name
    * `folder_path` - (Required) The S3Creds folder path to use
    * `service_account` - (Required) The IAM service account to use
      * `service_id` - (Required) The IAM service id
      * `private_key` - (Required) The IAM service private key
      * `name` - (Optional) Name of the service
      * `access_token_endpoint` - (Optional) The IAM access token endpoint
      * `token_endpoint` - (Optional) The IAM token endpoint

# Attribute reference
* `access_type` - The access type for this object store

