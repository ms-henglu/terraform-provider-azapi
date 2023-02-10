package tags

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SchemaTagsOC() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeMap,
		Optional:     true,
		Computed:     true,
		ValidateFunc: ValidateTags,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
}

func SchemaTags() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeMap,
		Optional:     true,
		ValidateFunc: ValidateTags,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
}

func SchemaTagsDataSource() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Computed: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
}

func ValidateTags(v interface{}, _ string) (warnings []string, errors []error) {
	tagsMap := v.(map[string]interface{})

	if len(tagsMap) > 50 {
		errors = append(errors, fmt.Errorf("a maximum of 50 tags can be applied to each ARM resource"))
	}

	for k, v := range tagsMap {
		if len(k) > 512 {
			errors = append(errors, fmt.Errorf("the maximum length for a tag key is 512 characters: %q is %d characters", k, len(k)))
		}

		value, err := TagValueToString(v)
		if err != nil {
			errors = append(errors, err)
		} else if len(value) > 256 {
			errors = append(errors, fmt.Errorf("the maximum length for a tag value is 256 characters: the value for %q is %d characters", k, len(value)))
		}
	}

	return warnings, errors
}

func TagValueToString(v interface{}) (string, error) {
	switch value := v.(type) {
	case string:
		return value, nil
	case int:
		return fmt.Sprintf("%d", value), nil
	default:
		return "", fmt.Errorf("unknown tag type %T in tag value", value)
	}
}

func FlattenTags(raw interface{}) types.Map {
	if raw == nil {
		return types.MapNull(types.StringType)
	}
	if input, ok := raw.(map[string]interface{}); ok {
		out := make(map[string]attr.Value)
		for k, v := range input {
			out[k] = types.StringValue(v.(string))
		}
		return types.MapValueMust(types.StringType, out)
	}
	return types.MapNull(types.StringType)
}

func ExpandTags(value types.Map) map[string]string {
	tagsMap := make(map[string]interface{})
	if diags := value.ElementsAs(context.TODO(), &tagsMap, false); diags.HasError() {
		return nil
	}
	output := make(map[string]string, len(tagsMap))

	for i, v := range tagsMap {
		// Validate should have ignored this error already
		value, _ := TagValueToString(v)
		output[i] = value
	}

	return output
}

type tagsValidator struct{}

var _ validator.Map = tagsValidator{}

func (v tagsValidator) ValidateMap(ctx context.Context, request validator.MapRequest, response *validator.MapResponse) {
	value := request.ConfigValue

	if value.IsUnknown() || value.IsNull() {
		return
	}

	tagsMap := value.Elements()

	if len(tagsMap) > 50 {
		response.Diagnostics.AddAttributeError(
			request.Path,
			"Tags Length Exceeds Limit",
			"a maximum of 50 tags can be applied to each ARM resource")
	}

	for k, v := range tagsMap {
		if len(k) > 512 {
			response.Diagnostics.AddAttributeError(
				request.Path,
				"Tag Key Length Exceeds Limit",
				fmt.Sprintf("the maximum length for a tag key is 512 characters: %q is %d characters", k, len(k)))
		}

		value, err := TagValueToString2(v)
		if err != nil {
			response.Diagnostics.AddAttributeError(
				request.Path,
				"Invalid Tag Value",
				fmt.Sprintf("invalid tag value: %q", k))
		} else if len(value) > 256 {
			response.Diagnostics.AddAttributeError(
				request.Path,
				"Tag Value Length Exceeds Limit",
				fmt.Sprintf("the maximum length for a tag value is 256 characters: the value for %q is %d characters", k, len(value)))
		}
	}
}

func TagValueToString2(v attr.Value) (string, error) {
	if v.IsUnknown() || v.IsNull() {
		return "", nil
	}
	switch {
	case v.Type(context.TODO()).Equal(types.StringType):
		return v.String(), nil
	case v.Type(context.TODO()).Equal(types.Int64Type):
		return v.String(), nil
	default:
		return "", fmt.Errorf("unknown tag type %T in tag value", v.String())
	}
}

func (v tagsValidator) Description(ctx context.Context) string {
	return "validate this in tags format"
}

func (v tagsValidator) MarkdownDescription(ctx context.Context) string {
	return "validate this in tags format"
}

func Validator() tagsValidator {
	return tagsValidator{}
}
