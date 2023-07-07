package json

import (
	"context"
	"fmt"
	"github.com/Azure/terraform-provider-azapi/utils"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ basetypes.StringValuable = JsonStringValue{}

type JsonStringValue struct {
	basetypes.StringValue
}

func (v JsonStringValue) Equal(o attr.Value) bool {
	other, ok := o.(JsonStringValue)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (v JsonStringValue) Type(ctx context.Context) attr.Type {
	return JsonStringType{}
}

var _ basetypes.StringValuableWithSemanticEquals = JsonStringValue{}

func (v JsonStringValue) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	// The framework should always pass the correct value type, but always check
	newValue, ok := newValuable.(JsonStringValue)

	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", newValuable),
		)

		return false, diags
	}

	priorJson := utils.NormalizeJson(v.ValueString())

	newJson := utils.NormalizeJson(newValue.ValueString())

	return priorJson == newJson, diags
}
