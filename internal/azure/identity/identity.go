package identity

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

func SchemaIdentity() *schema.SingleNestedAttribute {
	return &schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Required: true,
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
				Optional: true,
				ElementType: types.ListType{
					ElemType: types.StringType,
				},
				Validators: []validator.List{},
				// TODO@ms-henglu： validate each element in this list
			},

			"principal_id": schema.StringAttribute{
				Computed: true,
			},

			"tenant_id": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func ExpandIdentity(input []interface{}) (interface{}, error) {
	if len(input) == 0 || input[0] == nil {
		return nil, nil
	}

	v := input[0].(map[string]interface{})

	config := map[string]interface{}{}
	identityType := IdentityType(v["type"].(string))
	config["type"] = identityType
	identityIds := v["identity_ids"].([]interface{})
	userAssignedIdentities := make(map[string]interface{}, len(identityIds))
	if len(identityIds) != 0 {
		if identityType != UserAssigned && identityType != SystemAssignedUserAssigned {
			return nil, fmt.Errorf("`identity_ids` can only be specified when `type` includes `UserAssigned`")
		}
		for _, id := range identityIds {
			userAssignedIdentities[id.(string)] = make(map[string]interface{})
		}
		config["userAssignedIdentities"] = userAssignedIdentities
	}
	return config, nil
}

func ExpandIdentity2(raw types.Object) (interface{}, error) {
	input := make(map[string]interface{})
	diags := raw.As(context.TODO(), &input, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, fmt.Errorf("%+v", diags)
	}

	config := map[string]interface{}{}
	identityType := IdentityType(input["type"].(string))
	config["type"] = identityType
	identityIds := input["identity_ids"].([]interface{})
	userAssignedIdentities := make(map[string]interface{}, len(identityIds))
	if len(identityIds) != 0 {
		if identityType != UserAssigned && identityType != SystemAssignedUserAssigned {
			return nil, fmt.Errorf("`identity_ids` can only be specified when `type` includes `UserAssigned`")
		}
		for _, id := range identityIds {
			userAssignedIdentities[id.(string)] = make(map[string]interface{})
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

func FlattenIdentity(identity interface{}) map[string]attr.Value {
	if identity == nil {
		return nil
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

		return map[string]attr.Value{
			"type":         types.StringValue(identityType),
			"identity_ids": types.ListValueMust(types.StringType, identityIds),
			"principal_id": types.StringValue(identityMap["principalId"].(string)),
			"tenant_id":    types.StringValue(identityMap["tenantId"].(string)),
		}
	}
	return nil
}
