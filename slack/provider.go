package slack

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type Config struct {
	Token string
}

func Provider() *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SLACK_TOKEN", nil),
				Description: "Authentication token (Requires scope: 'client')",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"slack_user": resourceUser(),
			// "slack_user_group": resourceUserGroup(),
		},
	}

	provider.ConfigureFunc = providerConfigure(provider)

	return provider
}

func providerConfigure(p *schema.Provider) schema.ConfigureFunc {
	return func(d *schema.ResourceData) (interface{}, error) {
		config := Config{
			Token: d.Get("token").(string),
		}

		return &config, nil
	}
}
