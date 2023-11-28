// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type Dynamic = basetypes.DynamicValue

// DynamicNull creates a String with a null value. Determine whether the value is
// null via the String type IsNull method.
func DynamicNull() basetypes.DynamicValue {
	return basetypes.NewDynamicNull()
}

// DynamicUnknown creates a String with an unknown value. Determine whether the
// value is unknown via the String type IsUnknown method.
func DynamicUnknown() basetypes.DynamicValue {
	return basetypes.NewDynamicUnknown()
}

// DynamicValue creates a String with a known value. Access the value via the String
// type ValueString method.
func DynamicValue(value attr.Value) basetypes.DynamicValue {
	return basetypes.NewDynamicValue(value)
}

// DynamicPointerValue creates a String with a null value if nil or a known value.
func DynamicPointerValue(value *attr.Value) basetypes.DynamicValue {
	return basetypes.NewDynamicPointerValue(value)
}
