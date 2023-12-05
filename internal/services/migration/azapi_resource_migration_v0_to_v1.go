package migration

import (
	"context"
	"github.com/Azure/terraform-provider-azapi/internal/services/validate"
	"github.com/Azure/terraform-provider-azapi/internal/tf"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type AzapiResourceV0ToV1 struct{}

func (s AzapiResourceV0ToV1) Schema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},

		"removing_special_chars": {
			Type:     schema.TypeBool,
			Optional: true,
		},

		"parent_id": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},

		"type": {
			Type:     schema.TypeString,
			Required: true,
		},

		"location": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},

		"identity": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Type:     schema.TypeString,
						Required: true,
					},

					"identity_ids": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Schema{
							Type:         schema.TypeString,
							ValidateFunc: validate.UserAssignedIdentityID,
						},
					},

					"principal_id": {
						Type:     schema.TypeString,
						Computed: true,
					},

					"tenant_id": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
		},

		"body": {
			Type:     schema.TypeString,
			Optional: true,
		},

		"ignore_casing": {
			Type:     schema.TypeBool,
			Optional: true,
		},

		"ignore_body_changes": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},

		"ignore_missing_property": {
			Type:     schema.TypeBool,
			Optional: true,
		},

		"response_export_values": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},

		"locks": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},

		"schema_validation_enabled": {
			Type:     schema.TypeBool,
			Optional: true,
		},

		"output": {
			Type:     schema.TypeString,
			Computed: true,
		},

		"tags": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
}

func (s AzapiResourceV0ToV1) UpgradeFunc() tf.StateUpgraderFunc {
	return func(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {

		return rawState, nil
	}
}
