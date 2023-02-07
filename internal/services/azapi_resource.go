package services

import (
	"context"
	"github.com/Azure/terraform-provider-azapi/internal/azure/identity"
	"github.com/Azure/terraform-provider-azapi/internal/azure/location"
	"github.com/Azure/terraform-provider-azapi/internal/azure/tags"
	"github.com/Azure/terraform-provider-azapi/internal/clients"
	myplanmodifier "github.com/Azure/terraform-provider-azapi/internal/planmodifier"
	"github.com/Azure/terraform-provider-azapi/internal/services/parse"
	"github.com/Azure/terraform-provider-azapi/internal/services/validate"
	myValidator "github.com/Azure/terraform-provider-azapi/internal/validator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AzapiResource struct {
	ProviderData clients.Client
}

type azapiResourceData struct {
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

func (a AzapiResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	//if request.Plan
	//response
}

func (a AzapiResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*clients.Client); ok && v != nil {
		a.ProviderData = *v
	}
}

func (a AzapiResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_resource"
}

func (a AzapiResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
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

func (a AzapiResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data azapiResourceData
	if diags := request.Plan.Get(ctx, &data); diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}
	id, err := parse.NewResourceID(data.Name.ValueString(), data.ParentID.ValueString(), data.Type.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Error Parsing ID", err.Error())
		return
	}
	a.ProviderData.ResourceClient.Get(ctx, id.AzureResourceId, id.AzureResourceType)
}

func (a AzapiResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {

}

func (a AzapiResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {

}

func (a AzapiResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {

}

var _ resource.Resource = &AzapiResource{}
var _ resource.ResourceWithConfigure = &AzapiResource{}
var _ resource.ResourceWithModifyPlan = &AzapiResource{}
