package mappers

import (
	"testing"

	customtypes "github.com/easytofu/terraform-provider-pingoneprovisioning/internal/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPingOneToMap_OmitsUnsetConnectionFields(t *testing.T) {
	t.Parallel()

	cfg := &customtypes.ConfigurationPingOne{
		AuthenticationMethod: types.StringNull(),
		BaseUrl:              types.StringNull(),
		ScimUrl:              types.StringNull(),
		CreateUsers:          types.BoolValue(true),
		DisableUsers:         types.BoolValue(true),
		UpdateUsers:          types.BoolValue(true),
		RemoveAction:         types.StringValue("Delete"),
	}

	m := PingOneToMap(cfg)

	if _, ok := m["AUTHENTICATION_METHOD"]; ok {
		t.Fatalf("expected AUTHENTICATION_METHOD to be omitted when unset")
	}
	if _, ok := m["BASE_URL"]; ok {
		t.Fatalf("expected BASE_URL to be omitted when unset")
	}
	if _, ok := m["SCIM_URL"]; ok {
		t.Fatalf("expected SCIM_URL to be omitted when unset")
	}

	if v, ok := m["CREATE_USERS"]; !ok || v != true {
		t.Fatalf("expected CREATE_USERS=true, got %#v (present=%v)", v, ok)
	}
	if v, ok := m["DISABLE_USERS"]; !ok || v != true {
		t.Fatalf("expected DISABLE_USERS=true, got %#v (present=%v)", v, ok)
	}
	if v, ok := m["UPDATE_USERS"]; !ok || v != true {
		t.Fatalf("expected UPDATE_USERS=true, got %#v (present=%v)", v, ok)
	}
	if v, ok := m["REMOVE_ACTION"]; !ok || v != "Delete" {
		t.Fatalf("expected REMOVE_ACTION=\"Delete\", got %#v (present=%v)", v, ok)
	}
}

func TestPingOneFromMap_EmptyStringsBecomeNull(t *testing.T) {
	t.Parallel()

	cfg := map[string]interface{}{
		"AUTHENTICATION_METHOD": "",
		"BASE_URL":              "",
		"SCIM_URL":              "",
	}

	out := &customtypes.ConfigurationPingOne{}
	PingOneFromMap(out, cfg)

	if !out.AuthenticationMethod.IsNull() {
		t.Fatalf("expected authentication_method to be null, got %q", out.AuthenticationMethod.ValueString())
	}
	if !out.BaseUrl.IsNull() {
		t.Fatalf("expected base_url to be null, got %q", out.BaseUrl.ValueString())
	}
	if !out.ScimUrl.IsNull() {
		t.Fatalf("expected scim_url to be null, got %q", out.ScimUrl.ValueString())
	}
}
