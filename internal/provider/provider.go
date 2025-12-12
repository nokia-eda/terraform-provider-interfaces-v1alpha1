package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/nokia/eda/apps/terraform-provider-interfaces/internal/eda/apiclient"
	"github.com/nokia/eda/apps/terraform-provider-interfaces/internal/eda/utils"
	"github.com/nokia/eda/apps/terraform-provider-interfaces/internal/tfutils"
)

const (
	// Environment variables
	ENV_EDA_BASE_URL        = "BASE_URL"
	ENV_KC_REALM            = "KEYCLOAK_MASTER_REALM"
	ENV_KC_CLIENT_ID        = "KEYCLOAK_ADMIN_CLIENT_ID"
	ENV_KC_USERNAME         = "KEYCLOAK_ADMIN_USERNAME"
	ENV_KC_PASSWORD         = "KEYCLOAK_ADMIN_PASSWORD"
	ENV_EDA_CLIENT_ID       = "CLIENT_ID"
	ENV_EDA_CLIENT_SECRET   = "CLIENT_SECRET"
	ENV_EDA_REALM           = "REALM"
	ENV_EDA_USERNAME        = "USERNAME"
	ENV_EDA_PASSWORD        = "PASSWORD"
	ENV_TLS_SKIP_VERIFY     = "TLS_SKIP_VERIFY"
	ENV_REST_DEBUG          = "REST_DEBUG"
	ENV_REST_TIMEOUT        = "REST_TIMEOUT"
	ENV_REST_RETRIES        = "REST_RETRIES"
	ENV_REST_RETRY_INTERVAL = "REST_RETRY_INTERVAL"

	// Default values
	DEF_KC_REALM            = "master"
	DEF_KC_CLIENT_ID        = "admin-cli"
	DEF_EDA_REALM           = "eda"
	DEF_EDA_CLIENT_ID       = "eda"
	DEF_USERNAME            = "admin"
	DEF_PASSWORD            = "admin"
	DEF_REST_TIMEOUT        = 15 * time.Second
	DEF_REST_RETRIES        = 3
	DEF_REST_RETRY_INTERVAL = 5 * time.Second
)

var _ provider.Provider = (*interfacesProvider)(nil)

func New(ver string) func() provider.Provider {
	return func() provider.Provider {
		return &interfacesProvider{version: ver}
	}
}

type interfacesProvider struct {
	version string
}

type providerModel struct {
	BaseURL           types.String `tfsdk:"base_url"`
	KcRealm           types.String `tfsdk:"keycloak_master_realm"`
	KcClientID        types.String `tfsdk:"keycloak_admin_client_id"`
	KcUsername        types.String `tfsdk:"keycloak_admin_username"`
	KcPassword        types.String `tfsdk:"keycloak_admin_password"`
	EdaRealm          types.String `tfsdk:"realm"`
	EdaClientID       types.String `tfsdk:"client_id"`
	EdaClientSecret   types.String `tfsdk:"client_secret"`
	EdaUsername       types.String `tfsdk:"username"`
	EdaPassword       types.String `tfsdk:"password"`
	TlsSkipVerify     types.Bool   `tfsdk:"tls_skip_verify"`
	RestDebug         types.Bool   `tfsdk:"rest_debug"`
	RestTimeout       types.String `tfsdk:"rest_timeout"`
	RestRetries       types.Int64  `tfsdk:"rest_retries"`
	RestRetryInterval types.String `tfsdk:"rest_retry_interval"`
}

func (p *interfacesProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Description: "Base URL",
				Optional:    true,
			},
			"keycloak_master_realm": schema.StringAttribute{
				Description: "Keycloak Realm",
				Optional:    true,
			},
			"keycloak_admin_client_id": schema.StringAttribute{
				Description: "Keycloak Client ID",
				Optional:    true,
			},
			"keycloak_admin_username": schema.StringAttribute{
				Description: "Keycloak Username",
				Optional:    true,
			},
			"keycloak_admin_password": schema.StringAttribute{
				Description: "Keycloak Password",
				Optional:    true,
				Sensitive:   true,
			},
			"realm": schema.StringAttribute{
				Description: "EDA Realm",
				Optional:    true,
			},
			"client_id": schema.StringAttribute{
				Description: "EDA Client ID",
				Optional:    true,
			},
			"client_secret": schema.StringAttribute{
				Description: "EDA Client Secret",
				Optional:    true,
			},
			"username": schema.StringAttribute{
				Description: "EDA Username",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "EDA Password",
				Optional:    true,
				Sensitive:   true,
			},
			"tls_skip_verify": schema.BoolAttribute{
				Description: "TLS skip verify",
				Optional:    true,
			},
			"rest_debug": schema.BoolAttribute{
				Description: "REST Debug",
				Optional:    true,
			},
			"rest_timeout": schema.StringAttribute{
				Description: "REST Timeout",
				Optional:    true,
			},
			"rest_retries": schema.Int64Attribute{
				Description: "REST Retries",
				Optional:    true,
			},
			"rest_retry_interval": schema.StringAttribute{
				Description: "REST Retry Interval",
				Optional:    true,
			},
		},
	}
}

func (p *interfacesProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	anyData, err := tfutils.ModelToAnyMap(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Error reading provider config", err.Error())
		return
	}

	config := apiclient.Config{}
	err = utils.Convert(anyData, &config)
	if err != nil {
		resp.Diagnostics.AddError("Config data conversion error", err.Error())
		return
	}

	validate(&resp.Diagnostics, &config)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Info(ctx, "Configure()::Provider config", map[string]any{"config": config.String()})

	// Create a new EDA ApiService client using the configuration values
	client, err := apiclient.NewEdaApiClient(ctx, &config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create EDA API Client",
			"An unexpected error occurred when creating the EDA API service client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"EDA API Client Error: "+err.Error(),
		)
		return
	}
	// Make the EDA API client available during DataSource and Resource type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured EDA API client", map[string]any{"success": true})
}

func validate(diags *diag.Diagnostics, cfg *apiclient.Config) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = utils.GetEnvWithDefault(ENV_EDA_BASE_URL, "")
	}
	if cfg.BaseURL == "" {
		diags.AddAttributeError(
			path.Root("base_url"), "Unknown EDA Base URL",
			"The provider cannot create the EDA API client as there is an unknown configuration value for the EDA Base URL. "+
				"Either set the value statically in the configuration, or use the EDA_BASE_URL environment variable.")
	}
	if cfg.KcUsername == "" {
		cfg.KcUsername = utils.GetEnvWithDefault(ENV_KC_USERNAME, DEF_USERNAME)
	}
	if cfg.KcPassword == "" {
		cfg.KcPassword = utils.GetEnvWithDefault(ENV_KC_PASSWORD, DEF_PASSWORD)
	}
	if cfg.KcRealm == "" {
		cfg.KcRealm = utils.GetEnvWithDefault(ENV_KC_REALM, DEF_KC_REALM)
	}
	if cfg.KcClientID == "" {
		cfg.KcClientID = utils.GetEnvWithDefault(ENV_KC_CLIENT_ID, DEF_KC_CLIENT_ID)
	}
	if cfg.EdaUsername == "" {
		cfg.EdaUsername = utils.GetEnvWithDefault(ENV_EDA_USERNAME, DEF_USERNAME)
	}
	if cfg.EdaPassword == "" {
		cfg.EdaPassword = utils.GetEnvWithDefault(ENV_EDA_PASSWORD, DEF_PASSWORD)
	}
	if cfg.EdaRealm == "" {
		cfg.EdaRealm = utils.GetEnvWithDefault(ENV_EDA_REALM, DEF_EDA_REALM)
	}
	if cfg.EdaClientID == "" {
		cfg.EdaClientID = utils.GetEnvWithDefault(ENV_EDA_CLIENT_ID, DEF_EDA_CLIENT_ID)
	}
	if cfg.EdaClientSecret == "" {
		cfg.EdaClientSecret = utils.GetEnvWithDefault(ENV_EDA_CLIENT_SECRET, "")
	}
	if cfg.TlsSkipVerify == false {
		cfg.TlsSkipVerify = utils.GetEnvBoolWithDefault(ENV_TLS_SKIP_VERIFY, false)
	}
	if cfg.RestDebug == false {
		cfg.RestDebug = utils.GetEnvBoolWithDefault(ENV_REST_DEBUG, false)
	}
	if cfg.RestTimeout == 0*time.Second {
		cfg.RestTimeout = utils.GetEnvDurationWithDefault(ENV_REST_TIMEOUT, DEF_REST_TIMEOUT)
	}
	if cfg.RestRetries == 0 {
		cfg.RestRetries = utils.GetEnvIntWithDefault(ENV_REST_RETRIES, DEF_REST_RETRIES)
	}
	if cfg.RestRetryInterval == 0*time.Second {
		cfg.RestRetryInterval = utils.GetEnvDurationWithDefault(ENV_REST_RETRY_INTERVAL, DEF_REST_RETRY_INTERVAL)
	}
}

func (p *interfacesProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "interfaces-v1alpha1"
	resp.Version = p.version
}

func (p *interfacesProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAppGroupDataSource,
		NewBreakoutDataSource,
		NewBreakoutListDataSource,
		NewCheckInterfacesDataSource,
		NewCheckInterfacesListDataSource,
		NewInterfaceDataSource,
		NewInterfaceListDataSource,
		NewResourceListDataSource,
		NewWorkflowGetInputsRespElemDataSource,
	}
}

func (p *interfacesProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewBreakoutResource,
		NewCheckInterfacesResource,
		NewInterfaceResource,
	}
}
