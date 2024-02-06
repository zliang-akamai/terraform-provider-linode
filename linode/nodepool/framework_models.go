package nodepool

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/linode/linodego"
	"github.com/linode/terraform-provider-linode/v2/linode/helper"
)

type NodePoolModel struct {
	ID         types.String             `tfsdk:"id"`
	PoolID     types.Int64              `tfsdk:"pool_id"`
	ClusterID  types.Int64              `tfsdk:"cluster_id"`
	Count      types.Int64              `tfsdk:"node_count"`
	Type       types.String             `tfsdk:"type"`
	Tags       types.List               `tfsdk:"tags"`
	Nodes      types.List               `tfsdk:"nodes"`
	Autoscaler *NodePoolAutoscalerModel `tfsdk:"autoscaler"`
}

type NodePoolAutoscalerModel struct {
	Min types.Int64 `tfsdk:"min"`
	Max types.Int64 `tfsdk:"max"`
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
	pool.ID = types.StringValue(fmt.Sprintf("%d:%d", clusterID, p.ID))
	pool.PoolID = types.Int64Value(int64(p.ID))
	pool.ClusterID = types.Int64Value(int64(clusterID))
	pool.Count = types.Int64Value(int64(p.Count))
	pool.Type = types.StringValue(p.Type)
	tags, d := types.ListValueFrom(ctx, types.StringType, p.Tags)
	if d != nil {
		diags.Append(d...)
		return
	}
	pool.Tags = tags

	if p.Autoscaler.Enabled {
		var autoscaler NodePoolAutoscalerModel
		autoscaler.Min = types.Int64Value(int64(p.Autoscaler.Min))
		autoscaler.Max = types.Int64Value(int64(p.Autoscaler.Max))
		pool.Autoscaler = &autoscaler
	} else {
		pool.Autoscaler = nil
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
}

func (pool *NodePoolModel) ExtractClusterAndNodePoolIDs(diags *diag.Diagnostics) (int, int) {
	clusterID := helper.FrameworkSafeInt64ToInt(pool.ClusterID.ValueInt64(), diags)
	poolID := helper.FrameworkSafeInt64ToInt(pool.PoolID.ValueInt64(), diags)
	return clusterID, poolID
}

func (pool *NodePoolModel) getLKENodePoolAutoscaler(count int, diags *diag.Diagnostics) *linodego.LKENodePoolAutoscaler {
	var autoscaler linodego.LKENodePoolAutoscaler
	if pool.Autoscaler != nil {
		autoscaler.Enabled = true
		autoscaler.Min = helper.FrameworkSafeInt64ToInt(pool.Autoscaler.Min.ValueInt64(), diags)
		autoscaler.Max = helper.FrameworkSafeInt64ToInt(pool.Autoscaler.Max.ValueInt64(), diags)
	} else {
		autoscaler.Enabled = false
		autoscaler.Min = count
		autoscaler.Max = count
	}
	return &autoscaler
}
