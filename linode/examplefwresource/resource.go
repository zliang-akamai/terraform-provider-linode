package examplefwresource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/linode/terraform-provider-linode/v2/linode/helper"
)

var frameworkResourceSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
		},
	},
	Blocks: map[string]schema.Block{
		"example_block_attr": schema.SingleNestedBlock{
			Attributes: map[string]schema.Attribute{
				"example_nested_attr": schema.StringAttribute{
					Optional: true,
					Computed: true,
				},
				"example_nested_attr_with_default": schema.StringAttribute{
					Default:  stringdefault.StaticString("default"),
					Optional: true,
					Computed: true,
				},
			},
		},
	},
}

type Resource struct {
	helper.BaseResource
}

type ExampleBlockModel struct {
	ExampleNestedAttr            types.String `tfsdk:"example_nested_attr"`
	ExampleNestedAttrWithDefault types.String `tfsdk:"example_nested_attr_with_default"`
}

type ResourceModel struct {
	ID               types.String        `tfsdk:"id"`
	ExampleBlockAttr []ExampleBlockModel `tfsdk:"example_block_attr"`
}

func (rm *ResourceModel) ComputeValues() {
	if rm.ExampleBlockAttr == nil || len(rm.ExampleBlockAttr) == 0 {
		rm.ExampleBlockAttr = make([]ExampleBlockModel, 1)
		rm.ExampleBlockAttr[0] = ExampleBlockModel{
			ExampleNestedAttr:            types.StringValue("test value"),
			ExampleNestedAttrWithDefault: types.StringValue("test value"),
		}
	}
}

func NewResource() resource.Resource {
	return &Resource{
		BaseResource: helper.NewBaseResource(
			helper.BaseResourceConfig{
				Name:   "linode_example_fw_block",
				IDType: types.StringType,
				Schema: &frameworkResourceSchema,
			},
		),
	}
}

func (r *Resource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// var block []map[string]string
	// resp.Diagnostics.Append(data.ExampleBlockAttr.ElementsAs(ctx, &block, false)...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }

	data.ComputeValues()

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ComputeValues()

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *Resource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
}
