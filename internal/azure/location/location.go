package location

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func SchemaLocationOC() *schema.Schema {
	return &schema.Schema{
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

func NormalizeLocation() planmodifier.String {
	return normalizeLocation{}
}

type normalizeLocation struct{}

func (m normalizeLocation) Description(_ context.Context) string {
	return "Normalize the location"
}

func (m normalizeLocation) MarkdownDescription(_ context.Context) string {
	return "Normalize the location"
}

func (m normalizeLocation) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.PlanValue.IsUnknown() {
		return
	}

	resp.PlanValue = types.StringValue(Normalize(req.PlanValue.ValueString()))
}
