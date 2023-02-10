package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/terraform-provider-azapi/internal/azure"
	"github.com/Azure/terraform-provider-azapi/internal/azure/identity"
	"github.com/Azure/terraform-provider-azapi/internal/azure/location"
	"github.com/Azure/terraform-provider-azapi/internal/azure/tags"
	azuretypes "github.com/Azure/terraform-provider-azapi/internal/azure/types"
	"github.com/Azure/terraform-provider-azapi/internal/clients"
	"github.com/Azure/terraform-provider-azapi/internal/locks"
	myplanmodifier "github.com/Azure/terraform-provider-azapi/internal/planmodifier"
	"github.com/Azure/terraform-provider-azapi/internal/services/parse"
	"github.com/Azure/terraform-provider-azapi/internal/services/validate"
	"github.com/Azure/terraform-provider-azapi/internal/tf"
	myValidator "github.com/Azure/terraform-provider-azapi/internal/validator"
	"github.com/Azure/terraform-provider-azapi/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"log"
)

var _ resource.Resource = &AzapiResource{}
var _ resource.ResourceWithConfigure = &AzapiResource{}
var _ resource.ResourceWithModifyPlan = &AzapiResource{}
var _ resource.ResourceWithValidateConfig = &AzapiResource{}

//var _ resource.ResourceWithUpgradeState = &AzapiResource{}

type AzapiResource struct {
	ProviderData *clients.Client
}

type AzapiResourceModel struct {
	ID                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	ParentID                types.String `tfsdk:"parent_id"`
	Type                    types.String `tfsdk:"type"`
	Location                types.String `tfsdk:"location"`
	Identity                types.List   `tfsdk:"identity"`
	Body                    types.String `tfsdk:"body"`
	Locks                   types.List   `tfsdk:"locks"`
	RemovingSpecialChars    types.Bool   `tfsdk:"removing_special_chars"`
	SchemaValidationEnabled types.Bool   `tfsdk:"schema_validation_enabled"`
	IgnoreCasing            types.Bool   `tfsdk:"ignore_casing"`
	IgnoreMissingProperty   types.Bool   `tfsdk:"ignore_missing_property"`
	ResponseExportValues    types.List   `tfsdk:"response_export_values"`
	Output                  types.String `tfsdk:"output"`
	Tags                    types.Map    `tfsdk:"tags"`
}

func (r *AzapiResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*clients.Client); ok && v != nil {
		r.ProviderData = v
	}
}

func (r *AzapiResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_resource"
}

func (r *AzapiResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"parent_id": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					validate.StringIsResourceID(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"type": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					validate.StringIsResourceType(),
				},
			},

			"location": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{
					//	location.NormalizeLocation(),
				},
			},

			"body": schema.StringAttribute{
				Optional: true,
				Computed: true,
				//DiffSuppressFunc: tf.SuppressJsonOrderingDifference,
				PlanModifiers: []planmodifier.String{
					myplanmodifier.DefaultAttribute(types.StringValue("{}")),
				},
				Validators: []validator.String{
					myValidator.StringIsJSON(),
				},
			},

			"ignore_casing": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					myplanmodifier.DefaultAttribute(types.BoolValue(false)),
				},
			},

			"ignore_missing_property": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					myplanmodifier.DefaultAttribute(types.BoolValue(true)),
				},
			},

			"response_export_values": schema.ListAttribute{
				Optional: true,
				ElementType: types.ListType{
					ElemType: types.StringType,
				},
				// TODO@ms-henglu： validate each element in this list: not empty
			},

			"locks": schema.ListAttribute{
				Optional: true,
				ElementType: types.ListType{
					ElemType: types.StringType,
				},
				// TODO@ms-henglu： validate each element in this list: not empty
			},

			"removing_special_chars": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					myplanmodifier.DefaultAttribute(types.BoolValue(false)),
				},
			},

			"schema_validation_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					myplanmodifier.DefaultAttribute(types.BoolValue(true)),
				},
			},

			"output": schema.StringAttribute{
				Computed: true,
			},

			"tags": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.Map{
					tags.Validator(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"identity": identity.SchemaIdentity(),
		},
	}
}

func (r *AzapiResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var config AzapiResourceModel
	if diags := request.Config.Get(ctx, &config); diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	// validate parent_id if it's known
	if !config.ParentID.IsUnknown() {
		_, err := parse.NewResourceID(config.Name.ValueString(), config.ParentID.ValueString(), config.Type.ValueString())
		if err != nil {
			response.Diagnostics.AddError("Validation", err.Error())
			return
		}
	}

	// validate body if it's known
	if !config.Body.IsUnknown() {
		return
	}
	var body map[string]interface{}
	err := json.Unmarshal([]byte(config.Body.ValueString()), &body)
	if err != nil {
		response.Diagnostics.AddError("Unmarshal", err.Error())
		return
	}
	validateDuplicatedDefinition(config, body, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	// swagger schema based validation
	if config.Type.IsUnknown() {
		return
	}
	azureResourceType, apiVersion, err := utils.GetAzureResourceTypeApiVersion(config.Type.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Validation", err.Error())
		return
	}
	resourceDef, _ := azure.GetResourceDefinition(azureResourceType, apiVersion)
	if config.SchemaValidationEnabled.ValueBool() {
		r.expandBody(config, body, resourceDef, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}
		validateBodySchema(azureResourceType, apiVersion, resourceDef, body, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}
	}
}

func (r *AzapiResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	if request.Plan.Raw.IsNull() {
		// If the entire plan is null, the resource is planned for destruction.
		return
	}

	var plan AzapiResourceModel
	if diags := request.Plan.Get(ctx, &plan); diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}
	var state *AzapiResourceModel
	if diags := request.State.Get(ctx, &state); diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	if state == nil || !plan.Identity.Equal(state.Identity) ||
		!plan.Tags.Equal(state.Tags) ||
		!plan.ResponseExportValues.Equal(state.ResponseExportValues) ||
		plan.Body.IsUnknown() ||
		utils.NormalizeJson(plan.Body.ValueString()) != utils.NormalizeJson(state.Body.ValueString()) {
		plan.Output = types.StringUnknown()
	}

	// body refers other resource, can't be verified during plan
	if plan.Body.IsUnknown() || plan.Type.IsUnknown() {
		response.Plan.Set(ctx, plan)
		return
	}
	azureResourceType, apiVersion, err := utils.GetAzureResourceTypeApiVersion(plan.Type.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Validation", err.Error())
		return
	}

	var body map[string]interface{}
	err = json.Unmarshal([]byte(plan.Body.ValueString()), &body)
	if err != nil {
		response.Diagnostics.AddError("Unmarshal", err.Error())
		return
	}
	resourceDef, _ := azure.GetResourceDefinition(azureResourceType, apiVersion)
	if plan.Tags.IsNull() && body["tags"] == nil && !r.ProviderData.Features.DefaultTags.IsNull() {
		if isResourceHasProperty(resourceDef, "tags") {
			if state == nil || !state.Tags.Equal(r.ProviderData.Features.DefaultTags) {
				plan.Tags = r.ProviderData.Features.DefaultTags
			}
		}
	}

	if plan.Location.IsNull() && body["location"] == nil && !r.ProviderData.Features.DefaultLocation.IsNull() {
		if isResourceHasProperty(resourceDef, "location") {
			if state == nil || location.Normalize(state.Location.ValueString()) != location.Normalize(r.ProviderData.Features.DefaultLocation.String()) {
				plan.Location = r.ProviderData.Features.DefaultLocation
			}
		}
	}

	response.Plan.Set(ctx, plan)
}

func (r *AzapiResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	state := response.State
	r.CreateUpdate(ctx, request.Plan, &state, &response.Diagnostics)
	response.State = state
}

func (r *AzapiResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	state := response.State
	r.CreateUpdate(ctx, request.Plan, &state, &response.Diagnostics)
	response.State = state
}

func (r *AzapiResource) CreateUpdate(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State, diagnostics *diag.Diagnostics) {
	client := r.ProviderData.ResourceClient
	var model *AzapiResourceModel
	diagnostics.Append(plan.Get(ctx, &model)...)
	if diagnostics.HasError() {
		return
	}

	id, err := parse.NewResourceID(model.Name.ValueString(), model.ParentID.ValueString(), model.Type.ValueString())
	if err != nil {
		diagnostics.AddError("Error Parsing ID", err.Error())
		return
	}

	if isNewResource := state.Raw.IsNull(); isNewResource {
		_, err = client.Get(ctx, id.AzureResourceId, id.ApiVersion)
		if err == nil {
			diagnostics.AddError("Import As Exists Error", tf.ImportAsExistsError("azapi_resource", id.ID()).Error())
			return
		}
		if !utils.ResponseErrorWasNotFound(err) {
			diagnostics.AddError("Reading Resource", fmt.Errorf("checking for presence of existing %s: %+v", id, err).Error())
			return
		}
	}

	var body map[string]interface{}
	err = json.Unmarshal([]byte(model.Body.ValueString()), &body)
	if err != nil {
		diagnostics.AddError("JSON Unmarshal Error", err.Error())
		return
	}

	validateDuplicatedDefinition(*model, body, diagnostics)
	if diagnostics.HasError() {
		return
	}

	r.expandBody(*model, body, id.ResourceDef, diagnostics)
	if diagnostics.HasError() {
		return
	}

	if model.SchemaValidationEnabled.ValueBool() {
		validateBodySchema(id.AzureResourceType, id.ApiVersion, id.ResourceDef, body, diagnostics)
		if diagnostics.HasError() {
			return
		}
	}

	for _, element := range model.Locks.Elements() {
		lockId := element.(types.String).ValueString()
		locks.ByID(lockId)
		defer locks.UnlockByID(lockId)
	}

	_, err = client.CreateOrUpdate(ctx, id.AzureResourceId, id.ApiVersion, body)
	if err != nil {
		diagnostics.AddError("Create Error", fmt.Errorf("creating/updating %q: %+v", id, err).Error())
	}

	model.ID = types.StringValue(id.ID())
	diags := state.Set(ctx, model)
	diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	readResponse := resource.ReadResponse{
		State: *state,
	}
	r.Read(ctx, resource.ReadRequest{
		State: *state,
	}, &readResponse)
	*state = readResponse.State
}

func (r *AzapiResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	client := r.ProviderData.ResourceClient

	var model *AzapiResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &model)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := parse.ResourceIDWithResourceType(model.ID.ValueString(), model.Type.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Error Parsing ID", err.Error())
		return
	}

	responseBody, err := client.Get(ctx, id.AzureResourceId, id.ApiVersion)
	if err != nil {
		if utils.ResponseErrorWasNotFound(err) {
			log.Printf("[INFO] Error reading %q - removing from state", id.ID())
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError("Reading Error", fmt.Errorf("reading %q: %+v", id, err).Error())
		return
	}

	bodyJson := model.Body.ValueString()
	var requestBody interface{}
	err = json.Unmarshal([]byte(bodyJson), &requestBody)
	if err != nil && len(bodyJson) != 0 {
		response.Diagnostics.AddError("Unmarshal Error", err.Error())
		return
	}

	// if it's imported
	if model.Name.IsNull() {
		if id.ResourceDef != nil {
			writeOnlyBody := (*id.ResourceDef).GetWriteOnly(responseBody)
			if bodyMap, ok := writeOnlyBody.(map[string]interface{}); ok {
				delete(bodyMap, "location")
				delete(bodyMap, "tags")
				delete(bodyMap, "name")
				delete(bodyMap, "identity")
				writeOnlyBody = bodyMap
			}
			data, err := json.Marshal(writeOnlyBody)
			if err != nil {
				response.Diagnostics.AddError("Marshal Error", err.Error())
				return
			}
			model.Body = types.StringValue(string(data))
		}
		model.IgnoreCasing = types.BoolValue(false)
		model.IgnoreMissingProperty = types.BoolValue(true)
		model.SchemaValidationEnabled = types.BoolValue(true)
		model.RemovingSpecialChars = types.BoolValue(false)
	} else {
		option := utils.UpdateJsonOption{
			IgnoreCasing:          model.IgnoreCasing.ValueBool(),
			IgnoreMissingProperty: model.IgnoreMissingProperty.ValueBool(),
		}
		data, err := json.Marshal(utils.GetUpdatedJson(requestBody, responseBody, option))
		if err != nil {
			response.Diagnostics.AddError("Marshal Error", err.Error())
			return
		}
		model.Body = types.StringValue(string(data))
	}

	model.Name = types.StringValue(id.Name)
	model.ParentID = types.StringValue(id.ParentId)
	model.Type = types.StringValue(fmt.Sprintf("%s@%s", id.AzureResourceType, id.ApiVersion))

	if bodyMap, ok := responseBody.(map[string]interface{}); ok {
		model.Tags = tags.FlattenTags(bodyMap["tags"])
		if location.Normalize(model.Location.ValueString()) != location.Normalize(bodyMap["location"].(string)) {
			model.Location = types.StringValue(bodyMap["location"].(string))
		}
		model.Identity = identity.FlattenIdentity(bodyMap["identity"])
	}
	model.Output = types.StringValue(flattenOutput(responseBody, model.ResponseExportValues.Elements()))

	diags := response.State.Set(ctx, model)
	_ = diags
}

func (r *AzapiResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	client := r.ProviderData.ResourceClient

	var model *AzapiResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &model)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := parse.ResourceIDWithResourceType(model.ID.ValueString(), model.Type.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Error Parsing ID", err.Error())
		return
	}

	for _, element := range model.Locks.Elements() {
		lockId := element.(types.String).ValueString()
		locks.ByID(lockId)
		defer locks.UnlockByID(lockId)
	}

	_, err = client.Delete(ctx, id.AzureResourceId, id.ApiVersion)
	if err != nil {
		if utils.ResponseErrorWasNotFound(err) {
			return
		}
		response.Diagnostics.AddError("Delete Error", fmt.Errorf("deleting %q: %+v", id, err).Error())
	}
}

func (r *AzapiResource) expandBody(model AzapiResourceModel, body map[string]interface{}, resourceDef *azuretypes.ResourceType, diagnostics *diag.Diagnostics) map[string]interface{} {
	if !model.Tags.IsNull() {
		body["tags"] = tags.ExpandTags(model.Tags)
	} else if body["tags"] == nil && !r.ProviderData.Features.DefaultTags.IsNull() && isResourceHasProperty(resourceDef, "tags") {
		body["tags"] = tags.ExpandTags(r.ProviderData.Features.DefaultTags)
	}

	if !model.Location.IsNull() {
		body["location"] = location.Normalize(model.Location.ValueString())
	} else if body["location"] == nil && !r.ProviderData.Features.DefaultLocation.IsNull() && isResourceHasProperty(resourceDef, "location") {
		body["location"] = r.ProviderData.Features.DefaultLocation.String()
	}

	if !model.Identity.IsNull() {
		identityModel, err := identity.ExpandIdentity(model.Identity)
		if err != nil {
			diagnostics.AddError("Validation", err.Error())
		}
		if identityModel != nil {
			body["identity"] = identityModel
		}
	}
	return body
}

func validateDuplicatedDefinition(model AzapiResourceModel, body map[string]interface{}, diagnostics *diag.Diagnostics) {
	if !model.Identity.IsNull() && body["identity"] != nil {
		diagnostics.AddError("Validation", fmt.Errorf("can't specify both property `%[1]s` and `%[1]s` in `body`", "identity").Error())
	}
	if !model.Tags.IsNull() && body["tags"] != nil {
		diagnostics.AddError("Validation", fmt.Errorf("can't specify both property `%[1]s` and `%[1]s` in `body`", "tags").Error())
	}
	if !model.Location.IsNull() && body["location"] != nil {
		diagnostics.AddError("Validation", fmt.Errorf("can't specify both property `%[1]s` and `%[1]s` in `body`", "location").Error())
	}
}
