package nodepool

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/linode/linodego"
	"github.com/linode/terraform-provider-linode/v2/linode/helper"
)

type NodePoolModel struct {
	ID         types.String              `tfsdk:"id"`
	ClusterID  types.Int64               `tfsdk:"cluster_id"`
	Count      types.Int64               `tfsdk:"node_count"`
	Type       types.String              `tfsdk:"type"`
	Tags       types.Set                 `tfsdk:"tags"`
	Nodes      types.List                `tfsdk:"nodes"`
	Autoscaler []NodePoolAutoscalerModel `tfsdk:"autoscaler"`
}

type NodePoolAutoscalerModel struct {
	Min types.Int64 `tfsdk:"min"`
	Max types.Int64 `tfsdk:"max"`
}

type NodePoolNodeModel struct {
	ID         types.String `tfsdk:"id"`
	InstanceID types.Int64  `tfsdk:"instance_id"`
	Status     types.String `tfsdk:"status"`
}

func flattenLKENodePoolLinode(node *linodego.LKENodePoolLinode) (*basetypes.ObjectValue, diag.Diagnostics) {
	result := make(map[string]attr.Value)

	result["id"] = types.StringValue(node.ID)
	result["instance_id"] = types.Int64Value(int64(node.InstanceID))
	result["status"] = types.StringValue(string(node.Status))

	obj, errors := types.ObjectValue(nodeObjectType.AttrTypes, result)
	if errors.HasError() {
		return nil, errors
	}
	return &obj, nil
}

func parseNodeList(nodes []linodego.LKENodePoolLinode,
) (*basetypes.ListValue, diag.Diagnostics) {
	resultList := make([]attr.Value, len(nodes))
	for i, node := range nodes {
		result, errors := flattenLKENodePoolLinode(&node)
		if errors.HasError() {
			return nil, errors
		}

		resultList[i] = result
	}
	result, errors := basetypes.NewListValue(
		nodeObjectType,
		resultList,
	)
	if errors.HasError() {
		return nil, errors
	}

	return &result, nil
}

func (pool *NodePoolModel) ParseNodePool(ctx context.Context, clusterID int, p *linodego.LKENodePool, diags *diag.Diagnostics) {
	pool.ID = types.StringValue(strconv.Itoa(p.ID))
	pool.ClusterID = types.Int64Value(int64(clusterID))
	pool.Count = types.Int64Value(int64(p.Count))
	pool.Type = types.StringValue(p.Type)
	tags, d := types.SetValueFrom(ctx, types.StringType, p.Tags)
	if d != nil {
		diags.Append(d...)
		return
	}
	pool.Tags = tags

	if p.Autoscaler.Enabled {
		pool.Autoscaler = []NodePoolAutoscalerModel{
			{
				Min: types.Int64Value(int64(p.Autoscaler.Min)),
				Max: types.Int64Value(int64(p.Autoscaler.Max)),
			},
		}
	}

	nodes, errs := parseNodeList(p.Linodes)
	if errs != nil {
		diags.Append(errs...)
	}
	pool.Nodes = *nodes
}

func (pool *NodePoolModel) SetNodePoolCreateOptions(ctx context.Context, p *linodego.LKENodePoolCreateOptions, diags *diag.Diagnostics) {
	p.Count = helper.FrameworkSafeInt64ToInt(
		pool.Count.ValueInt64(),
		diags,
	)
	p.Type = pool.Type.ValueString()

	if !pool.Tags.IsNull() {
		diags.Append(pool.Tags.ElementsAs(ctx, &p.Tags, false)...)
	}

	p.Autoscaler = pool.getLKENodePoolAutoscaler(p.Count, diags)
	if p.Autoscaler.Enabled && p.Count == 0 {
		p.Count = p.Autoscaler.Min
	}
}

func (pool *NodePoolModel) SetNodePoolUpdateOptions(ctx context.Context, p *linodego.LKENodePoolUpdateOptions, diags *diag.Diagnostics) {
	p.Count = helper.FrameworkSafeInt64ToInt(
		pool.Count.ValueInt64(),
		diags,
	)
	if diags.HasError() {
		return
	}

	if !pool.Tags.IsNull() {
		diags.Append(pool.Tags.ElementsAs(ctx, &p.Tags, false)...)
		if diags.HasError() {
			return
		}
	}

	p.Autoscaler = pool.getLKENodePoolAutoscaler(p.Count, diags)
	if p.Autoscaler.Enabled && p.Count == 0 {
		p.Count = p.Autoscaler.Min
	}
}

func (pool *NodePoolModel) ExtractClusterAndNodePoolIDs(diags *diag.Diagnostics) (int, int) {
	clusterID := helper.FrameworkSafeInt64ToInt(pool.ClusterID.ValueInt64(), diags)
	poolID, err := strconv.Atoi(pool.ID.ValueString())
	if err != nil {
		diags.AddError("Failed to parse poolID", err.Error())
	}
	return clusterID, poolID
}

func (pool *NodePoolModel) getLKENodePoolAutoscaler(count int, diags *diag.Diagnostics) *linodego.LKENodePoolAutoscaler {
	var autoscaler linodego.LKENodePoolAutoscaler
	if len(pool.Autoscaler) > 0 {
		autoscaler.Enabled = true
		autoscaler.Min = helper.FrameworkSafeInt64ToInt(pool.Autoscaler[0].Min.ValueInt64(), diags)
		autoscaler.Max = helper.FrameworkSafeInt64ToInt(pool.Autoscaler[0].Max.ValueInt64(), diags)
	} else {
		autoscaler.Enabled = false
		autoscaler.Min = count
		autoscaler.Max = count
	}
	return &autoscaler
}
