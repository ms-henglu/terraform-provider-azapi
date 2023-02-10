package features

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type UserFeatures struct {
	DefaultTags         types.Map
	DefaultLocation     types.String
	DefaultNaming       string
	DefaultNamingPrefix string
	DefaultNamingSuffix string
	CafEnabled          bool
}

func Default() UserFeatures {
	return UserFeatures{
		DefaultTags:         types.MapNull(types.StringType),
		DefaultLocation:     types.StringNull(),
		DefaultNaming:       "",
		DefaultNamingPrefix: "",
		DefaultNamingSuffix: "",
		CafEnabled:          false,
	}
}
