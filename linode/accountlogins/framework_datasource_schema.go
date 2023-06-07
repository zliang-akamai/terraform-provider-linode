package accountlogins

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/linode/terraform-provider-linode/linode/accountlogin"
	"github.com/linode/terraform-provider-linode/linode/helper/frameworkfilter"
)

var filterConfig = frameworkfilter.Config{
	"ip":         {APIFilterable: false},
	"restricted": {APIFilterable: false},
	"username":   {APIFilterable: false},
	"status":     {APIFilterable: false},
}

var accountLoginSchema = schema.NestedBlockObject{
	Attributes: accountlogin.Attributes,
}

var frameworkDataSourceSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "The data source's unique ID.",
			Computed:    true,
		},
	},
	Blocks: map[string]schema.Block{
		"filter": filterConfig.Schema(),
		"logins": schema.ListNestedBlock{
			Description:  "The returned list of account logins.",
			NestedObject: accountLoginSchema,
		},
	},
}