package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/terraform-providers/terraform-provider-newrelic-infra/newrelicinfra"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: newrelicinfra.Provider})
}
