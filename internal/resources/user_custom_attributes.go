package resources

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/patrickcping/pingone-go-sdk-v2/management"
	"github.com/pingidentity/terraform-provider-pingoneprovisioning/internal/providerdata"
	customtypes "github.com/pingidentity/terraform-provider-pingoneprovisioning/internal/types"
	"github.com/pingidentity/terraform-provider-pingoneprovisioning/internal/utils"
)

var (
	_ resource.Resource                = &userCustomAttributesResource{}
	_ resource.ResourceWithConfigure   = &userCustomAttributesResource{}
	_ resource.ResourceWithImportState = &userCustomAttributesResource{}
)

type userCustomAttributesResource struct {
	client *providerdata.Client
}

func NewUserCustomAttributesResource() resource.Resource {
	return &userCustomAttributesResource{}
}

func (r *userCustomAttributesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_custom_attributes"
}

func (r *userCustomAttributesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages custom schema attributes for an existing PingOne user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Internal identifier for this custom attribute mapping.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_id": schema.StringAttribute{
				Description: "The ID of the environment.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_id": schema.StringAttribute{
				Description: "The PingOne user ID to update.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"attributes": schema.DynamicAttribute{
				Description: "Map of custom user attribute values keyed by schema attribute name.",
				Required:    true,
			},
		},
	}
}

func (r *userCustomAttributesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	clientData, ok := req.ProviderData.(*providerdata.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *providerdata.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = clientData
}

func (r *userCustomAttributesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan customtypes.UserCustomAttributesModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, diags := expandCustomAttributes(ctx, plan.Attributes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := patchUserCustomAttributes(ctx, r.client.API, plan.EnvironmentId.ValueString(), plan.UserId.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating User Custom Attributes",
			fmt.Sprintf("Could not update user custom attributes: %s", utils.HandleSDKError(err, httpResp)),
		)
		return
	}

	plan.Id = types.StringValue(buildUserCustomAttributesID(plan.EnvironmentId.ValueString(), plan.UserId.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userCustomAttributesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state customtypes.UserCustomAttributesModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	userMap, httpResp, err := readUserCustomAttributes(ctx, r.client.API, state.EnvironmentId.ValueString(), state.UserId.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading User",
			fmt.Sprintf("Could not read user custom attributes: %s", utils.HandleSDKError(err, httpResp)),
		)
		return
	}

	updated, mapDiags := mergeCustomAttributesFromAPI(ctx, state.Attributes, userMap)
	resp.Diagnostics.Append(mapDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Id = types.StringValue(buildUserCustomAttributesID(state.EnvironmentId.ValueString(), state.UserId.ValueString()))
	state.Attributes = updated
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *userCustomAttributesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan customtypes.UserCustomAttributesModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, diags := expandCustomAttributes(ctx, plan.Attributes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := patchUserCustomAttributes(ctx, r.client.API, plan.EnvironmentId.ValueString(), plan.UserId.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating User Custom Attributes",
			fmt.Sprintf("Could not update user custom attributes: %s", utils.HandleSDKError(err, httpResp)),
		)
		return
	}

	plan.Id = types.StringValue(buildUserCustomAttributesID(plan.EnvironmentId.ValueString(), plan.UserId.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userCustomAttributesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Intentionally leave custom attributes in place; removing this resource only clears Terraform state.
}

func (r *userCustomAttributesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := utils.SplitImportID(req.ID, 2)
	if parts == nil {
		resp.Diagnostics.AddError(
			"Invalid Import Identifier",
			"Expected import identifier format: <environment_id>/<user_id>.",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), buildUserCustomAttributesID(parts[0], parts[1]))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("attributes"), types.DynamicUnknown())...)
}

func buildUserCustomAttributesID(environmentID string, userID string) string {
	return fmt.Sprintf("%s/%s", strings.TrimSpace(environmentID), strings.TrimSpace(userID))
}

func patchUserCustomAttributes(ctx context.Context, apiClient *management.APIClient, environmentID string, userID string, payload map[string]interface{}) (*http.Response, error) {
	if apiClient == nil {
		return nil, fmt.Errorf("nil api client")
	}

	cfg := apiClient.GetConfig()
	if cfg == nil {
		return nil, fmt.Errorf("api client has nil config")
	}
	if cfg.HTTPClient == nil {
		return nil, fmt.Errorf("api client has nil http client")
	}

	basePath, err := cfg.ServerURLWithContext(ctx, "UsersApiService.UpdateUserPatch")
	if err != nil {
		return nil, err
	}
	basePath = strings.TrimRight(strings.TrimSpace(basePath), "/")

	endpoint := fmt.Sprintf(
		"%s/environments/%s/users/%s",
		basePath,
		url.PathEscape(environmentID),
		url.PathEscape(userID),
	)

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode >= 300 {
		return resp, fmt.Errorf("%s", utils.HandleSDKError(fmt.Errorf("%s", resp.Status), resp))
	}

	return resp, nil
}

func readUserCustomAttributes(ctx context.Context, apiClient *management.APIClient, environmentID string, userID string) (map[string]interface{}, *http.Response, error) {
	if apiClient == nil {
		return nil, nil, fmt.Errorf("nil api client")
	}

	cfg := apiClient.GetConfig()
	if cfg == nil {
		return nil, nil, fmt.Errorf("api client has nil config")
	}
	if cfg.HTTPClient == nil {
		return nil, nil, fmt.Errorf("api client has nil http client")
	}

	basePath, err := cfg.ServerURLWithContext(ctx, "UsersApiService.ReadUser")
	if err != nil {
		return nil, nil, err
	}
	basePath = strings.TrimRight(strings.TrimSpace(basePath), "/")

	endpoint := fmt.Sprintf(
		"%s/environments/%s/users/%s",
		basePath,
		url.PathEscape(environmentID),
		url.PathEscape(userID),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, resp, err
	}

	if resp.StatusCode >= 300 {
		return nil, resp, fmt.Errorf("%s", utils.HandleSDKError(fmt.Errorf("%s", resp.Status), resp))
	}

	decoded, err := utils.DecodeResponseJSON(resp)
	if err != nil {
		return nil, resp, err
	}

	userMap, ok := decoded.(map[string]interface{})
	if !ok {
		return nil, resp, fmt.Errorf("unexpected user response shape")
	}

	return userMap, resp, nil
}

func mergeCustomAttributesFromAPI(ctx context.Context, current types.Dynamic, userMap map[string]interface{}) (types.Dynamic, diag.Diagnostics) {
	var diags diag.Diagnostics

	if current.IsNull() || current.IsUnknown() || current.IsUnderlyingValueUnknown() {
		return current, diags
	}

	if current.IsUnderlyingValueNull() {
		return current, diags
	}

	underlying := current.UnderlyingValue()
	if underlying == nil {
		return current, diags
	}

	var elements map[string]attr.Value
	var attrTypes map[string]attr.Type
	var elemType attr.Type
	switch v := underlying.(type) {
	case types.Map:
		elements = v.Elements()
		elemType = v.ElementType(ctx)
	case types.Object:
		elements = v.Attributes()
		attrTypes = v.AttributeTypes(ctx)
	default:
		diags.AddError(
			"Invalid Custom Attributes Type",
			fmt.Sprintf("The attributes value must be an object or map, got %T.", underlying),
		)
		return types.DynamicUnknown(), diags
	}

	merged := make(map[string]attr.Value, len(elements))
	for key := range elements {
		raw, ok := userMap[key]
		if !ok {
			if attrTypes != nil {
				merged[key] = nullValueForType(attrTypes[key])
			} else if elemType != nil {
				merged[key] = nullValueForType(elemType)
			} else {
				merged[key] = types.DynamicNull()
			}
			continue
		}

		var targetType attr.Type
		if attrTypes != nil {
			targetType = attrTypes[key]
		} else {
			targetType = elemType
		}

		converted, convDiags := apiValueToAttrValue(ctx, raw, targetType)
		diags.Append(convDiags...)
		if diags.HasError() {
			return types.DynamicUnknown(), diags
		}
		merged[key] = converted
	}

	if attrTypes != nil {
		objVal, objDiags := types.ObjectValue(attrTypes, merged)
		diags.Append(objDiags...)
		if diags.HasError() {
			return types.DynamicUnknown(), diags
		}
		return types.DynamicValue(objVal), diags
	}

	mapVal, mapDiags := types.MapValue(elemType, merged)
	diags.Append(mapDiags...)
	if diags.HasError() {
		return types.DynamicUnknown(), diags
	}
	return types.DynamicValue(mapVal), diags
}

func expandCustomAttributes(ctx context.Context, attrs types.Dynamic) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	if attrs.IsNull() {
		diags.AddError("Missing Custom Attributes", "The attributes map must be provided.")
		return nil, diags
	}
	if attrs.IsUnknown() {
		diags.AddError("Unknown Custom Attributes", "The attributes map must be known before applying.")
		return nil, diags
	}
	if attrs.IsUnderlyingValueUnknown() {
		diags.AddError("Unknown Custom Attributes", "The attributes map must be known before applying.")
		return nil, diags
	}
	if attrs.IsUnderlyingValueNull() {
		return map[string]interface{}{}, diags
	}

	underlying := attrs.UnderlyingValue()
	if underlying == nil {
		return map[string]interface{}{}, diags
	}

	var elements map[string]attr.Value
	switch v := underlying.(type) {
	case types.Map:
		elements = v.Elements()
	case types.Object:
		elements = v.Attributes()
	default:
		diags.AddError(
			"Invalid Custom Attributes Type",
			fmt.Sprintf("The attributes value must be an object or map, got %T.", underlying),
		)
		return nil, diags
	}

	expanded := make(map[string]interface{}, len(elements))
	for key, value := range elements {
		raw, valDiags := attributeValueToInterface(ctx, value)
		diags.Append(valDiags...)
		if diags.HasError() {
			return nil, diags
		}
		expanded[key] = raw
	}

	return expanded, diags
}

func attributeValueToInterface(ctx context.Context, value attr.Value) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	if value == nil {
		return nil, diags
	}

	switch v := value.(type) {
	case types.Dynamic:
		if v.IsUnknown() {
			diags.AddError("Unknown Custom Attribute Value", "Custom attribute values must be known before applying.")
			return nil, diags
		}
		if v.IsNull() {
			return nil, diags
		}
		underlying := v.UnderlyingValue()
		if underlying == nil {
			return nil, diags
		}
		if underlying.IsUnknown() {
			diags.AddError("Unknown Custom Attribute Value", "Custom attribute values must be known before applying.")
			return nil, diags
		}
		if underlying.IsNull() {
			return nil, diags
		}
		return attributeValueToInterface(ctx, underlying)
	case types.String:
		if v.IsUnknown() {
			diags.AddError("Unknown Custom Attribute Value", "Custom attribute values must be known before applying.")
			return nil, diags
		}
		if v.IsNull() {
			return nil, diags
		}
		return v.ValueString(), diags
	case types.Bool:
		if v.IsUnknown() {
			diags.AddError("Unknown Custom Attribute Value", "Custom attribute values must be known before applying.")
			return nil, diags
		}
		if v.IsNull() {
			return nil, diags
		}
		return v.ValueBool(), diags
	case types.Int64:
		if v.IsUnknown() {
			diags.AddError("Unknown Custom Attribute Value", "Custom attribute values must be known before applying.")
			return nil, diags
		}
		if v.IsNull() {
			return nil, diags
		}
		return v.ValueInt64(), diags
	case types.Float64:
		if v.IsUnknown() {
			diags.AddError("Unknown Custom Attribute Value", "Custom attribute values must be known before applying.")
			return nil, diags
		}
		if v.IsNull() {
			return nil, diags
		}
		return v.ValueFloat64(), diags
	case types.Number:
		if v.IsUnknown() {
			diags.AddError("Unknown Custom Attribute Value", "Custom attribute values must be known before applying.")
			return nil, diags
		}
		if v.IsNull() {
			return nil, diags
		}
		if v.ValueBigFloat() == nil {
			return nil, diags
		}
		floatVal, _ := v.ValueBigFloat().Float64()
		return floatVal, diags
	case types.List:
		if v.IsUnknown() {
			diags.AddError("Unknown Custom Attribute Value", "Custom attribute values must be known before applying.")
			return nil, diags
		}
		if v.IsNull() {
			return nil, diags
		}
		elements := v.Elements()
		result := make([]interface{}, 0, len(elements))
		for _, elem := range elements {
			raw, elemDiags := attributeValueToInterface(ctx, elem)
			diags.Append(elemDiags...)
			if diags.HasError() {
				return nil, diags
			}
			result = append(result, raw)
		}
		return result, diags
	case types.Set:
		if v.IsUnknown() {
			diags.AddError("Unknown Custom Attribute Value", "Custom attribute values must be known before applying.")
			return nil, diags
		}
		if v.IsNull() {
			return nil, diags
		}
		elements := v.Elements()
		result := make([]interface{}, 0, len(elements))
		for _, elem := range elements {
			raw, elemDiags := attributeValueToInterface(ctx, elem)
			diags.Append(elemDiags...)
			if diags.HasError() {
				return nil, diags
			}
			result = append(result, raw)
		}
		return result, diags
	case types.Tuple:
		if v.IsUnknown() {
			diags.AddError("Unknown Custom Attribute Value", "Custom attribute values must be known before applying.")
			return nil, diags
		}
		if v.IsNull() {
			return nil, diags
		}
		elements := v.Elements()
		result := make([]interface{}, 0, len(elements))
		for _, elem := range elements {
			raw, elemDiags := attributeValueToInterface(ctx, elem)
			diags.Append(elemDiags...)
			if diags.HasError() {
				return nil, diags
			}
			result = append(result, raw)
		}
		return result, diags
	case types.Map:
		if v.IsUnknown() {
			diags.AddError("Unknown Custom Attribute Value", "Custom attribute values must be known before applying.")
			return nil, diags
		}
		if v.IsNull() {
			return nil, diags
		}
		elements := v.Elements()
		result := make(map[string]interface{}, len(elements))
		for key, elem := range elements {
			raw, elemDiags := attributeValueToInterface(ctx, elem)
			diags.Append(elemDiags...)
			if diags.HasError() {
				return nil, diags
			}
			result[key] = raw
		}
		return result, diags
	case types.Object:
		if v.IsUnknown() {
			diags.AddError("Unknown Custom Attribute Value", "Custom attribute values must be known before applying.")
			return nil, diags
		}
		if v.IsNull() {
			return nil, diags
		}
		attrs := v.Attributes()
		result := make(map[string]interface{}, len(attrs))
		for key, elem := range attrs {
			raw, elemDiags := attributeValueToInterface(ctx, elem)
			diags.Append(elemDiags...)
			if diags.HasError() {
				return nil, diags
			}
			result[key] = raw
		}
		return result, diags
	default:
		diags.AddError(
			"Unsupported Custom Attribute Value",
			fmt.Sprintf("Custom attribute value type %T is not supported.", value),
		)
		return nil, diags
	}
}

func interfaceToDynamicValue(ctx context.Context, value interface{}) (types.Dynamic, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch v := value.(type) {
	case nil:
		return types.DynamicNull(), diags
	case string:
		return types.DynamicValue(types.StringValue(v)), diags
	case bool:
		return types.DynamicValue(types.BoolValue(v)), diags
	case float64:
		return types.DynamicValue(types.NumberValue(big.NewFloat(v))), diags
	case int:
		return types.DynamicValue(types.NumberValue(big.NewFloat(float64(v)))), diags
	case int64:
		return types.DynamicValue(types.NumberValue(big.NewFloat(float64(v)))), diags
	case []interface{}:
		elements := make([]attr.Value, 0, len(v))
		for _, elem := range v {
			dyn, elemDiags := interfaceToDynamicValue(ctx, elem)
			diags.Append(elemDiags...)
			if diags.HasError() {
				return types.DynamicUnknown(), diags
			}
			elements = append(elements, dyn)
		}

		listVal, listDiags := types.ListValue(types.DynamicType, elements)
		diags.Append(listDiags...)
		if diags.HasError() {
			return types.DynamicUnknown(), diags
		}
		return types.DynamicValue(listVal), diags
	case map[string]interface{}:
		elements := make(map[string]attr.Value, len(v))
		for key, elem := range v {
			dyn, elemDiags := interfaceToDynamicValue(ctx, elem)
			diags.Append(elemDiags...)
			if diags.HasError() {
				return types.DynamicUnknown(), diags
			}
			elements[key] = dyn
		}

		mapVal, mapDiags := types.MapValue(types.DynamicType, elements)
		diags.Append(mapDiags...)
		if diags.HasError() {
			return types.DynamicUnknown(), diags
		}
		return types.DynamicValue(mapVal), diags
	default:
		diags.AddError(
			"Unsupported Custom Attribute Value",
			fmt.Sprintf("API value type %T is not supported in custom attributes.", value),
		)
		return types.DynamicUnknown(), diags
	}
}

func apiValueToAttrValue(ctx context.Context, value interface{}, targetType attr.Type) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	if targetType == nil {
		diags.AddError("Invalid Custom Attribute Type", "Custom attribute type is nil.")
		return types.DynamicUnknown(), diags
	}

	if value == nil {
		return nullValueForType(targetType), diags
	}

	switch t := targetType.(type) {
	case basetypes.StringType:
		s, ok := value.(string)
		if !ok {
			diags.AddError("Invalid Custom Attribute Value", fmt.Sprintf("Expected string value, got %T.", value))
			return types.StringNull(), diags
		}
		return types.StringValue(s), diags
	case basetypes.BoolType:
		b, ok := value.(bool)
		if !ok {
			diags.AddError("Invalid Custom Attribute Value", fmt.Sprintf("Expected bool value, got %T.", value))
			return types.BoolNull(), diags
		}
		return types.BoolValue(b), diags
	case basetypes.Int64Type:
		switch v := value.(type) {
		case int64:
			return types.Int64Value(v), diags
		case int:
			return types.Int64Value(int64(v)), diags
		case float64:
			if math.Trunc(v) != v {
				diags.AddError("Invalid Custom Attribute Value", fmt.Sprintf("Expected whole number for int64, got %v.", v))
				return types.Int64Null(), diags
			}
			return types.Int64Value(int64(v)), diags
		default:
			diags.AddError("Invalid Custom Attribute Value", fmt.Sprintf("Expected int64 value, got %T.", value))
			return types.Int64Null(), diags
		}
	case basetypes.Float64Type:
		switch v := value.(type) {
		case float64:
			return types.Float64Value(v), diags
		case int:
			return types.Float64Value(float64(v)), diags
		case int64:
			return types.Float64Value(float64(v)), diags
		default:
			diags.AddError("Invalid Custom Attribute Value", fmt.Sprintf("Expected float64 value, got %T.", value))
			return types.Float64Null(), diags
		}
	case basetypes.NumberType:
		switch v := value.(type) {
		case float64:
			return types.NumberValue(big.NewFloat(v)), diags
		case int:
			return types.NumberValue(big.NewFloat(float64(v))), diags
		case int64:
			return types.NumberValue(big.NewFloat(float64(v))), diags
		default:
			diags.AddError("Invalid Custom Attribute Value", fmt.Sprintf("Expected number value, got %T.", value))
			return types.NumberNull(), diags
		}
	case types.ListType:
		rawList, ok := value.([]interface{})
		if !ok {
			diags.AddError("Invalid Custom Attribute Value", fmt.Sprintf("Expected list value, got %T.", value))
			return types.ListNull(t.ElemType), diags
		}
		elements := make([]attr.Value, 0, len(rawList))
		for _, elem := range rawList {
			converted, convDiags := apiValueToAttrValue(ctx, elem, t.ElemType)
			diags.Append(convDiags...)
			if diags.HasError() {
				return types.ListNull(t.ElemType), diags
			}
			elements = append(elements, converted)
		}
		listVal, listDiags := types.ListValue(t.ElemType, elements)
		diags.Append(listDiags...)
		if diags.HasError() {
			return types.ListNull(t.ElemType), diags
		}
		return listVal, diags
	case types.SetType:
		rawList, ok := value.([]interface{})
		if !ok {
			diags.AddError("Invalid Custom Attribute Value", fmt.Sprintf("Expected set value, got %T.", value))
			return types.SetNull(t.ElemType), diags
		}
		elements := make([]attr.Value, 0, len(rawList))
		for _, elem := range rawList {
			converted, convDiags := apiValueToAttrValue(ctx, elem, t.ElemType)
			diags.Append(convDiags...)
			if diags.HasError() {
				return types.SetNull(t.ElemType), diags
			}
			elements = append(elements, converted)
		}
		setVal, setDiags := types.SetValue(t.ElemType, elements)
		diags.Append(setDiags...)
		if diags.HasError() {
			return types.SetNull(t.ElemType), diags
		}
		return setVal, diags
	case types.TupleType:
		rawList, ok := value.([]interface{})
		if !ok {
			diags.AddError("Invalid Custom Attribute Value", fmt.Sprintf("Expected tuple value, got %T.", value))
			return types.TupleNull(t.ElementTypes()), diags
		}
		elemTypes := t.ElementTypes()
		if len(rawList) != len(elemTypes) {
			diags.AddError("Invalid Custom Attribute Value", "Tuple length does not match expected type.")
			return types.TupleNull(elemTypes), diags
		}
		elements := make([]attr.Value, 0, len(rawList))
		for i, elem := range rawList {
			converted, convDiags := apiValueToAttrValue(ctx, elem, elemTypes[i])
			diags.Append(convDiags...)
			if diags.HasError() {
				return types.TupleNull(elemTypes), diags
			}
			elements = append(elements, converted)
		}
		tupleVal, tupleDiags := types.TupleValue(elemTypes, elements)
		diags.Append(tupleDiags...)
		if diags.HasError() {
			return types.TupleNull(elemTypes), diags
		}
		return tupleVal, diags
	case types.MapType:
		rawMap, ok := value.(map[string]interface{})
		if !ok {
			diags.AddError("Invalid Custom Attribute Value", fmt.Sprintf("Expected map value, got %T.", value))
			return types.MapNull(t.ElemType), diags
		}
		elements := make(map[string]attr.Value, len(rawMap))
		for key, elem := range rawMap {
			converted, convDiags := apiValueToAttrValue(ctx, elem, t.ElemType)
			diags.Append(convDiags...)
			if diags.HasError() {
				return types.MapNull(t.ElemType), diags
			}
			elements[key] = converted
		}
		mapVal, mapDiags := types.MapValue(t.ElemType, elements)
		diags.Append(mapDiags...)
		if diags.HasError() {
			return types.MapNull(t.ElemType), diags
		}
		return mapVal, diags
	case types.ObjectType:
		rawMap, ok := value.(map[string]interface{})
		if !ok {
			diags.AddError("Invalid Custom Attribute Value", fmt.Sprintf("Expected object value, got %T.", value))
			return types.ObjectNull(t.AttributeTypes()), diags
		}
		attrTypes := t.AttributeTypes()
		elements := make(map[string]attr.Value, len(attrTypes))
		for key, attrType := range attrTypes {
			rawVal, hasKey := rawMap[key]
			if !hasKey {
				elements[key] = nullValueForType(attrType)
				continue
			}
			converted, convDiags := apiValueToAttrValue(ctx, rawVal, attrType)
			diags.Append(convDiags...)
			if diags.HasError() {
				return types.ObjectNull(attrTypes), diags
			}
			elements[key] = converted
		}
		objVal, objDiags := types.ObjectValue(attrTypes, elements)
		diags.Append(objDiags...)
		if diags.HasError() {
			return types.ObjectNull(attrTypes), diags
		}
		return objVal, diags
	case basetypes.DynamicType:
		dyn, dynDiags := interfaceToDynamicValue(ctx, value)
		diags.Append(dynDiags...)
		return dyn, diags
	default:
		diags.AddError("Unsupported Custom Attribute Type", fmt.Sprintf("Custom attribute type %T is not supported.", targetType))
		return types.DynamicUnknown(), diags
	}
}

func nullValueForType(targetType attr.Type) attr.Value {
	switch t := targetType.(type) {
	case basetypes.StringType:
		return types.StringNull()
	case basetypes.BoolType:
		return types.BoolNull()
	case basetypes.Int64Type:
		return types.Int64Null()
	case basetypes.Float64Type:
		return types.Float64Null()
	case basetypes.NumberType:
		return types.NumberNull()
	case types.ListType:
		return types.ListNull(t.ElemType)
	case types.SetType:
		return types.SetNull(t.ElemType)
	case types.MapType:
		return types.MapNull(t.ElemType)
	case types.TupleType:
		return types.TupleNull(t.ElementTypes())
	case types.ObjectType:
		return types.ObjectNull(t.AttributeTypes())
	case basetypes.DynamicType:
		return types.DynamicNull()
	default:
		return types.DynamicNull()
	}
}
