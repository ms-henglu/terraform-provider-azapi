package identity

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"strings"

	"github.com/Azure/terraform-provider-azapi/internal/services/parse"
)

type IdentityType string

const (
	None                       IdentityType = "None"
	SystemAssigned             IdentityType = "SystemAssigned"
	UserAssigned               IdentityType = "UserAssigned"
	SystemAssignedUserAssigned IdentityType = "SystemAssigned, UserAssigned"
)

type Model struct {
	Type        types.String `tfsdk:"type"`
	IdentityIDs types.List   `tfsdk:"identity_ids"`
	PrincipalID types.String `tfsdk:"principal_id"`
	TenantID    types.String `tfsdk:"tenant_id"`
}

func SchemaIdentity() *schema.ListNestedBlock {
	return &schema.ListNestedBlock{
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"type": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.OneOf(
							string(None),
							string(UserAssigned),
							string(SystemAssigned),
							string(SystemAssignedUserAssigned),
						),
					},
				},

				"identity_ids": schema.ListAttribute{
					Optional:    true,
					ElementType: types.StringType,
					Validators:  []validator.List{},
					// TODO@ms-henglu： validate each element in this list
				},

				"principal_id": schema.StringAttribute{
					Computed: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},

				"tenant_id": schema.StringAttribute{
					Computed: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
			},
		},
	}
}

func ExpandIdentity(raw types.List) (interface{}, error) {
	if len(raw.Elements()) != 1 {
		return nil, nil
	}
	var input Model
	diags := raw.Elements()[0].(types.Object).As(context.TODO(), &input, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, fmt.Errorf("%+v", diags)
	}

	config := map[string]interface{}{}
	identityType := IdentityType(input.Type.ValueString())
	config["type"] = identityType
	identityIds := input.IdentityIDs.Elements()
	userAssignedIdentities := make(map[string]interface{}, len(identityIds))
	if len(identityIds) != 0 {
		if identityType != UserAssigned && identityType != SystemAssignedUserAssigned {
			return nil, fmt.Errorf("`identity_ids` can only be specified when `type` includes `UserAssigned`")
		}
		for _, id := range identityIds {
			userAssignedIdentities[id.(types.String).ValueString()] = make(map[string]interface{})
		}
		config["userAssignedIdentities"] = userAssignedIdentities
	}
	return config, nil
}

func AttributeTypes() map[string]attr.Type {
	out := make(map[string]attr.Type)
	out["type"] = types.StringType
	out["identity_ids"] = types.ListType{ElemType: types.StringType}
	out["principal_id"] = types.StringType
	out["tenant_id"] = types.StringType
	return out
}

func FlattenIdentity(identity interface{}) types.List {
	if identity == nil {
		return types.ListNull(types.ObjectType{AttrTypes: AttributeTypes()})
	}
	if identityMap, ok := identity.(map[string]interface{}); ok {
		identityIds := make([]attr.Value, 0)
		if identityMap["userAssignedIdentities"] != nil {
			userAssignedIdentities := identityMap["userAssignedIdentities"].(map[string]interface{})
			for key := range userAssignedIdentities {
				identityId, err := parse.UserAssignedIdentitiesID(key)
				if err == nil {
					identityIds = append(identityIds, types.StringValue(identityId.ID()))
				}
			}
		}

		identityType := identityMap["type"].(string)
		switch {
		case strings.Contains(identityType, ","):
			identityType = string(SystemAssignedUserAssigned)
		case strings.EqualFold(identityType, string(UserAssigned)):
			identityType = string(UserAssigned)
		case strings.EqualFold(identityType, string(SystemAssigned)):
			identityType = string(SystemAssigned)
		default:
			identityType = string(None)
		}

		out := map[string]attr.Value{
			"type":         types.StringValue(identityType),
			"identity_ids": types.ListValueMust(types.StringType, identityIds),
			"principal_id": types.StringNull(),
			"tenant_id":    types.StringNull(),
		}
		if principalId := identityMap["principalId"].(string); principalId != "" {
			out["principal_id"] = types.StringValue(principalId)
		}
		if tenantId := identityMap["tenantId"].(string); tenantId != "" {
			out["tenant_id"] = types.StringValue(tenantId)
		}
		return types.ListValueMust(types.ObjectType{AttrTypes: AttributeTypes()}, []attr.Value{types.ObjectValueMust(AttributeTypes(), out)})
	}
	return types.ListNull(types.ObjectType{AttrTypes: AttributeTypes()})
}
