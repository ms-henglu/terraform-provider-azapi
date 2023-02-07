package provider

import (
	"context"
	"fmt"
	"github.com/Azure/terraform-provider-azapi/internal/services"
	"os"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/terraform-provider-azapi/internal/azure"
	"github.com/Azure/terraform-provider-azapi/internal/azure/location"
	"github.com/Azure/terraform-provider-azapi/internal/azure/tags"
	"github.com/Azure/terraform-provider-azapi/internal/clients"
	"github.com/Azure/terraform-provider-azapi/internal/features"
	myValidator "github.com/Azure/terraform-provider-azapi/internal/validator"
	"github.com/Azure/terraform-provider-azapi/version"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/meta"
)

func AzureProvider() provider.Provider {
	return &Provider{}
}

type Provider struct {
}

type providerData struct {
	SubscriptionID              types.String `tfsdk:"subscription_id"`
	ClientID                    types.String `tfsdk:"client_id"`
	TenantID                    types.String `tfsdk:"tenant_id"`
	Environment                 types.String `tfsdk:"environment"`
	ClientCertificatePath       types.String `tfsdk:"client_certificate_path"`
	ClientCertificatePassword   types.String `tfsdk:"client_certificate_password"`
	ClientSecret                types.String `tfsdk:"client_secret"`
	SkipProviderRegistration    types.Bool   `tfsdk:"skip_provider_registration"`
	OIDCRequestToken            types.String `tfsdk:"oidc_request_token"`
	OIDCRequestURL              types.String `tfsdk:"oidc_request_url"`
	OIDCToken                   types.String `tfsdk:"oidc_token"`
	OIDCTokenFilePath           types.String `tfsdk:"oidc_token_file_path"`
	UseOIDC                     types.Bool   `tfsdk:"use_oidc"`
	PartnerID                   types.String `tfsdk:"partner_id"`
	DisableCorrelationRequestID types.Bool   `tfsdk:"disable_correlation_request_id"`
	DisableTerraformPartnerID   types.Bool   `tfsdk:"disable_terraform_partner_id"`
	DefaultName                 types.String `tfsdk:"default_name"`
	DefaultNamingPrefix         types.String `tfsdk:"default_naming_prefix"`
	DefaultNamingSuffix         types.String `tfsdk:"default_naming_suffix"`
	DefaultLocation             types.String `tfsdk:"default_location"`
	DefaultTags                 types.Map    `tfsdk:"default_tags"`
}

func (p Provider) Metadata(ctx context.Context, request provider.MetadataRequest, response *provider.MetadataResponse) {
	response.TypeName = "azapi"
}

func (p Provider) Schema(ctx context.Context, request provider.SchemaRequest, response *provider.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "",
		Attributes: map[string]schema.Attribute{
			"subscription_id": schema.StringAttribute{
				Optional: true,
				//DefaultFunc: schema.EnvDefaultFunc("ARM_SUBSCRIPTION_ID", ""),
				Description: "The Subscription ID which should be used.",
			},

			"client_id": schema.StringAttribute{
				Optional: true,
				//DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_ID", ""),
				Description: "The Client ID which should be used.",
			},

			"tenant_id": schema.StringAttribute{
				Optional: true,
				//DefaultFunc: schema.EnvDefaultFunc("ARM_TENANT_ID", ""),
				Description: "The Tenant ID which should be used.",
			},

			// TODO@mgd: this is blocked by https://github.com/Azure/azure-sdk-for-go/issues/17159
			// "auxiliary_tenant_ids": {
			// 	Type:     schema.TypeList,
			// 	Optional: true,
			// 	MaxItems: 3,
			// 	Elem: &schema.Schema{
			// 		Type: schema.TypeString,
			// 	},
			// },

			"environment": schema.StringAttribute{
				Optional: true,
				//DefaultFunc:  schema.EnvDefaultFunc("ARM_ENVIRONMENT", "public"),
				Validators: []validator.String{
					stringvalidator.OneOfCaseInsensitive(
						"public",
						"usgovernment",
						"china",
					),
				},
				Description: "The Cloud Environment which should be used. Possible values are public, usgovernment and china. Defaults to public.",
			},

			// TODO@mgd: the metadata_host is used to retrieve metadata from Azure to identify current environment, this is used to eliminate Azure Stack usage, in which case the provider doesn't support.
			// "metadata_host": {
			// 	Type:        schema.TypeString,
			// 	Required:    true,
			// 	DefaultFunc: schema.EnvDefaultFunc("ARM_METADATA_HOSTNAME", ""),
			// 	Description: "The Hostname which should be used for the Azure Metadata Service.",
			// },

			// Client Certificate specific fields
			"client_certificate_path": schema.StringAttribute{
				Optional: true,
				//DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_CERTIFICATE_PATH", ""),
				Description: "The path to the Client Certificate associated with the Service Principal for use when authenticating as a Service Principal using a Client Certificate.",
			},

			"client_certificate_password": schema.StringAttribute{
				Optional: true,
				//DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_CERTIFICATE_PASSWORD", ""),
				Description: "The password associated with the Client Certificate. For use when authenticating as a Service Principal using a Client Certificate",
			},

			// Client Secret specific fields
			"client_secret": schema.StringAttribute{
				Optional: true,
				//DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_SECRET", ""),
				Description: "The Client Secret which should be used. For use When authenticating as a Service Principal using a Client Secret.",
			},

			"skip_provider_registration": schema.BoolAttribute{
				Optional: true,
				//DefaultFunc: schema.EnvDefaultFunc("ARM_SKIP_PROVIDER_REGISTRATION", false),
				Description: "Should the Provider skip registering all of the Resource Providers that it supports, if they're not already registered?",
			},

			// OIDC specific fields
			"oidc_request_token": schema.StringAttribute{
				Optional: true,
				//DefaultFunc: schema.MultiEnvDefaultFunc([]string{"ARM_OIDC_REQUEST_TOKEN", "ACTIONS_ID_TOKEN_REQUEST_TOKEN"}, ""),
				Description: "The bearer token for the request to the OIDC provider. For use When authenticating as a Service Principal using OpenID Connect.",
			},

			"oidc_request_url": schema.StringAttribute{
				Optional: true,
				//DefaultFunc: schema.MultiEnvDefaultFunc([]string{"ARM_OIDC_REQUEST_URL", "ACTIONS_ID_TOKEN_REQUEST_URL"}, ""),
				Description: "The URL for the OIDC provider from which to request an ID token. For use When authenticating as a Service Principal using OpenID Connect.",
			},

			"oidc_token": schema.StringAttribute{
				Optional: true,
				//DefaultFunc: schema.EnvDefaultFunc("ARM_OIDC_TOKEN", ""),
				Description: "The OIDC ID token for use when authenticating as a Service Principal using OpenID Connect.",
			},

			"oidc_token_file_path": schema.StringAttribute{
				Optional: true,
				//DefaultFunc: schema.EnvDefaultFunc("ARM_OIDC_TOKEN_FILE_PATH", ""),
				Description: "The path to a file containing an OIDC ID token for use when authenticating as a Service Principal using OpenID Connect.",
			},

			"use_oidc": schema.BoolAttribute{
				Optional: true,
				//DefaultFunc: schema.EnvDefaultFunc("ARM_USE_OIDC", false),
				Description: "Allow OpenID Connect to be used for authentication",
			},

			// TODO@mgd: azidentity doesn't support msi_endpoint
			// // Managed Service Identity specific fields
			// "use_msi": {
			// 	Type:        schema.TypeBool,
			// 	Optional:    true,
			// 	DefaultFunc: schema.EnvDefaultFunc("ARM_USE_MSI", false),
			// 	Description: "Allowed Managed Service Identity be used for Authentication.",
			// },
			// "msi_endpoint": {
			// 	Type:        schema.TypeString,
			// 	Optional:    true,
			// 	DefaultFunc: schema.EnvDefaultFunc("ARM_MSI_ENDPOINT", ""),
			// 	Description: "The path to a custom endpoint for Managed Service Identity - in most circumstances this should be detected automatically. ",
			// },

			// Managed Tracking GUID for User-agent
			"partner_id": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.Any(
						myValidator.StringIsUUID(),
						myValidator.StringIsEmpty(),
					),
				},
				//DefaultFunc:  schema.EnvDefaultFunc("ARM_PARTNER_ID", ""),
				Description: "A GUID/UUID that is registered with Microsoft to facilitate partner resource usage attribution.",
			},

			"disable_correlation_request_id": schema.BoolAttribute{
				Optional: true,
				//DefaultFunc: schema.EnvDefaultFunc("ARM_DISABLE_CORRELATION_REQUEST_ID", false),
				Description: "This will disable the x-ms-correlation-request-id header.",
			},

			"disable_terraform_partner_id": schema.BoolAttribute{
				Optional: true,
				//DefaultFunc: schema.EnvDefaultFunc("ARM_DISABLE_TERRAFORM_PARTNER_ID", false),
				Description: "This will disable the Terraform Partner ID which is used if a custom `partner_id` isn't specified.",
			},

			"default_name": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						path.MatchRoot("default_naming_prefix"),
						path.MatchRoot("default_naming_suffix"),
					),
				},
			},

			"default_naming_prefix": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						path.MatchRoot("default_name"),
					),
				},
			},

			"default_naming_suffix": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						path.MatchRoot("default_name"),
					),
				},
			},

			"default_location": schema.StringAttribute{
				Optional:   true,
				Validators: []validator.String{myValidator.StringIsNotEmpty()},
			},

			"default_tags": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.Map{
					tags.Validator(),
				},
			},
		},
	}
}

func (p Provider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
	var config providerData
	diags := request.Config.Get(ctx, &config)
	response.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	// set the defaults from environment variables
	if config.SubscriptionID.IsNull() {
		if v := os.Getenv("ARM_SUBSCRIPTION_ID"); v != "" {
			config.SubscriptionID = types.StringValue(v)
		}
	}
	if config.ClientID.IsNull() {
		if v := os.Getenv("ARM_CLIENT_ID"); v != "" {
			config.ClientID = types.StringValue(v)
		}
	}
	if config.TenantID.IsNull() {
		if v := os.Getenv("ARM_TENANT_ID"); v != "" {
			config.TenantID = types.StringValue(v)
		}
	}
	if config.Environment.IsNull() {
		if v := os.Getenv("ARM_ENVIRONMENT"); v != "" {
			config.Environment = types.StringValue(v)
		} else {
			config.Environment = types.StringValue("public")
		}
	}
	if config.ClientCertificatePath.IsNull() {
		if v := os.Getenv("ARM_CLIENT_CERTIFICATE_PATH"); v != "" {
			config.ClientCertificatePath = types.StringValue(v)
		}
	}
	if config.ClientCertificatePassword.IsNull() {
		if v := os.Getenv("ARM_CLIENT_CERTIFICATE_PASSWORD"); v != "" {
			config.ClientCertificatePassword = types.StringValue(v)
		}
	}
	if config.ClientSecret.IsNull() {
		if v := os.Getenv("ARM_CLIENT_SECRET"); v != "" {
			config.ClientSecret = types.StringValue(v)
		}
	}
	if config.SkipProviderRegistration.IsNull() {
		if v := os.Getenv("ARM_SKIP_PROVIDER_REGISTRATION"); v != "" {
			config.SkipProviderRegistration = types.BoolValue(strings.EqualFold(v, "true"))
		} else {
			config.SkipProviderRegistration = types.BoolValue(false)
		}
	}
	if config.OIDCRequestToken.IsNull() {
		if v := os.Getenv("ARM_OIDC_REQUEST_TOKEN"); v != "" {
			config.OIDCRequestToken = types.StringValue(v)
		} else if v := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN"); v != "" {
			config.OIDCRequestToken = types.StringValue(v)
		}
	}
	if config.OIDCRequestURL.IsNull() {
		if v := os.Getenv("ARM_OIDC_REQUEST_URL"); v != "" {
			config.OIDCRequestURL = types.StringValue(v)
		} else if v := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_URL"); v != "" {
			config.OIDCRequestURL = types.StringValue(v)
		}
	}
	if config.OIDCToken.IsNull() {
		if v := os.Getenv("ARM_OIDC_TOKEN"); v != "" {
			config.OIDCToken = types.StringValue(v)
		}
	}
	if config.UseOIDC.IsNull() {
		if v := os.Getenv("ARM_USE_OIDC"); v != "" {
			config.UseOIDC = types.BoolValue(strings.EqualFold(v, "true"))
		} else {
			config.UseOIDC = types.BoolValue(false)
		}
	}
	if config.PartnerID.IsNull() {
		if v := os.Getenv("ARM_PARTNER_ID"); v != "" {
			config.PartnerID = types.StringValue(v)
		}
	}
	if config.DisableCorrelationRequestID.IsNull() {
		if v := os.Getenv("ARM_DISABLE_CORRELATION_REQUEST_ID"); v != "" {
			config.DisableCorrelationRequestID = types.BoolValue(strings.EqualFold(v, "true"))
		} else {
			config.DisableCorrelationRequestID = types.BoolValue(false)
		}
	}
	if config.DisableTerraformPartnerID.IsNull() {
		if v := os.Getenv("ARM_DISABLE_TERRAFORM_PARTNER_ID"); v != "" {
			config.DisableTerraformPartnerID = types.BoolValue(strings.EqualFold(v, "true"))
		} else {
			config.DisableTerraformPartnerID = types.BoolValue(false)
		}
	}

	// var auxTenants []string
	// if v, ok := d.Get("auxiliary_tenant_ids").([]interface{}); ok && len(v) > 0 {
	// 	auxTenants = *utils.ExpandStringSlice(v)
	// } else if v := os.Getenv("ARM_AUXILIARY_TENANT_IDS"); v != "" {
	// 	auxTenants = strings.Split(v, ";")
	// }

	var cloudConfig cloud.Configuration
	env := config.Environment.ValueString()
	switch strings.ToLower(env) {
	case "public":
		cloudConfig = cloud.AzurePublic
	case "usgovernment":
		cloudConfig = cloud.AzureGovernment
	case "china":
		cloudConfig = cloud.AzureChina
	default:
		response.Diagnostics.AddError("Invalid `environment`", fmt.Sprintf("unknown `environment` specified: %q", env))
		return
	}

	// Maps the auth related environment variables used in the provider to what azidentity honors.
	if v := config.TenantID.ValueString(); len(v) != 0 {
		// #nosec G104
		os.Setenv("AZURE_TENANT_ID", v)
	}
	if v := config.ClientID.ValueString(); len(v) != 0 {
		// #nosec G104
		os.Setenv("AZURE_CLIENT_ID", v)
	}
	if v := config.ClientSecret.ValueString(); len(v) != 0 {
		// #nosec G104
		os.Setenv("AZURE_CLIENT_SECRET", v)
	}
	if v := config.ClientCertificatePath.ValueString(); len(v) != 0 {
		// #nosec G104
		os.Setenv("AZURE_CLIENT_CERTIFICATE_PATH", v)
	}
	if v := config.ClientCertificatePassword.ValueString(); len(v) != 0 {
		// #nosec G104
		os.Setenv("AZURE_CLIENT_CERTIFICATE_PASSWORD", v)
	}

	cred, err := getCredential(config, cloudConfig)
	if err != nil {
		response.Diagnostics.AddError("Failed to Obtain a Credential", fmt.Sprintf("failed to obtain a credential: %v", err))
		return
	}

	copt := &clients.Option{
		SubscriptionId:       config.SubscriptionID.ValueString(),
		Cred:                 cred,
		CloudCfg:             cloudConfig,
		ApplicationUserAgent: buildUserAgent(request.TerraformVersion, config.PartnerID.ValueString(), config.DisableTerraformPartnerID.ValueBool()),
		Features: features.UserFeatures{
			DefaultTags:         tags.ExpandTags2(config.DefaultTags),
			DefaultLocation:     location.Normalize(config.DefaultName.ValueString()),
			DefaultNaming:       config.DefaultName.ValueString(),
			DefaultNamingPrefix: config.DefaultNamingPrefix.ValueString(),
			DefaultNamingSuffix: config.DefaultNamingSuffix.ValueString(),
		},
		SkipProviderRegistration:    config.SkipProviderRegistration.ValueBool(),
		DisableCorrelationRequestID: config.DisableCorrelationRequestID.ValueBool(),
	}

	client := &clients.Client{}
	if err := client.Build(ctx, copt); err != nil {
		response.Diagnostics.AddError("Error Building Client", err.Error())
		return
	}

	// load schema
	var mutex sync.Mutex
	mutex.Lock()
	azure.GetAzureSchema()
	mutex.Unlock()

	response.ResourceData = client
	response.DataSourceData = client
}

func (p Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		//func() datasource.DataSource {
		//	return &DataSource{}
		//},
	}
}

func (p Provider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource {
			return &services.AzapiResource{}
		},
	}
}

func buildUserAgent(terraformVersion string, partnerID string, disableTerraformPartnerID bool) string {
	if terraformVersion == "" {
		// Terraform 0.12 introduced this field to the protocol
		// We can therefore assume that if it's missing it's 0.10 or 0.11
		terraformVersion = "0.11+compatible"
	}

	tfUserAgent := fmt.Sprintf("HashiCorp Terraform/%s (+https://www.terraform.io) Terraform Plugin SDK/%s", terraformVersion, meta.SDKVersionString())
	providerUserAgent := fmt.Sprintf("terraform-provider-azapi/%s", version.ProviderVersion)
	userAgent := strings.TrimSpace(fmt.Sprintf("%s %s", tfUserAgent, providerUserAgent))

	// append the CloudShell version to the user agent if it exists
	if azureAgent := os.Getenv("AZURE_HTTP_USER_AGENT"); azureAgent != "" {
		userAgent = fmt.Sprintf("%s %s", userAgent, azureAgent)
	}

	// only one pid can be interpreted currently
	// hence, send partner ID if present, otherwise send Terraform GUID
	// unless users have opted out
	if partnerID == "" && !disableTerraformPartnerID {
		// Microsoft’s Terraform Partner ID is this specific GUID
		partnerID = "222c6c49-1b0a-5959-a213-6608f9eb8820"
	}

	if partnerID != "" {
		userAgent = fmt.Sprintf("%s pid-%s", userAgent, partnerID)
	}
	return userAgent
}

func getCredential(config providerData, cloudConfig cloud.Configuration) (azcore.TokenCredential, error) {
	if config.UseOIDC.ValueBool() {
		return NewOidcCredential(&OidcCredentialOptions{
			ClientOptions: azcore.ClientOptions{
				Cloud: cloudConfig,
			},
			TenantID:      config.TenantID.ValueString(),
			ClientID:      config.ClientID.ValueString(),
			RequestToken:  config.OIDCRequestToken.ValueString(),
			RequestUrl:    config.OIDCRequestURL.ValueString(),
			Token:         config.OIDCToken.ValueString(),
			TokenFilePath: config.OIDCTokenFilePath.ValueString(),
		})
	}

	return azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{
		ClientOptions: azcore.ClientOptions{
			Cloud: cloudConfig,
		},
		TenantID: config.TenantID.ValueString(),
	})
}
