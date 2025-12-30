package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPropagationStoreLookupModeFromValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		id            types.String
		storeName     types.String
		storeType     types.String
		wantMode      propagationStoreLookupMode
		wantErrCount  int
		wantErrPath   path.Path
		wantErrIsPath bool
	}{
		{
			name:         "id_only",
			id:           types.StringValue("store-123"),
			storeName:    types.StringNull(),
			storeType:    types.StringNull(),
			wantMode:     propagationStoreLookupModeId,
			wantErrCount: 0,
		},
		{
			name:         "name_type_only",
			id:           types.StringNull(),
			storeName:    types.StringValue("PingOne Directory"),
			storeType:    types.StringValue("directory"),
			wantMode:     propagationStoreLookupModeNameType,
			wantErrCount: 0,
		},
		{
			name:          "name_only_errors",
			id:            types.StringNull(),
			storeName:     types.StringValue("PingOne Directory"),
			storeType:     types.StringNull(),
			wantMode:      propagationStoreLookupModeInvalid,
			wantErrCount:  1,
			wantErrPath:   path.Root("type"),
			wantErrIsPath: true,
		},
		{
			name:          "type_only_errors",
			id:            types.StringNull(),
			storeName:     types.StringNull(),
			storeType:     types.StringValue("directory"),
			wantMode:      propagationStoreLookupModeInvalid,
			wantErrCount:  1,
			wantErrPath:   path.Root("name"),
			wantErrIsPath: true,
		},
		{
			name:         "nothing_set_errors",
			id:           types.StringNull(),
			storeName:    types.StringNull(),
			storeType:    types.StringNull(),
			wantMode:     propagationStoreLookupModeInvalid,
			wantErrCount: 1,
		},
		{
			name:         "id_and_name_type_prefers_name_type",
			id:           types.StringValue("store-123"),
			storeName:    types.StringValue("PingOne Directory"),
			storeType:    types.StringValue("directory"),
			wantMode:     propagationStoreLookupModeNameType,
			wantErrCount: 0,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotMode, gotDiags := propagationStoreLookupModeFromValues(tt.id, tt.storeName, tt.storeType)

			if gotMode != tt.wantMode {
				t.Fatalf("mode mismatch: got %v want %v", gotMode, tt.wantMode)
			}

			if gotDiags.ErrorsCount() != tt.wantErrCount {
				t.Fatalf("diagnostic count mismatch: got %d want %d", gotDiags.ErrorsCount(), tt.wantErrCount)
			}

			if !tt.wantErrIsPath || tt.wantErrCount == 0 {
				return
			}

			var withPath diag.DiagnosticWithPath
			for _, d := range gotDiags.Errors() {
				if dp, ok := d.(diag.DiagnosticWithPath); ok {
					withPath = dp
					break
				}
			}

			if withPath == nil {
				t.Fatalf("expected at least one path-based diagnostic, got none")
			}

			if withPath.Path().String() != tt.wantErrPath.String() {
				t.Fatalf("diagnostic path mismatch: got %q want %q", withPath.Path().String(), tt.wantErrPath.String())
			}
		})
	}
}
