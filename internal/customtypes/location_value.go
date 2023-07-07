package customtypes

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ = basetypes.StringValuable(&LocationValue{})
	_ = basetypes.StringValuableWithSemanticEquals(&LocationValue{})
	_ = fmt.Stringer(&LocationValue{})
)

type LocationValue struct {
	basetypes.StringValue
}

func NewLocationNull() LocationValue {
	return LocationValue{
		StringValue: basetypes.NewStringNull(),
	}
}

func NewLocationUnknown() LocationValue {
	return LocationValue{
		StringValue: basetypes.NewStringUnknown(),
	}
}

func NewLocationValue(value string) LocationValue {
	return LocationValue{
		StringValue: basetypes.NewStringValue(value),
	}
}

func NewLocationPointerValue(value *string) LocationValue {
	if value == nil {
		return NewLocationNull()
	}

	return NewLocationValue(*value)
}

func (v LocationValue) Equal(o attr.Value) bool {
	other, ok := o.(LocationValue)
	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (v LocationValue) Type(_ context.Context) attr.Type {
	return LocationType{}
}

func (v LocationValue) String() string {
	return "LocationValue"
}

func (v LocationValue) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(LocationValue)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			fmt.Sprintf("Expected value type %T but got value type %T. Please report this to the provider developers.", v, newValuable),
		)

		return false, diags
	}

	priorLocation := NormalizeLocation(v.ValueString())

	newLocation := NormalizeLocation(newValue.ValueString())

	return priorLocation == newLocation, diags
}

func NormalizeLocation(input string) string {
	return strings.ReplaceAll(strings.ToLower(input), " ", "")
}
