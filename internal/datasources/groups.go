// pingone/providers/pingone-propagation/internal/datasources/groups.go
package datasources

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/patrickcping/pingone-go-sdk-v2/management"
	"github.com/pingidentity/terraform-provider-pingoneprovisioning/internal/mappers"
	"github.com/pingidentity/terraform-provider-pingoneprovisioning/internal/providerdata"
	customtypes "github.com/pingidentity/terraform-provider-pingoneprovisioning/internal/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &groupsDataSource{}
	_ datasource.DataSourceWithConfigure = &groupsDataSource{}
)

type groupsDataSource struct {
	client *providerdata.Client
}

type groupsDataSourceModel struct {
	EnvironmentId types.String `tfsdk:"environment_id"`
	Filter        types.String `tfsdk:"filter"`
	Groups        types.List   `tfsdk:"groups"`
	Ids           types.List   `tfsdk:"ids"`
}

func NewGroupsDataSource() datasource.DataSource {
	return &groupsDataSource{}
}

func (d *groupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_groups"
}

func (d *groupsDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches PingOne groups for an environment.",
		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{
				Description: "The ID of the environment.",
				Required:    true,
			},
			"filter": schema.StringAttribute{
				Description: "Optional SCIM filter to apply when listing groups.",
				Optional:    true,
			},
			"groups": schema.ListAttribute{
				Description: "List of groups returned by the query.",
				Computed:    true,
				ElementType: customtypes.GroupModelType(),
			},
			"ids": schema.ListAttribute{
				Description: "List of group IDs found.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *groupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*providerdata.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *providerdata.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *groupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state groupsDataSourceModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	environmentID := strings.TrimSpace(state.EnvironmentId.ValueString())
	filter := ""
	if !state.Filter.IsNull() && !state.Filter.IsUnknown() {
		filter = strings.TrimSpace(state.Filter.ValueString())
	}

	tflog.Info(ctx, "Starting read of PingOne groups", map[string]interface{}{
		"environment_id": environmentID,
		"filter":         filter,
	})

	apiClient := d.client.API
	request := apiClient.GroupsApi.ReadAllGroups(ctx, environmentID)
	if filter != "" {
		request = request.Filter(filter)
	}

	iterator := request.Execute()

	type groupEntry struct {
		id    string
		model customtypes.GroupModel
	}

	seen := make(map[string]bool)
	var entries []groupEntry

	for cursor, iterErr := range iterator {
		if iterErr != nil {
			resp.Diagnostics.AddError(
				"Error Reading Groups",
				fmt.Sprintf("Could not iterate groups: %s", iterErr),
			)
			return
		}

		if cursor.HTTPResponse == nil {
			continue
		}

		bodyBytes, readErr := io.ReadAll(cursor.HTTPResponse.Body)
		_ = cursor.HTTPResponse.Body.Close()
		if readErr != nil {
			resp.Diagnostics.AddError(
				"Error Reading Groups",
				fmt.Sprintf("Could not read groups response: %s", readErr),
			)
			return
		}

		var rawResponse map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &rawResponse); err != nil {
			resp.Diagnostics.AddError(
				"Error Parsing Groups",
				fmt.Sprintf("Could not parse groups response: %s", err),
			)
			return
		}

		embedded, ok := rawResponse["_embedded"].(map[string]interface{})
		if !ok {
			continue
		}

		groups, ok := embedded["groups"].([]interface{})
		if !ok {
			continue
		}

		for _, g := range groups {
			gMap, ok := g.(map[string]interface{})
			if !ok {
				continue
			}

			groupJSON, err := json.Marshal(gMap)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error Parsing Group",
					fmt.Sprintf("Could not marshal group response: %s", err),
				)
				return
			}

			var groupObj management.Group
			if err := json.Unmarshal(groupJSON, &groupObj); err != nil {
				resp.Diagnostics.AddError(
					"Error Parsing Group",
					fmt.Sprintf("Could not unmarshal group response: %s", err),
				)
				return
			}

			id := strings.TrimSpace(groupObj.GetId())
			if id == "" || seen[id] {
				continue
			}
			seen[id] = true

			entry := groupEntry{
				id:    id,
				model: mappers.GroupToModel(&groupObj, environmentID),
			}
			entries = append(entries, entry)
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].id < entries[j].id
	})

	groupsOut := make([]customtypes.GroupModel, 0, len(entries))
	idsOut := make([]string, 0, len(entries))
	for _, entry := range entries {
		groupsOut = append(groupsOut, entry.model)
		idsOut = append(idsOut, entry.id)
	}

	tflog.Info(ctx, "Finished reading PingOne groups", map[string]interface{}{
		"total_found": len(groupsOut),
	})

	groupsList, diags := types.ListValueFrom(ctx, customtypes.GroupModelType(), groupsOut)
	resp.Diagnostics.Append(diags...)
	state.Groups = groupsList

	idsList, diags := types.ListValueFrom(ctx, types.StringType, idsOut)
	resp.Diagnostics.Append(diags...)
	state.Ids = idsList

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
