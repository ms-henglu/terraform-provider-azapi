package utils

import (
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func ToAttrValue(input interface{}) attr.Value {
	if input == nil {
		return types.DynamicNull()
	}
	switch v := input.(type) {
	case string:
		return types.StringValue(v)
	case bool:
		return types.BoolValue(v)
	case int:
		return types.NumberValue(big.NewFloat(float64(v)))
	case int64:
		return types.NumberValue(big.NewFloat(float64(v)))
	case float32:
		return types.NumberValue(big.NewFloat(float64(v)))
	case float64:
		return types.NumberValue(big.NewFloat(v))
	case big.Float:
		return types.NumberValue(&v)
	case *big.Float:
		return types.NumberValue(v)
	case []interface{}:
		out := make([]attr.Value, 0)
		outTypes := make([]attr.Type, 0)
		for _, item := range v {
			out = append(out, ToAttrValue(item))
			outTypes = append(outTypes, AttrTypeOf(item))
		}
		return types.TupleValueMust(outTypes, out)
	case map[string]interface{}:
		out := make(map[string]attr.Value)
		outTypes := make(map[string]attr.Type)
		for k, item := range v {
			out[k] = ToAttrValue(item)
			outTypes[k] = AttrTypeOf(item)
		}
		return types.ObjectValueMust(outTypes, out)
	}
	return nil
}

func AttrTypeOf(input interface{}) attr.Type {
	if input == nil {
		return types.DynamicType
	}
	switch v := input.(type) {
	case string:
		return types.StringType
	case bool:
		return types.BoolType
	case int, int64, float32, float64, big.Float, *big.Float:
		return types.NumberType
	case []interface{}:
		outTypes := make([]attr.Type, 0)
		for _, item := range v {
			outTypes = append(outTypes, AttrTypeOf(item))
		}
		return types.TupleType{ElemTypes: outTypes}
	case map[string]interface{}:
		outTypes := make(map[string]attr.Type)
		for k, item := range v {
			outTypes[k] = AttrTypeOf(item)
		}
		return types.ObjectType{AttrTypes: outTypes}
	}
	return nil
}

func FromTFValue(tfValue tftypes.Value) (interface{}, error) {
	if tfValue.IsNull() {
		return nil, nil
	}
	if !tfValue.IsKnown() {
		return "<unknown>", nil
	}
	tfValueType := tfValue.Type()
	switch {
	case tfValueType.Is(tftypes.String):
		var s string
		err := tfValue.As(&s)
		if err != nil {
			return nil, err
		}
		return s, nil
	case tfValueType.Is(tftypes.Number):
		n := big.NewFloat(0)
		err := tfValue.As(&n)
		if err != nil {
			return nil, err
		}
		out, _ := n.Float64()
		return out, nil
	case tfValueType.Is(tftypes.Bool):
		var b bool
		err := tfValue.As(&b)
		if err != nil {
			return nil, err
		}
		return b, nil
	case tfValueType.Is(tftypes.List{}), tfValueType.Is(tftypes.Set{}), tfValueType.Is(tftypes.Tuple{}):
		var l []tftypes.Value
		err := tfValue.As(&l)
		if err != nil {
			return nil, err
		}
		var out []interface{}
		for _, v := range l {
			vv, err := FromTFValue(v)
			if err != nil {
				return nil, err
			}
			out = append(out, vv)
		}
		return out, nil
	case tfValueType.Is(tftypes.Map{}), tfValueType.Is(tftypes.Object{}):
		var m map[string]tftypes.Value
		err := tfValue.As(&m)
		if err != nil {
			return nil, err
		}
		out := make(map[string]interface{})
		for k, v := range m {
			vv, err := FromTFValue(v)
			if err != nil {
				return nil, err
			}
			out[k] = vv
		}
		return out, nil
	}
	return nil, nil
}
