package customtypes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ = basetypes.StringTypable(&LocationType{})
	_ = fmt.Stringer(&LocationType{})
	_ = basetypes.StringValuableWithSemanticEquals(&LocationValue{})
)

type LocationType struct {
	basetypes.StringType
}

func (t LocationType) Equal(o attr.Type) bool {
	other, ok := o.(LocationType)
	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (t LocationType) String() string {
	return "LocationType"
}

func (t LocationType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	value := LocationValue{
		StringValue: in,
	}

	return value, nil
}

func (t LocationType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, fmt.Errorf("unexpected error converting value from Terraform: %w", err)
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

func (t LocationType) ValueType(_ context.Context) attr.Value {
	return LocationValue{}
}
