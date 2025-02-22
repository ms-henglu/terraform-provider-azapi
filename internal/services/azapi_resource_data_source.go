package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/terraform-provider-azapi/internal/azure/identity"
	"github.com/Azure/terraform-provider-azapi/internal/azure/location"
	"github.com/Azure/terraform-provider-azapi/internal/azure/tags"
	"github.com/Azure/terraform-provider-azapi/internal/clients"
	"github.com/Azure/terraform-provider-azapi/internal/services/parse"
	"github.com/Azure/terraform-provider-azapi/internal/services/validate"
	"github.com/Azure/terraform-provider-azapi/internal/tf"
	"github.com/Azure/terraform-provider-azapi/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func AzApiDataSource() *schema.Resource {
	return &schema.Resource{
		Read: resourceAzApiDataSourceRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				ValidateFunc:  validation.StringIsNotEmpty,
				ConflictsWith: []string{"resource_id"},
			},

			"parent_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ValidateFunc:  validate.ResourceID,
				ConflictsWith: []string{"resource_id"},
			},

			"resource_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ValidateFunc:  validate.ResourceID,
				ConflictsWith: []string{"name", "parent_id"},
			},

			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validate.ResourceType,
			},

			"response_export_values": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringIsNotEmpty,
				},
			},

			"location": location.SchemaLocationDataSource(),

			"identity": identity.SchemaIdentityDataSource(),

			"output": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tags.SchemaTagsDataSource(),
		},
	}
}

func resourceAzApiDataSourceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ResourceClient
	ctx, cancel := tf.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	var id parse.ResourceId
	resourceType := d.Get("type").(string)
	if name := d.Get("name").(string); len(name) != 0 {
		parentId := d.Get("parent_id").(string)
		if parentId == "" && strings.HasPrefix(strings.ToUpper(resourceType), strings.ToUpper(arm.ResourceGroupResourceType.String())) {
			parentId = fmt.Sprintf("/subscriptions/%s", meta.(*clients.Client).Account.GetSubscriptionId())
		}

		buildId, err := parse.NewResourceID(name, parentId, resourceType)
		if err != nil {
			return err
		}
		id = buildId
	} else {
		resourceId := d.Get("resource_id").(string)
		if resourceId == "" && strings.HasPrefix(strings.ToUpper(resourceType), strings.ToUpper(arm.SubscriptionResourceType.String())) {
			resourceId = fmt.Sprintf("/subscriptions/%s", meta.(*clients.Client).Account.GetSubscriptionId())
		}
		buildId, err := parse.ResourceIDWithResourceType(resourceId, resourceType)
		if err != nil {
			return err
		}
		id = buildId
	}

	responseBody, err := client.Get(ctx, id.AzureResourceId, id.ApiVersion)
	if err != nil {
		if utils.ResponseErrorWasNotFound(err) {
			return fmt.Errorf("not found %q: %+v", id, err)
		}
		return fmt.Errorf("reading %q: %+v", id, err)
	}
	d.SetId(id.ID())
	// #nosec G104
	d.Set("name", id.Name)
	// #nosec G104
	d.Set("parent_id", id.ParentId)
	if bodyMap, ok := responseBody.(map[string]interface{}); ok {
		// #nosec G104
		d.Set("tags", tags.FlattenTags(bodyMap["tags"]))
		// #nosec G104
		d.Set("location", bodyMap["location"])
		// #nosec G104
		d.Set("identity", identity.FlattenIdentity(bodyMap["identity"]))
	}
	// #nosec G104
	d.Set("output", flattenOutput(responseBody, d.Get("response_export_values").([]interface{})))
	return nil
}
