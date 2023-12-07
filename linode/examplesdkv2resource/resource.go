package examplesdkv2resource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var resourceSchema = map[string]*schema.Schema{
	"example_block_attr": {
		Type: schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"example_nested_attr": {
					Type:     schema.TypeString,
					Computed: true,
					Optional: true,
				},
				"example_nested_attr_with_default": {
					Type:     schema.TypeString,
					Default:  "default",
					Optional: true,
				},
			},
		},
		Computed: true,
		Optional: true,
	},
}

func Resource() *schema.Resource {
	return &schema.Resource{
		Schema:        resourceSchema,
		ReadContext:   readResource,
		CreateContext: createResource,
		UpdateContext: updateResource,
		DeleteContext: deleteResource,
	}
}

func putResource(ctx context.Context, d *schema.ResourceData) {
	existingBlock := d.Get("example_block_attr").([]any)
	if len(existingBlock) == 0 {
		newBlock := make([]map[string]interface{}, 1)

		newBlock[0] = map[string]interface{}{
			"example_nested_attr":              "test value",
			"example_nested_attr_with_default": "test value",
		}

		d.Set("example_block_attr", newBlock)

		tflog.Info(ctx, "entering create new block")

	} else {
		d.Set("example_block_attr", d.Get("example_block_attr"))
	}
}

func createResource(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	d.SetId("1234")
	putResource(ctx, d)
	return nil
}

func readResource(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	d.Set("example_block_attr", d.Get("example_block_attr"))
	return nil
}

func updateResource(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	putResource(ctx, d)
	return readResource(ctx, d, meta)
}

func deleteResource(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	d.SetId("")
	return nil
}
