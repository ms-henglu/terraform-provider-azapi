package location

import (
	"context"
	"fmt"
	
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ basetypes.StringTypable = LocationStringType{}

type LocationStringType struct {
	basetypes.StringType
}

func (t LocationStringType) Equal(o attr.Type) bool {
	other, ok := o.(LocationStringType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (t LocationStringType) String() string {
	return "LocationStringType"
}

func (t LocationStringType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	value := LocationStringValue{
		StringValue: in,
	}

	return value, nil
}

func (t LocationStringType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)

	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)

	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}

func (t LocationStringType) ValueType(ctx context.Context) attr.Value {
	return LocationStringValue{}
}
