package location

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func SchemaLocationOC() schema.StringAttribute {
	return schema.StringAttribute{
		Optional: true,
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
		CustomType: LocationStringType{},
	}
}

func SchemaLocation() schema.StringAttribute {
	return schema.StringAttribute{
		Optional: true,
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
		CustomType: LocationStringType{},
	}
}

func SchemaLocationDataSource() schema.StringAttribute {
	return schema.StringAttribute{
		Computed: true,
	}
}

func Normalize(input string) string {
	return strings.ReplaceAll(strings.ToLower(input), " ", "")
}
