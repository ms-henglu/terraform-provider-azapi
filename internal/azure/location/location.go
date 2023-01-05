package location

import (
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"strings"
)

func SchemaLocationOC() *schema.Attribute {
	return &schema.StringAttribute{
		Type:             schema.TypeString,
		Optional:         true,
		ForceNew:         true,
		Computed:         true,
		StateFunc:        LocationStateFunc,
		DiffSuppressFunc: LocationDiffSuppressFunc,
	}
}

func SchemaLocation() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringIsNotEmpty,
	}
}

func SchemaLocationDataSource() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Computed: true,
	}
}

func LocationDiffSuppressFunc(_, old, new string, _ *schema.ResourceData) bool {
	return Normalize(old) == Normalize(new)
}

func LocationStateFunc(location interface{}) string {
	input := location.(string)
	return Normalize(input)
}

func Normalize(input string) string {
	return strings.ReplaceAll(strings.ToLower(input), " ", "")
}
