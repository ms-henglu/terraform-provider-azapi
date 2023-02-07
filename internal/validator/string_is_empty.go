package validator

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type stringIsEmpty struct{}

func (v stringIsEmpty) Description(ctx context.Context) string {
	return "validate this in UUID format"
}

func (v stringIsEmpty) MarkdownDescription(ctx context.Context) string {
	return "validate this in UUID format"
}

func (_ stringIsEmpty) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	str := req.ConfigValue

	if str.IsUnknown() || str.IsNull() {
		return
	}

	if str.ValueString() != "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"String Not Empty",
			fmt.Sprintf("expected an empty string, got %v", str.ValueString()),
		)
	}
}

func StringIsEmpty() stringIsEmpty {
	return stringIsEmpty{}
}
