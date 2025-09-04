package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/benemon/terraform-provider-openshift-assisted-installer/internal/client"
)

var _ resource.Resource = &ClusterInstallationResource{}

func NewClusterInstallationResource() resource.Resource {
	return &ClusterInstallationResource{}
}

type ClusterInstallationResource struct {
	client *client.Client
}

type ClusterInstallationResourceModel struct {
	Timeouts          timeouts.Value `tfsdk:"timeouts"`
	ID                types.String   `tfsdk:"id"`
	ClusterID         types.String   `tfsdk:"cluster_id"`
	WaitForHosts      types.Bool     `tfsdk:"wait_for_hosts"`
	ExpectedHostCount types.Int64    `tfsdk:"expected_host_count"`
	Status            types.String   `tfsdk:"status"`
	StatusInfo        types.String   `tfsdk:"status_info"`
	InstallStartedAt  types.String   `tfsdk:"install_started_at"`
	InstallCompletedAt types.String  `tfsdk:"install_completed_at"`
}

func (r *ClusterInstallationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_installation"
}

func (r *ClusterInstallationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages the installation of an OpenShift cluster created through the Assisted Installer.

This resource triggers and monitors the cluster installation process. It should be used after:
1. The cluster resource has been created
2. The infra-env resource has been created
3. Hosts have been discovered from the ISO and validated

Example usage with separate modules:
- Module 1: Create cluster and infra-env, output ISO URL
- Module 2: After hosts are ready, use this resource to trigger installation`,

		Attributes: map[string]schema.Attribute{
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
			}),
			"id": schema.StringAttribute{
				MarkdownDescription: "Installation resource ID (same as cluster_id)",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the cluster to install",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"wait_for_hosts": schema.BoolAttribute{
				MarkdownDescription: "Whether to wait for expected number of hosts before triggering installation. Defaults to true.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"expected_host_count": schema.Int64Attribute{
				MarkdownDescription: "Number of hosts expected to be discovered before installation can begin. Required if wait_for_hosts is true. Defaults to 3 for multi-node clusters.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(3),
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Current installation status",
				Computed:            true,
			},
			"status_info": schema.StringAttribute{
				MarkdownDescription: "Detailed status information",
				Computed:            true,
			},
			"install_started_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when installation was triggered",
				Computed:            true,
			},
			"install_completed_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when installation completed",
				Computed:            true,
			},
		},
	}
}

func (r *ClusterInstallationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *ClusterInstallationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ClusterInstallationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := data.Timeouts.Create(ctx, 90*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	clusterID := data.ClusterID.ValueString()
	
	// Set the ID immediately (same as cluster ID for this resource)
	data.ID = types.StringValue(clusterID)
	
	tflog.Info(ctx, "Starting cluster installation", map[string]interface{}{
		"cluster_id": clusterID,
	})

	// Get current cluster state
	cluster, err := r.client.GetCluster(ctx, clusterID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving cluster",
			fmt.Sprintf("Could not get cluster %s: %s", clusterID, err),
		)
		return
	}

	// Check if already installing or installed
	if cluster.Status == "installed" {
		tflog.Info(ctx, "Cluster already installed", map[string]interface{}{
			"cluster_id": clusterID,
		})
		data.Status = types.StringValue(cluster.Status)
		data.StatusInfo = types.StringValue(cluster.StatusInfo)
		data.InstallCompletedAt = types.StringValue(time.Now().UTC().Format(time.RFC3339))
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	if cluster.Status == "installing" || cluster.Status == "finalizing" {
		tflog.Info(ctx, "Cluster installation already in progress", map[string]interface{}{
			"cluster_id": clusterID,
			"status": cluster.Status,
		})
	} else {
		// Wait for hosts if requested
		if data.WaitForHosts.ValueBool() {
			expectedHosts := int(data.ExpectedHostCount.ValueInt64())
			tflog.Info(ctx, "Waiting for hosts to be ready", map[string]interface{}{
				"cluster_id": clusterID,
				"expected_hosts": expectedHosts,
			})

			err = r.waitForClusterReady(ctx, clusterID, expectedHosts)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error waiting for cluster to be ready",
					fmt.Sprintf("Cluster %s did not become ready for installation: %s", clusterID, err),
				)
				return
			}
		}

		// Trigger installation
		tflog.Info(ctx, "Triggering cluster installation", map[string]interface{}{
			"cluster_id": clusterID,
		})

		data.InstallStartedAt = types.StringValue(time.Now().UTC().Format(time.RFC3339))
		
		err = r.client.InstallCluster(ctx, clusterID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error triggering installation",
				fmt.Sprintf("Could not trigger installation for cluster %s: %s", clusterID, err),
			)
			return
		}
	}

	// Wait for installation to complete
	tflog.Info(ctx, "Waiting for installation to complete", map[string]interface{}{
		"cluster_id": clusterID,
		"timeout": createTimeout.String(),
	})

	err = r.waitForInstallationComplete(ctx, clusterID, createTimeout)
	if err != nil {
		// Still save state even if installation fails/times out
		cluster, _ = r.client.GetCluster(ctx, clusterID)
		data.Status = types.StringValue(cluster.Status)
		data.StatusInfo = types.StringValue(cluster.StatusInfo)
		
		resp.Diagnostics.AddError(
			"Installation did not complete",
			fmt.Sprintf("Cluster %s installation did not complete: %s. Current status: %s", clusterID, err, cluster.Status),
		)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	// Get final cluster state
	cluster, err = r.client.GetCluster(ctx, clusterID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving cluster after installation",
			fmt.Sprintf("Could not get cluster %s after installation: %s", clusterID, err),
		)
		return
	}

	data.Status = types.StringValue(cluster.Status)
	data.StatusInfo = types.StringValue(cluster.StatusInfo)
	data.InstallCompletedAt = types.StringValue(time.Now().UTC().Format(time.RFC3339))

	tflog.Info(ctx, "Cluster installation completed successfully", map[string]interface{}{
		"cluster_id": clusterID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ClusterInstallationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ClusterInstallationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := data.ClusterID.ValueString()

	cluster, err := r.client.GetCluster(ctx, clusterID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading cluster",
			fmt.Sprintf("Could not read cluster %s: %s", clusterID, err),
		)
		return
	}

	data.Status = types.StringValue(cluster.Status)
	data.StatusInfo = types.StringValue(cluster.StatusInfo)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ClusterInstallationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Installation cannot be updated - it's a one-time action
	resp.Diagnostics.AddError(
		"Installation cannot be updated",
		"The cluster installation is a one-time action and cannot be modified. To reinstall, delete and recreate the installation resource.",
	)
}

func (r *ClusterInstallationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Deletion is a no-op - we don't uninstall clusters
	// The cluster itself is managed by the cluster resource
	tflog.Info(ctx, "Cluster installation resource deleted (no-op - cluster remains installed)")
}

// Helper function to wait for cluster to be ready for installation
func (r *ClusterInstallationResource) waitForClusterReady(ctx context.Context, clusterID string, expectedHosts int) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for cluster to be ready")
		case <-ticker.C:
			cluster, err := r.client.GetCluster(ctx, clusterID)
			if err != nil {
				return fmt.Errorf("failed to get cluster status: %w", err)
			}

			tflog.Debug(ctx, "Checking cluster readiness", map[string]interface{}{
				"cluster_id": clusterID,
				"status": cluster.Status,
				"host_count": cluster.HostCount,
				"expected_hosts": expectedHosts,
			})

			// Check if cluster is ready for installation
			if cluster.Status == "ready" {
				if cluster.HostCount >= expectedHosts {
					tflog.Info(ctx, "Cluster is ready for installation", map[string]interface{}{
						"cluster_id": clusterID,
						"host_count": cluster.HostCount,
					})
					return nil
				}
			}

			// Check for error states
			if cluster.Status == "error" {
				return fmt.Errorf("cluster is in error state: %s", cluster.StatusInfo)
			}
		}
	}
}

// Helper function to wait for installation to complete
func (r *ClusterInstallationResource) waitForInstallationComplete(ctx context.Context, clusterID string, timeout time.Duration) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	deadline := time.Now().Add(timeout)

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for installation")
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("installation timeout exceeded (%v)", timeout)
			}

			cluster, err := r.client.GetCluster(ctx, clusterID)
			if err != nil {
				return fmt.Errorf("failed to get cluster status: %w", err)
			}

			tflog.Debug(ctx, "Checking installation status", map[string]interface{}{
				"cluster_id": clusterID,
				"status": cluster.Status,
				"status_info": cluster.StatusInfo,
			})

			switch cluster.Status {
			case "installed":
				return nil
			case "error", "cancelled":
				return fmt.Errorf("installation failed with status %s: %s", cluster.Status, cluster.StatusInfo)
			case "installing", "finalizing":
				// Continue waiting
				continue
			default:
				tflog.Warn(ctx, "Unexpected cluster status during installation", map[string]interface{}{
					"status": cluster.Status,
				})
			}
		}
	}
}