// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/client"
)

// Ensure OAIProvider satisfies various provider interfaces.
var _ provider.Provider = &OAIProvider{}
var _ provider.ProviderWithFunctions = &OAIProvider{}

// OAIProvider defines the provider implementation.
type OAIProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// OAIProviderModel describes the provider data model.
type OAIProviderModel struct {
	Endpoint     types.String `tfsdk:"endpoint"`
	OfflineToken types.String `tfsdk:"offline_token"`
	Timeout      types.String `tfsdk:"timeout"`
}

func (p *OAIProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "oai"
	resp.Version = p.version
}

func (p *OAIProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "OpenShift Assisted Service API endpoint",
				Optional:            true,
			},
			"offline_token": schema.StringAttribute{
				MarkdownDescription: "Offline token for the Assisted Service API (from console.redhat.com). Will be exchanged for access tokens automatically.",
				Optional:            true,
				Sensitive:           true,
			},
			"timeout": schema.StringAttribute{
				MarkdownDescription: "Timeout for API requests (e.g., '30s', '5m')",
				Optional:            true,
			},
		},
	}
}

func (p *OAIProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data OAIProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Set default endpoint if not provided
	endpoint := "https://api.openshift.com/api/assisted-install"
	if !data.Endpoint.IsNull() {
		endpoint = data.Endpoint.ValueString()
	}

	// Get offline token from configuration or environment variable
	offlineToken := ""
	if !data.OfflineToken.IsNull() {
		offlineToken = data.OfflineToken.ValueString()
	} else {
		// Fall back to OFFLINE_TOKEN environment variable
		offlineToken = os.Getenv("OFFLINE_TOKEN")
	}

	// Parse timeout
	timeout := 30 * time.Second
	if !data.Timeout.IsNull() {
		if parsedTimeout, err := time.ParseDuration(data.Timeout.ValueString()); err == nil {
			timeout = parsedTimeout
		}
	}

	// Create OAI API client with OAuth2 support
	oaiClient := client.NewClient(client.ClientConfig{
		BaseURL:      endpoint,
		OfflineToken: offlineToken,
		Timeout:      timeout,
	})

	resp.DataSourceData = oaiClient
	resp.ResourceData = oaiClient
}

func (p *OAIProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewClusterResource,
		NewInfraEnvResource,
		NewHostResource,
		NewManifestResource,
	}
}


func (p *OAIProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewOpenShiftVersionsDataSource,
		NewSupportedOperatorsDataSource,
		NewOperatorBundlesDataSource,
		NewSupportLevelsDataSource,
	}
}

func (p *OAIProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		// No functions for OAI provider
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &OAIProvider{
			version: version,
		}
	}
}
