package validator

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-uuid"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type stringIsUUID struct{}

func (v stringIsUUID) Description(ctx context.Context) string {
	return "validate this in UUID format"
}

func (v stringIsUUID) MarkdownDescription(ctx context.Context) string {
	return "validate this in UUID format"
}

func (_ stringIsUUID) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	str := req.ConfigValue

	if str.IsUnknown() || str.IsNull() {
		return
	}

	if _, err := uuid.ParseUUID(str.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid UUID",
			fmt.Sprintf("expected a valid UUID, got %v", err),
		)
	}
}

func StringIsUUID() stringIsUUID {
	return stringIsUUID{}
}
