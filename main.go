package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/terraform-providers/terraform-provider-newrelicinfra/newrelicinfra"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: newrelicinfra.Provider})
}
