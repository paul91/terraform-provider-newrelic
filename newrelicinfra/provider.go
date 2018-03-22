package newrelicinfra

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// Provider represents a resource provider in Terraform
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NEWRELIC_API_KEY", nil),
				Sensitive:   true,
			},
			"api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("NEWRELIC_INFRA_API_URL", "https://infra-api.newrelic.com/v2"),
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"newrelicinfra_alert_condition": resourceNewRelicInfraAlertCondition(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(data *schema.ResourceData) (interface{}, error) {
	config := Config{
		APIKey: data.Get("api_key").(string),
		APIURL: data.Get("api_url").(string),
	}
	log.Println("[INFO] Initializing New Relic Infra client")
	return config.Client()
}
