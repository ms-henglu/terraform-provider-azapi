package validate

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"regexp"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
)

func ResourceID(input interface{}, key string) (warnings []string, errors []error) {
	v, ok := input.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected %q to be a string", key))
		return
	}

	if v == "/" {
		return
	}

	r := regexp.MustCompile("^http[s]?:.*")
	if r.MatchString(v) {
		errors = append(errors, fmt.Errorf("expected %q not to contain protocol", key))
	}
	r = regexp.MustCompile(".*api-version=.*")
	if r.MatchString(v) {
		errors = append(errors, fmt.Errorf("expected %q not to contain api-version", key))
	}

	if _, err := arm.ParseResourceID(v); err != nil {
		errors = append(errors, err)
	}

	return
}

func ResourceType(i interface{}, k string) ([]string, []error) {
	v, ok := i.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be string", k)}
	}

	if v == "" {
		return nil, []error{fmt.Errorf("expected %q to not be an empty string, got %v", k, i)}
	}

	parts := strings.Split(v, "@")
	if len(parts) != 2 {
		return nil, []error{fmt.Errorf("expected %q to be <resource-type>@<api-version>", k)}
	}

	return nil, nil
}

type stringIsResourceID struct{}

func (v stringIsResourceID) Description(ctx context.Context) string {
	return "validate this in resource ID format"
}

func (v stringIsResourceID) MarkdownDescription(ctx context.Context) string {
	return "validate this in resource ID format"
}

func (_ stringIsResourceID) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	str := req.ConfigValue

	if str.IsUnknown() || str.IsNull() {
		return
	}

	if _, errs := ResourceID(str.ValueString(), req.Path.String()); len(errs) != 0 {
		for _, err := range errs {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid Resource ID",
				err.Error())
		}
	}
}

func StringIsResourceID() stringIsResourceID {
	return stringIsResourceID{}
}

type stringIsResourceType struct{}

func (v stringIsResourceType) Description(ctx context.Context) string {
	return "validate this in resource type format"
}

func (v stringIsResourceType) MarkdownDescription(ctx context.Context) string {
	return "validate this in resource type format"
}

func (_ stringIsResourceType) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	str := req.ConfigValue

	if str.IsUnknown() || str.IsNull() {
		return
	}

	if _, errs := ResourceType(str.ValueString(), req.Path.String()); len(errs) != 0 {
		for _, err := range errs {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid Resource Type",
				err.Error())
		}
	}
}

func StringIsResourceType() stringIsResourceType {
	return stringIsResourceType{}
}
