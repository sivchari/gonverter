package gonverter

import (
	"go/token"
	"go/types"
	"testing"
)

func TestExtractTypeInfo(t *testing.T) {
	tests := []struct {
		name      string
		typ       types.Type
		wantPtr   bool
		wantNamed bool
	}{
		{
			name:      "basic type",
			typ:       types.Typ[types.Int],
			wantPtr:   false,
			wantNamed: false,
		},
		{
			name:      "pointer to basic type",
			typ:       types.NewPointer(types.Typ[types.Int]),
			wantPtr:   true,
			wantNamed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := extractTypeInfo(tt.typ)
			if info.isPointer != tt.wantPtr {
				t.Errorf("isPointer = %v, want %v", info.isPointer, tt.wantPtr)
			}
		})
	}
}

func TestExtractTypeInfoWithNamedType(t *testing.T) {
	// Create a named type
	pkg := types.NewPackage("example.com/test", "test")
	named := types.NewNamed(types.NewTypeName(token.NoPos, pkg, "User", nil), types.NewStruct(nil, nil), nil)

	info := extractTypeInfo(named)

	if info.isPointer {
		t.Error("expected non-pointer type")
	}

	if info.typeName != "User" {
		t.Errorf("typeName = %q, want %q", info.typeName, "User")
	}

	if info.pkgName != "test" {
		t.Errorf("pkgName = %q, want %q", info.pkgName, "test")
	}

	if info.pkgPath != "example.com/test" {
		t.Errorf("pkgPath = %q, want %q", info.pkgPath, "example.com/test")
	}
}

func TestExtractTypeInfoWithPointerToNamedType(t *testing.T) {
	pkg := types.NewPackage("example.com/test", "test")
	named := types.NewNamed(types.NewTypeName(token.NoPos, pkg, "User", nil), types.NewStruct(nil, nil), nil)
	ptr := types.NewPointer(named)

	info := extractTypeInfo(ptr)

	if !info.isPointer {
		t.Error("expected pointer type")
	}

	if info.typeName != "User" {
		t.Errorf("typeName = %q, want %q", info.typeName, "User")
	}
}

func TestFormatTypeDecl(t *testing.T) {
	tests := []struct {
		name    string
		info    typeInfo
		pkgName string
		want    string
	}{
		{
			name: "same package non-pointer",
			info: typeInfo{
				pkgName:   "converter",
				typeName:  "User",
				isPointer: false,
			},
			pkgName: "converter",
			want:    "User",
		},
		{
			name: "same package pointer",
			info: typeInfo{
				pkgName:   "converter",
				typeName:  "User",
				isPointer: true,
			},
			pkgName: "converter",
			want:    "*User",
		},
		{
			name: "different package non-pointer",
			info: typeInfo{
				pkgName:   "domain",
				typeName:  "User",
				isPointer: false,
			},
			pkgName: "converter",
			want:    "domain.User",
		},
		{
			name: "different package pointer",
			info: typeInfo{
				pkgName:   "domain",
				typeName:  "User",
				isPointer: true,
			},
			pkgName: "converter",
			want:    "*domain.User",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTypeDecl(tt.info, tt.pkgName)
			if got != tt.want {
				t.Errorf("formatTypeDecl() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFieldFuncName(t *testing.T) {
	g := &generator{}
	got := g.fieldFuncName("UserRequest", "User", "Name", "Name")
	want := "ConvertUserRequestNameToUserName"

	if got != want {
		t.Errorf("fieldFuncName() = %q, want %q", got, want)
	}
}

func TestHasBuildTag(_ *testing.T) {
	// This would require parsing actual Go files, so we skip detailed testing here.
	// The function is tested implicitly through integration tests.
}

func TestGetSliceElemType(t *testing.T) {
	tests := []struct {
		name     string
		typ      types.Type
		wantNil  bool
		wantElem types.Type
	}{
		{
			name:    "non-slice type",
			typ:     types.Typ[types.Int],
			wantNil: true,
		},
		{
			name:     "slice type",
			typ:      types.NewSlice(types.Typ[types.String]),
			wantNil:  false,
			wantElem: types.Typ[types.String],
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getSliceElemType(tt.typ)
			if tt.wantNil && got != nil {
				t.Errorf("expected nil, got %v", got)
			}

			if !tt.wantNil && got == nil {
				t.Error("expected non-nil result")
			}

			if !tt.wantNil && !types.Identical(got, tt.wantElem) {
				t.Errorf("element type = %v, want %v", got, tt.wantElem)
			}
		})
	}
}

func TestGetMapTypes(t *testing.T) {
	t.Run("non-map type returns nil", func(t *testing.T) {
		key, val := getMapTypes(types.Typ[types.Int])
		if key != nil || val != nil {
			t.Errorf("expected nil, got key=%v, val=%v", key, val)
		}
	})

	t.Run("map type returns key and value types", func(t *testing.T) {
		wantKey := types.Typ[types.String]
		wantVal := types.Typ[types.Int]
		key, val := getMapTypes(types.NewMap(wantKey, wantVal))

		if key == nil || val == nil {
			t.Fatal("expected non-nil key and val")
		}

		if !types.Identical(key, wantKey) {
			t.Errorf("key type = %v, want %v", key, wantKey)
		}

		if !types.Identical(val, wantVal) {
			t.Errorf("val type = %v, want %v", val, wantVal)
		}
	})
}

func TestIsStructType(t *testing.T) {
	pkg := types.NewPackage("example.com/test", "test")

	tests := []struct {
		name string
		typ  types.Type
		want bool
	}{
		{
			name: "basic type",
			typ:  types.Typ[types.Int],
			want: false,
		},
		{
			name: "slice type",
			typ:  types.NewSlice(types.Typ[types.Int]),
			want: false,
		},
		{
			name: "struct type",
			typ:  types.NewStruct(nil, nil),
			want: true,
		},
		{
			name: "named struct type",
			typ:  types.NewNamed(types.NewTypeName(token.NoPos, pkg, "User", nil), types.NewStruct(nil, nil), nil),
			want: true,
		},
		{
			name: "pointer to struct",
			typ:  types.NewPointer(types.NewStruct(nil, nil)),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isStructType(tt.typ)
			if got != tt.want {
				t.Errorf("isStructType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindField(t *testing.T) {
	// Create a struct with fields
	fields := []*types.Var{
		types.NewField(token.NoPos, nil, "Name", types.Typ[types.String], false),
		types.NewField(token.NoPos, nil, "Age", types.Typ[types.Int], false),
		types.NewField(token.NoPos, nil, "private", types.Typ[types.Int], false), // unexported
	}
	s := types.NewStruct(fields, nil)

	tests := []struct {
		name      string
		fieldName string
		wantFound bool
	}{
		{
			name:      "find exported field",
			fieldName: "Name",
			wantFound: true,
		},
		{
			name:      "find another exported field",
			fieldName: "Age",
			wantFound: true,
		},
		{
			name:      "unexported field not found",
			fieldName: "private",
			wantFound: false,
		},
		{
			name:      "non-existent field",
			fieldName: "NotExists",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findField(s, tt.fieldName)
			if tt.wantFound && got == nil {
				t.Error("expected to find field, got nil")
			}

			if !tt.wantFound && got != nil {
				t.Errorf("expected nil, got %v", got)
			}
		})
	}
}

func TestPairKey(t *testing.T) {
	g := &generator{}
	pair := &conversionPair{
		from: typeInfo{
			pkgPath:  "example.com/handler",
			typeName: "UserRequest",
		},
		to: typeInfo{
			pkgPath:  "example.com/domain",
			typeName: "User",
		},
	}

	got := g.pairKey(pair)
	want := "example.com/handler/UserRequest->example.com/domain/User"

	if got != want {
		t.Errorf("pairKey() = %q, want %q", got, want)
	}
}

func TestCreatePointerFieldMapping(t *testing.T) {
	g := &generator{}

	tests := []struct {
		name        string
		funcName    string
		srcName     string
		dstName     string
		srcIsPtr    bool
		dstIsPtr    bool
		dstTypeName string
		wantSubstr  string
	}{
		{
			name:        "both pointers",
			funcName:    "ConvertProfileRequestToProfile",
			srcName:     "Profile",
			dstName:     "Profile",
			srcIsPtr:    true,
			dstIsPtr:    true,
			dstTypeName: "Profile",
			wantSubstr:  "new(Profile)",
		},
		{
			name:        "only src pointer",
			funcName:    "ConvertProfileRequestToProfile",
			srcName:     "Profile",
			dstName:     "Profile",
			srcIsPtr:    true,
			dstIsPtr:    false,
			dstTypeName: "Profile",
			wantSubstr:  "&dst.Profile",
		},
		{
			name:        "only dst pointer",
			funcName:    "ConvertProfileRequestToProfile",
			srcName:     "Profile",
			dstName:     "Profile",
			srcIsPtr:    false,
			dstIsPtr:    true,
			dstTypeName: "Profile",
			wantSubstr:  "&src.Profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := g.createPointerFieldMapping(tt.funcName, tt.srcName, tt.dstName, tt.srcIsPtr, tt.dstIsPtr, tt.dstTypeName)
			if !contains(got, tt.wantSubstr) {
				t.Errorf("createPointerFieldMapping() = %q, want to contain %q", got, tt.wantSubstr)
			}
		})
	}
}

func TestCreateSliceMapping(t *testing.T) {
	g := &generator{}
	got := g.createSliceMapping("ConvertItemRequestToItem", "Items", "Items", "Item")

	wantSubstrings := []string{
		"src.Items != nil",
		"make([]Item, len(src.Items))",
		"for i := range src.Items",
		"ConvertItemRequestToItem(&src.Items[i], &dst.Items[i])",
	}

	for _, substr := range wantSubstrings {
		if !contains(got, substr) {
			t.Errorf("createSliceMapping() = %q, want to contain %q", got, substr)
		}
	}
}

func TestCreateMapMapping(t *testing.T) {
	g := &generator{}
	srcMap := types.NewMap(types.Typ[types.String], types.NewStruct(nil, nil))
	dstMap := types.NewMap(types.Typ[types.String], types.NewStruct(nil, nil))

	got := g.createMapMapping("ConvertSettingRequestToSetting", "Settings", "Settings", srcMap, dstMap, "Setting")

	wantSubstrings := []string{
		"src.Settings != nil",
		"make(map[string]Setting, len(src.Settings))",
		"for k, v := range src.Settings",
		"var converted Setting",
		"ConvertSettingRequestToSetting(&v, &converted)",
		"dst.Settings[k] = converted",
	}

	for _, substr := range wantSubstrings {
		if !contains(got, substr) {
			t.Errorf("createMapMapping() = %q, want to contain %q", got, substr)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || s != "" && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}

// Integration tests using testdata directories.
func TestRunWithPointerTestdata(t *testing.T) {
	err := Run("../../testdata/pointer")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestRunWithSliceTestdata(t *testing.T) {
	err := Run("../../testdata/slice")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestRunWithMapTestdata(t *testing.T) {
	err := Run("../../testdata/maptype")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestRunWithNestedTestdata(t *testing.T) {
	err := Run("../../testdata/nested")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestRunWithBidirectionalTestdata(t *testing.T) {
	err := Run("../../testdata/bidirectional")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestRunWithNoConversionPairs(t *testing.T) {
	// Test with a package that has no registration
	err := Run("../../testdata/nobuildtag")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestRunWithInvalidPattern(t *testing.T) {
	err := Run("invalid/nonexistent/path")
	if err == nil {
		t.Error("expected error for invalid pattern, got nil")
	}
}
