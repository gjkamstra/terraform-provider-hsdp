package main

import (
	"aemian.com/terraform-provider-hsdpiam/hsdpiam"
	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/terraform"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() terraform.ResourceProvider {
			return hsdpiam.Provider()
		},
	})
}