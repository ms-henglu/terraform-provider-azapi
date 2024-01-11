package utils

import (
	"encoding/json"
	"fmt"
	"github.com/Azure/terraform-provider-azapi/internal/azure/types"
	"regexp"
	"strings"
)

func NormalizeJson(jsonString interface{}) string {
	if jsonString == nil || jsonString == "" {
		return ""
	}
	var j interface{}

	if err := json.Unmarshal([]byte(jsonString.(string)), &j); err != nil {
		return fmt.Sprintf("Error parsing JSON: %+v", err)
	}
	b, _ := json.Marshal(j)
	return string(b)
}

// MergeObject is used to merge object old and new, if overlaps, use new value
func MergeObject(old interface{}, new interface{}) interface{} {
	if new == nil {
		return new
	}
	switch oldValue := old.(type) {
	case map[string]interface{}:
		if newMap, ok := new.(map[string]interface{}); ok {
			res := make(map[string]interface{})
			for key, value := range oldValue {
				if _, ok := newMap[key]; ok {
					res[key] = MergeObject(value, newMap[key])
				} else {
					res[key] = value
				}
			}
			for key, newValue := range newMap {
				if res[key] == nil {
					res[key] = newValue
				}
			}
			return res
		}
	case []interface{}:
		if newArr, ok := new.([]interface{}); ok {
			if len(oldValue) != len(newArr) {
				return newArr
			}
			res := make([]interface{}, 0)
			for index := range oldValue {
				res = append(res, MergeObject(oldValue[index], newArr[index]))
			}
			return res
		}
	}
	return new
}

type UpdateJsonOption struct {
	IgnoreCasing          bool
	IgnoreMissingProperty bool
	Type                  *types.TypeBase
}

func AreSameItems(a interface{}, b interface{}) bool {
	if a == nil || b == nil {
		return a == b
	}
	aName := ""
	if aMap, ok := a.(map[string]interface{}); ok {
		if aMap["name"] != nil {
			aName = aMap["name"].(string)
		}
	}
	bName := ""
	if bMap, ok := b.(map[string]interface{}); ok {
		if bMap["name"] != nil {
			bName = bMap["name"].(string)
		}
	}
	return aName == bName && aName != ""
}

// UpdateObject is used to get an updated object which has same schema as old, but with new value
func UpdateObject(old interface{}, new interface{}, option UpdateJsonOption) interface{} {
	switch oldValue := old.(type) {
	case map[string]interface{}:
		if newMap, ok := new.(map[string]interface{}); ok {
			res := make(map[string]interface{})
			for key, value := range oldValue {
				switch {
				case newMap[key] != nil:
					var propertyType *types.TypeBase
					if option.Type != nil {
						propertyType = (*option.Type).TypeOfProperty(old, key)
					}
					res[key] = UpdateObject(value, newMap[key], UpdateJsonOption{
						IgnoreCasing:          option.IgnoreCasing,
						IgnoreMissingProperty: option.IgnoreMissingProperty,
						Type:                  propertyType,
					})
				case option.IgnoreMissingProperty || isZeroValue(value):
					res[key] = value
				}
			}
			return res
		}
	case []interface{}:
		if newArr, ok := new.([]interface{}); ok {
			if len(oldValue) != len(newArr) {
				out := make([]interface{}, 0)
				newArrAdded := make([]bool, len(newArr))
				for _, o := range oldValue {
					for index, n := range newArr {
						if !newArrAdded[index] && AreSameItems(o, n) {
							newArrAdded[index] = true
							out = append(out, UpdateObject(o, n, option))
							break
						}
					}
				}
				var itemType *types.TypeBase
				if option.Type != nil {
					itemType = (*option.Type).TypeOfProperty(old, types.ArrayItem)
				}
				for index, n := range newArr {
					if !newArrAdded[index] {
						if itemType != nil {
							n = (*itemType).GetWriteOnly(n)
						}
						out = append(out, n)
					}
				}
				return out
			}
			res := make([]interface{}, 0)
			for index := range oldValue {
				res = append(res, UpdateObject(oldValue[index], newArr[index], option))
			}
			return res
		}
	case string:
		if newStr, ok := new.(string); ok {
			if option.IgnoreCasing && strings.EqualFold(oldValue, newStr) {
				return oldValue
			}
			if option.IgnoreMissingProperty && (regexp.MustCompile(`^\*+$`).MatchString(newStr) || "<redacted>" == newStr || "" == newStr) {
				return oldValue
			}
		}
	}
	return new
}

// ExtractObject is used to extract object from old for a json path
func ExtractObject(old interface{}, path string) interface{} {
	if len(path) == 0 {
		return old
	}
	if oldMap, ok := old.(map[string]interface{}); ok {
		index := strings.Index(path, ".")
		if index != -1 {
			key := path[0:index]
			result := make(map[string]interface{}, 1)
			value := ExtractObject(oldMap[key], path[index+1:])
			if value == nil {
				return nil
			} else {
				result[key] = value
			}
			return result
		} else {
			if oldMap[path] != nil {
				result := make(map[string]interface{}, 1)
				result[path] = oldMap[path]
				return result
			} else {
				return nil
			}
		}
	}
	return nil
}

// OverrideWithPaths is used to override old object with new object for specific paths
func OverrideWithPaths(old interface{}, new interface{}, path string, pathSet map[string]bool) (interface{}, error) {
	if len(pathSet) == 0 || old == nil {
		return old, nil
	}
	switch oldValue := old.(type) {
	case map[string]interface{}:
		if newMap, ok := new.(map[string]interface{}); ok {
			outMap := make(map[string]interface{})
			for key, value := range oldValue {
				if newValue, ok := newMap[key]; ok {
					nestedPath := strings.TrimPrefix(path+"."+key, ".")
					out, err := OverrideWithPaths(value, newValue, nestedPath, pathSet)
					if err != nil {
						return nil, err
					}
					outMap[key] = out
				} else {
					outMap[key] = value
				}
			}
			return outMap, nil
		}
	case []interface{}:
		// Does not support override specific item in list
		for v := range pathSet {
			if strings.HasPrefix(v, path+".") {
				return nil, fmt.Errorf("ignoring specific item in list is not supported")
			}
		}
		if newArr, ok := new.([]interface{}); ok && pathSet[path] {
			return mergeArray(oldValue, newArr), nil
		}
	default:
		if _, ok := pathSet[path]; ok {
			return new, nil
		}
	}

	return old, nil
}

// mergeArray is used to merge two array, if overlaps, use old value. `name` is used as key to compare
func mergeArray(old []interface{}, new []interface{}) []interface{} {
	oldMap := make(map[string]interface{})
	for _, v := range old {
		if vMap, ok := v.(map[string]interface{}); ok {
			if name, ok := vMap["name"]; ok {
				oldMap[name.(string)] = v
			}
		}
	}
	out := make([]interface{}, 0)
	for _, v := range new {
		if vMap, ok := v.(map[string]interface{}); ok {
			if name, ok := vMap["name"]; ok {
				if oldV, ok := oldMap[name.(string)]; ok {
					out = append(out, oldV)
					continue
				}
			}
		}
		out = append(out, v)
	}
	return out
}

// NormalizeObject is used to remove customized type and replaced with builtin type
func NormalizeObject(input interface{}) interface{} {
	jsonString, _ := json.Marshal(input)
	var output interface{}
	_ = json.Unmarshal(jsonString, &output)
	return output
}

func isZeroValue(value interface{}) bool {
	if value == nil {
		return true
	}
	switch v := value.(type) {
	case map[string]interface{}:
		return len(v) == 0
	case []interface{}:
		return len(v) == 0
	case string:
		return len(v) == 0
	case int, int32, int64, float32, float64:
		return v == 0
	case bool:
		return !v
	}
	return false
}
