package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/terraform-provider-azapi/internal/azure/identity"
	"github.com/Azure/terraform-provider-azapi/internal/azure/location"
	"github.com/Azure/terraform-provider-azapi/internal/azure/tags"
	"github.com/Azure/terraform-provider-azapi/internal/clients"
	"github.com/Azure/terraform-provider-azapi/internal/locks"
	myplanmodifier "github.com/Azure/terraform-provider-azapi/internal/planmodifier"
	"github.com/Azure/terraform-provider-azapi/internal/services/parse"
	"github.com/Azure/terraform-provider-azapi/internal/services/validate"
	"github.com/Azure/terraform-provider-azapi/internal/tf"
	myValidator "github.com/Azure/terraform-provider-azapi/internal/validator"
	"github.com/Azure/terraform-provider-azapi/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"log"
)

type AzapiResource struct {
	ProviderData *clients.Client
}

type AzapiResourceModel struct {
	ID                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	ParentID                types.String `tfsdk:"parent_id"`
	Type                    types.String `tfsdk:"type"`
	Location                types.String `tfsdk:"location"`
	Identity                types.Object `tfsdk:"identity"`
	Body                    types.String `tfsdk:"body"`
	IgnoreCasing            types.Bool   `tfsdk:"ignore_casing"`
	IgnoreMissingProperty   types.Bool   `tfsdk:"ignore_missing_property"`
	ResponseExportValues    types.List   `tfsdk:"response_export_values"`
	Locks                   types.List   `tfsdk:"locks"`
	RemovingSpecialChars    types.Bool   `tfsdk:"removing_special_chars"`
	SchemaValidationEnabled types.Bool   `tfsdk:"schema_validation_enabled"`
	Output                  types.String `tfsdk:"output"`
	Tags                    types.Map    `tfsdk:"tags"`
}

func (r *AzapiResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	//if request.Plan
	//response
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
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					location.NormalizeLocation(),
				},
			},

			"identity": identity.SchemaIdentity(),

			"body": schema.StringAttribute{
				Optional: true,
				//Default:          "{}",
				//DiffSuppressFunc: tf.SuppressJsonOrderingDifference,
				PlanModifiers: []planmodifier.String{},
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
	}
}

func (r *AzapiResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	client := r.ProviderData.ResourceClient
	var model *AzapiResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &model)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := parse.NewResourceID(model.Name.ValueString(), model.ParentID.ValueString(), model.Type.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Error Parsing ID", err.Error())
		return
	}

	_, err = client.Get(ctx, id.AzureResourceId, id.ApiVersion)
	if err == nil {
		response.Diagnostics.AddError("Import As Exists Error", tf.ImportAsExistsError("azapi_resource", id.ID()).Error())
		return
	}
	if !utils.ResponseErrorWasNotFound(err) {
		response.Diagnostics.AddError("Reading Resource", fmt.Errorf("checking for presence of existing %s: %+v", id, err).Error())
		return
	}

	var body map[string]interface{}
	err = json.Unmarshal([]byte(model.Body.ValueString()), &body)
	if err != nil {
		response.Diagnostics.AddError("JSON Unmarshal Error", err.Error())
		return
	}

	if !model.Tags.IsNull() {
		tagsModel := tags.ExpandTags2(model.Tags)
		if len(tagsModel) != 0 {
			body["tags"] = tagsModel
		}
	}
	if !model.Location.IsNull() {
		body["location"] = location.Normalize(model.Location.ValueString())
	}
	if !model.Identity.IsNull() {
		identityModel, err := identity.ExpandIdentity2(model.Identity)
		if err != nil {
			response.Diagnostics.AddError("Expanding Identity Error", err.Error())
			return
		}
		if identityModel != nil {
			body["identity"] = identityModel
		}
	}

	if model.SchemaValidationEnabled.ValueBool() {
		if err := schemaValidation(id.AzureResourceType, id.ApiVersion, id.ResourceDef, body); err != nil {
			response.Diagnostics.AddError("Schema Validation Error", err.Error())
			return
		}
	}

	j, _ := json.Marshal(body)
	log.Printf("[INFO] request body: %v\n", string(j))

	for _, element := range model.Locks.Elements() {
		lockId := element.(types.String).ValueString()
		locks.ByID(lockId)
		defer locks.UnlockByID(lockId)
	}

	_, err = client.CreateOrUpdate(ctx, id.AzureResourceId, id.ApiVersion, body)
	if err != nil {
		response.Diagnostics.AddError("Create Error", fmt.Errorf("creating/updating %q: %+v", id, err).Error())
	}

	model.ID = types.StringValue(id.ID())
	diags := response.State.Set(ctx, model)
	response.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	rreq := resource.ReadRequest{
		State:        response.State,
		ProviderMeta: request.ProviderMeta,
	}
	rresp := resource.ReadResponse{
		State:       response.State,
		Diagnostics: response.Diagnostics,
	}
	r.Read(ctx, rreq, &rresp)

	*response = resource.CreateResponse{
		State:       rresp.State,
		Diagnostics: rresp.Diagnostics,
	}
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

	state := &AzapiResourceModel{}
	responseBody, err := client.Get(ctx, id.AzureResourceId, id.ApiVersion)
	if err != nil {
		if utils.ResponseErrorWasNotFound(err) {
			log.Printf("[INFO] Error reading %q - removing from state", id.ID())
			state.ID = types.StringNull()
			response.State.Set(ctx, state)
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
			state.Body = types.StringValue(string(data))
		}
		state.IgnoreCasing = types.BoolValue(false)
		state.IgnoreMissingProperty = types.BoolValue(true)
		state.SchemaValidationEnabled = types.BoolValue(true)
		state.RemovingSpecialChars = types.BoolValue(false)
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
		state.Body = types.StringValue(string(data))
	}

	state.Name = types.StringValue(id.Name)
	state.ParentID = types.StringValue(id.ParentId)
	state.Type = types.StringValue(fmt.Sprintf("%s@%s", id.AzureResourceType, id.ApiVersion))

	if bodyMap, ok := responseBody.(map[string]interface{}); ok {
		state.Tags = types.MapValueMust(types.StringType, tags.FlattenTags(bodyMap["tags"]))
		state.Location = types.StringValue(bodyMap["location"].(string))
		state.Identity = types.ObjectValueMust(identity.AttributeTypes(), identity.FlattenIdentity(bodyMap["identity"]))
	}

	state.Output = types.StringValue(flattenOutput(responseBody, model.ResponseExportValues.Elements()))

	response.State.Set(ctx, state)
}

func (r *AzapiResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	client := r.ProviderData.ResourceClient
	var model *AzapiResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &model)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := parse.NewResourceID(model.Name.ValueString(), model.ParentID.ValueString(), model.Type.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Error Parsing ID", err.Error())
		return
	}

	var body map[string]interface{}
	err = json.Unmarshal([]byte(model.Body.ValueString()), &body)
	if err != nil {
		response.Diagnostics.AddError("JSON Unmarshal Error", err.Error())
		return
	}

	if !model.Tags.IsNull() {
		tagsModel := tags.ExpandTags2(model.Tags)
		if len(tagsModel) != 0 {
			body["tags"] = tagsModel
		}
	}
	if !model.Location.IsNull() {
		body["location"] = location.Normalize(model.Location.ValueString())
	}
	if !model.Identity.IsNull() {
		identityModel, err := identity.ExpandIdentity2(model.Identity)
		if err != nil {
			response.Diagnostics.AddError("Expanding Identity Error", err.Error())
			return
		}
		if identityModel != nil {
			body["identity"] = identityModel
		}
	}

	if model.SchemaValidationEnabled.ValueBool() {
		if err := schemaValidation(id.AzureResourceType, id.ApiVersion, id.ResourceDef, body); err != nil {
			response.Diagnostics.AddError("Schema Validation Error", err.Error())
			return
		}
	}

	j, _ := json.Marshal(body)
	log.Printf("[INFO] request body: %v\n", string(j))

	for _, element := range model.Locks.Elements() {
		lockId := element.(types.String).ValueString()
		locks.ByID(lockId)
		defer locks.UnlockByID(lockId)
	}

	_, err = client.CreateOrUpdate(ctx, id.AzureResourceId, id.ApiVersion, body)
	if err != nil {
		response.Diagnostics.AddError("Create Error", fmt.Errorf("creating/updating %q: %+v", id, err).Error())
	}

	model.ID = types.StringValue(id.ID())
	diags := response.State.Set(ctx, model)
	response.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	rreq := resource.ReadRequest{
		State:        response.State,
		ProviderMeta: request.ProviderMeta,
	}
	rresp := resource.ReadResponse{
		State:       response.State,
		Diagnostics: response.Diagnostics,
	}
	r.Read(ctx, rreq, &rresp)

	*response = resource.UpdateResponse{
		State:       rresp.State,
		Diagnostics: rresp.Diagnostics,
	}
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

var _ resource.Resource = &AzapiResource{}
var _ resource.ResourceWithConfigure = &AzapiResource{}
var _ resource.ResourceWithModifyPlan = &AzapiResource{}
