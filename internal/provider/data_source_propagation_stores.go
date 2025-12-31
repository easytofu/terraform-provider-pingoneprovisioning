// pingone/providers/pingone-propagation/internal/datasources/propagation_stores.go
package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/easytofu/terraform-provider-pingoneprovisioning/internal/client"
	"github.com/easytofu/terraform-provider-pingoneprovisioning/internal/mappers"
	customtypes "github.com/easytofu/terraform-provider-pingoneprovisioning/internal/types"
	"github.com/easytofu/terraform-provider-pingoneprovisioning/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/patrickcping/pingone-go-sdk-v2/management"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &propagationStoresDataSource{}
	_ datasource.DataSourceWithConfigure = &propagationStoresDataSource{}
)

// propagationStoresDataSource is the data source implementation.
type propagationStoresDataSource struct {
	client *client.Client
}

type propagationStoresDataSourceModel struct {
	EnvironmentId types.String `tfsdk:"environment_id"`
	Type          types.String `tfsdk:"type"`
	StoreId       types.String `tfsdk:"store_id"`
	Stores        types.List   `tfsdk:"stores"`
	Ids           types.List   `tfsdk:"ids"`
}

func NewPropagationStoresDataSource() datasource.DataSource {
	return &propagationStoresDataSource{}
}

func (d *propagationStoresDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_propagation_stores"
}

func (d *propagationStoresDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches PingOne provisioning propagation stores.",
		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{
				Description: "The ID of the environment.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "Optional filter by propagation store type.",
				Optional:    true,
			},
			"store_id": schema.StringAttribute{
				Description: "Optional filter by a specific propagation store ID.",
				Optional:    true,
			},
			"stores": schema.ListAttribute{
				Description: "List of propagation stores.",
				Computed:    true,
				ElementType: customtypes.PropagationStoreModelType(),
			},
			"ids": schema.ListAttribute{
				Description: "List of propagation store IDs found.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *propagationStoresDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

// Read refreshes the Terraform state with the latest data.
func (d *propagationStoresDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state propagationStoresDataSourceModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	environmentID := state.EnvironmentId.ValueString()
	filterType := state.Type.ValueString()
	filterTypeAPI := utils.NormalizePropagationStoreTypeForAPI(filterType)
	filterStoreId := state.StoreId.ValueString()

	tflog.Info(ctx, "Starting read of propagation stores", map[string]interface{}{
		"environment_id": environmentID,
		"filter_type":    filterType,
		"filter_storeId": filterStoreId,
	})

	apiClient := d.client.API

	iterator := apiClient.PropagationStoresApi.ReadAllStores(ctx, environmentID).Execute()

	var propagationStores []customtypes.PropagationStoreModel
	var ids []string

	for cursor, iterErr := range iterator {
		if iterErr != nil {
			resp.Diagnostics.AddError(
				"Error Reading Propagation Stores",
				fmt.Sprintf("Could not iterate propagation stores: %s", iterErr),
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
				"Error Reading Propagation Stores",
				fmt.Sprintf("Could not read propagation stores response: %s", readErr),
			)
			return
		}

		var rawResponse map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &rawResponse); err != nil {
			resp.Diagnostics.AddError(
				"Error Parsing Propagation Stores",
				fmt.Sprintf("Could not parse propagation stores response: %s", err),
			)
			return
		}

		embedded, ok := rawResponse["_embedded"].(map[string]interface{})
		if !ok {
			continue
		}

		stores, ok := embedded["stores"].([]interface{})
		if !ok {
			continue
		}

		for _, s := range stores {
			sMap, ok := s.(map[string]interface{})
			if !ok {
				continue
			}

			storeTypeRaw, _ := sMap["type"].(string)
			storeStatusRaw, _ := sMap["status"].(string)

			storeJSON, err := json.Marshal(sMap)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error Parsing Propagation Store",
					fmt.Sprintf("Could not marshal propagation store response: %s", err),
				)
				return
			}

			var storeObj management.PropagationStore
			if err := json.Unmarshal(storeJSON, &storeObj); err != nil {
				resp.Diagnostics.AddError(
					"Error Parsing Propagation Store",
					fmt.Sprintf("Could not unmarshal propagation store response: %s", err),
				)
				return
			}

			if !state.Type.IsNull() && !state.Type.IsUnknown() {
				if !strings.EqualFold(storeTypeRaw, filterTypeAPI) {
					continue
				}
			}

			if !state.StoreId.IsNull() && !state.StoreId.IsUnknown() {
				if storeObj.GetId() != filterStoreId {
					continue
				}
			}

			storeModel := d.apiToModel(&storeObj, storeTypeRaw, storeStatusRaw, environmentID)
			propagationStores = append(propagationStores, storeModel)
			ids = append(ids, storeObj.GetId())
		}
	}

	tflog.Info(ctx, "Finished reading propagation stores", map[string]interface{}{
		"total_found": len(propagationStores),
	})

	storesList, diags := types.ListValueFrom(ctx, customtypes.PropagationStoreModelType(), propagationStores)
	resp.Diagnostics.Append(diags...)
	state.Stores = storesList

	idsList, diags := types.ListValueFrom(ctx, types.StringType, ids)
	resp.Diagnostics.Append(diags...)
	state.Ids = idsList

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// apiToModel maps the SDK object to the Terraform model.
func (d *propagationStoresDataSource) apiToModel(apiObj *management.PropagationStore, rawType string, rawStatus string, environmentId string) customtypes.PropagationStoreModel {
	apiType := rawType
	if apiType == "" || apiType == "UNKNOWN" {
		apiType = string(apiObj.GetType())
	}
	tfType := utils.NormalizePropagationStoreTypeForTerraform(apiType, "")

	model := customtypes.PropagationStoreModel{
		Id:            types.StringValue(apiObj.GetId()),
		EnvironmentId: types.StringValue(environmentId),
		Name:          types.StringValue(apiObj.GetName()),
		Type:          types.StringValue(tfType),
	}

	if v, ok := apiObj.GetDescriptionOk(); ok {
		model.Description = types.StringValue(*v)
	} else {
		model.Description = types.StringNull()
	}

	if image, ok := apiObj.GetImageOk(); ok && image != nil {
		if v, ok := image.GetIdOk(); ok && v != nil {
			model.ImageId = types.StringValue(*v)
		} else {
			model.ImageId = types.StringNull()
		}

		if v, ok := image.GetHrefOk(); ok && v != nil {
			model.ImageHref = types.StringValue(*v)
		} else {
			model.ImageHref = types.StringNull()
		}
	} else {
		model.ImageId = types.StringNull()
		model.ImageHref = types.StringNull()
	}

	if v, ok := apiObj.GetManagedOk(); ok {
		model.Managed = types.BoolValue(*v)
	} else {
		model.Managed = types.BoolNull()
	}

	status := rawStatus
	if status == "" {
		if v, ok := apiObj.GetStatusOk(); ok && v != nil && string(*v) != "" && string(*v) != "UNKNOWN" {
			status = string(*v)
		}
	}
	if status != "" {
		model.Status = types.StringValue(status)
	} else {
		model.Status = types.StringNull()
	}

	if v, ok := apiObj.GetSyncStatusOk(); ok && v != nil {
		attrs := map[string]attr.Value{
			"last_sync_time": types.StringNull(),
			"next_sync_time": types.StringNull(),
			"status":         types.StringNull(),
			"details":        types.StringNull(),
		}

		if vv, ok := v.GetLastSyncAtOk(); ok && vv != nil {
			attrs["last_sync_time"] = types.StringValue(vv.Format(time.RFC3339))
		}
		if vv, ok := v.GetSyncStateOk(); ok && vv != nil {
			attrs["status"] = types.StringValue(string(*vv))
		}
		if vv, ok := v.GetDetailsOk(); ok && vv != nil {
			attrs["details"] = types.StringValue(*vv)
		}

		model.SyncStatus = types.ObjectValueMust(map[string]attr.Type{
			"last_sync_time": types.StringType,
			"next_sync_time": types.StringType,
			"status":         types.StringType,
			"details":        types.StringType,
		}, attrs)
	} else {
		model.SyncStatus = types.ObjectNull(map[string]attr.Type{
			"last_sync_time": types.StringType,
			"next_sync_time": types.StringType,
			"status":         types.StringType,
			"details":        types.StringType,
		})
	}

	config := apiObj.GetConfiguration()
	mappers.ApplyPropagationStoreConfigurationFromMap(&model, tfType, config, nil)

	return model
}
