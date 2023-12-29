// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package basetypes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ DynamicValuable = DynamicValue{}
)

// DynamicValuable extends attr.Value for dynamic value types.
// Implement this interface to create a custom Dynamic value type.
type DynamicValuable interface {
	attr.Value

	// ToDynamicValue should convert the value type to a Dynamic value.
	ToDynamicValue(ctx context.Context) (DynamicValue, diag.Diagnostics)
}

// DynamicValuableWithSemanticEquals extends DynamicValuable with semantic
// equality logic.
type DynamicValuableWithSemanticEquals interface {
	DynamicValuable

	// DynamicSemanticEquals should return true if the given value is
	// semantically equal to the current value. This logic is used to prevent
	// Terraform data consistency errors and resource drift where a value change
	// may have inconsequential differences, such as spacing character removal
	// in JSON formatted strings.
	//
	// Only known values are compared with this method as changing a value's
	// state implicitly represents a different value.
	DynamicSemanticEquals(context.Context, DynamicValuable) (bool, diag.Diagnostics)
}

// NewDynamicNull creates a Dynamic with a null value. Determine whether the value is
// null via the Dynamic type IsNull method.
func NewDynamicNull() DynamicValue {
	return DynamicValue{
		state: attr.ValueStateNull,
	}
}

// NewDynamicUnknown creates a String with an unknown value. Determine whether the
// value is unknown via the Dynamic type IsUnknown method.
func NewDynamicUnknown() DynamicValue {
	return DynamicValue{
		state: attr.ValueStateUnknown,
	}
}

// NewDynamicValue creates a Dynamic with a known value. Access the value via the String
// type ValueString method.
func NewDynamicValue(value attr.Value) DynamicValue {
	return DynamicValue{
		state: attr.ValueStateKnown,
		value: value,
	}
}

// NewDynamicPointerValue creates a String with a null value if nil or a known
// value. Access the value via the String type ValueStringPointer method.
func NewDynamicPointerValue(value *attr.Value) DynamicValue {
	if value == nil {
		return NewDynamicNull()
	}

	return NewDynamicValue(*value)
}

// DynamicValue represents a UTF-8 string value.
type DynamicValue struct {
	// state represents whether the value is null, unknown, or known. The
	// zero-value is null.
	state attr.ValueState

	// value contains the known value, if not null or unknown.
	value attr.Value
}

// Type returns a StringType.
func (s DynamicValue) Type(_ context.Context) attr.Type {
	return DynamicType{}
}

// ToTerraformValue returns the data contained in the *String as a tftypes.Value.
func (s DynamicValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	switch s.state {
	case attr.ValueStateKnown:
		return s.value.ToTerraformValue(ctx)
	case attr.ValueStateNull:
		return tftypes.NewValue(tftypes.DynamicPseudoType, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(tftypes.DynamicPseudoType, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled DynamicPseudoType state in ToTerraformValue: %s", s.state))
	}
}

// Equal returns true if `other` is a String and has the same value as `s`.
func (s DynamicValue) Equal(other attr.Value) bool {
	o, ok := other.(DynamicValue)

	if !ok {
		return false
	}

	if s.state != o.state {
		return false
	}

	if s.state != attr.ValueStateKnown {
		return true
	}

	return s.value.Equal(o.value)
}

// IsNull returns true if the String represents a null value.
func (s DynamicValue) IsNull() bool {
	return s.state == attr.ValueStateNull
}

// IsUnknown returns true if the String represents a currently unknown value.
func (s DynamicValue) IsUnknown() bool {
	return s.state == attr.ValueStateUnknown
}

// String returns a human-readable representation of the String value. Use
// the ValueString method for Terraform data handling instead.
//
// The string returned here is not protected by any compatibility guarantees,
// and is intended for logging and error reporting.
func (s DynamicValue) String() string {
	if s.IsUnknown() {
		return attr.UnknownValueString
	}

	if s.IsNull() {
		return attr.NullValueString
	}

	return s.value.String()
}

// ValueDynamic returns the known string value. If String is null or unknown, returns
// "".
func (s DynamicValue) ValueDynamic() attr.Value {
	return s.value
}

// ValueDynamicPointer returns a pointer to the known string value, nil for a
// null value, or a pointer to "" for an unknown value.
func (s DynamicValue) ValueDynamicPointer() *attr.Value {
	if s.IsNull() {
		return nil
	}

	return &s.value
}

// ToDynamicValue returns String.
func (s DynamicValue) ToDynamicValue(context.Context) (DynamicValue, diag.Diagnostics) {
	return s, nil
}
