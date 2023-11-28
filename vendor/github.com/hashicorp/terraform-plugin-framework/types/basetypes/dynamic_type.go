// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package basetypes

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// DynamicTypable extends attr.Type for string types.
// Implement this interface to create a custom StringType type.
type DynamicTypable interface {
	attr.Type

	// ValueFromDynamic should convert the String to a DynamicValuable type.
	ValueFromDynamic(context.Context, DynamicValue) (DynamicValuable, diag.Diagnostics)
}

var _ DynamicTypable = DynamicType{}

// DynamicType is the base framework type for a string. StringValue is the
// associated value type.
type DynamicType struct{}

// ApplyTerraform5AttributePathStep applies the given AttributePathStep to the
// type.
func (t DynamicType) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	return DynamicType{}, nil
}

// Equal returns true if the given type is equivalent.
func (t DynamicType) Equal(o attr.Type) bool {
	_, ok := o.(DynamicType)

	return ok
}

// String returns a human-readable string of the type name.
func (t DynamicType) String() string {
	return "basetypes.DynamicType"
}

// TerraformType returns the tftypes.Type that should be used to represent this
// framework type.
func (t DynamicType) TerraformType(_ context.Context) tftypes.Type {
	return tftypes.DynamicPseudoType
}

// ValueFromDynamic returns a StringValuable type given a StringValue.
func (t DynamicType) ValueFromDynamic(_ context.Context, v DynamicValue) (DynamicValuable, diag.Diagnostics) {
	return v, nil
}

// ValueFromTerraform returns a Value given a tftypes.Value.  This is meant to
// convert the tftypes.Value into a more convenient Go type for the provider to
// consume the data with.
func (t DynamicType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if !in.IsKnown() {
		return NewDynamicUnknown(), nil
	}

	if in.IsNull() {
		return NewDynamicNull(), nil
	}

	out, err := fromTerraformType(in.Type()).ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	return NewDynamicValue(out), nil
}

// ValueType returns the Value type.
func (t DynamicType) ValueType(_ context.Context) attr.Value {
	// This Value does not need to be valid.
	return DynamicValue{}
}

func fromTerraformType(input tftypes.Type) attr.Type {
	switch {
	case input.Is(tftypes.String):
		return StringType{}
	case input.Is(tftypes.Number):
		return NumberType{}
	case input.Is(tftypes.Bool):
		return BoolType{}
	case input.Is(tftypes.List{}):
		return ListType{ElemType: fromTerraformType(input.(tftypes.List).ElementType)}
	case input.Is(tftypes.Set{}):
		return SetType{ElemType: fromTerraformType(input.(tftypes.Set).ElementType)}
	case input.Is(tftypes.Tuple{}):
		elemType := make([]attr.Type, len(input.(tftypes.Tuple).ElementTypes))
		for i := range input.(tftypes.Tuple).ElementTypes {
			elemType[i] = fromTerraformType(input.(tftypes.Tuple).ElementTypes[i])
		}
		return TupleType{ElemTypes: elemType}
	case input.Is(tftypes.Object{}):
		attrTypes := make(map[string]attr.Type)
		for k, _ := range input.(tftypes.Object).AttributeTypes {
			attrTypes[k] = fromTerraformType(input.(tftypes.Object).AttributeTypes[k])
		}
		return ObjectType{AttrTypes: attrTypes}
	case input.Is(tftypes.Map{}):
		return MapType{ElemType: fromTerraformType(input.(tftypes.Map).ElementType)}
	default:
		return DynamicType{}
	}
}
