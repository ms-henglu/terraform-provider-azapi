package location

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ basetypes.StringValuable = LocationStringValue{}

type LocationStringValue struct {
	basetypes.StringValue
}

func (v LocationStringValue) Equal(o attr.Value) bool {
	other, ok := o.(LocationStringValue)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (v LocationStringValue) Type(ctx context.Context) attr.Type {
	return LocationStringType{}
}

var _ basetypes.StringValuableWithSemanticEquals = LocationStringValue{}

func (v LocationStringValue) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	// The framework should always pass the correct value type, but always check
	newValue, ok := newValuable.(LocationStringValue)

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

	priorLocation := Normalize(v.ValueString())

	newLocation := Normalize(newValue.ValueString())

	return priorLocation == newLocation, diags
}
