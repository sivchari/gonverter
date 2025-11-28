package gonverter

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
)

const (
	runtimePkgSuffix = "gonverter/runtime"
	convertPrefix    = "Convert"
	buildTag         = "gonverter"
)

// Run executes the code generation for the given package pattern.
func Run(pattern string) error {
	g := &generator{
		fset:           token.NewFileSet(),
		customFuncs:    make(map[string]bool),
		generatedPairs: make(map[string]bool),
	}

	return g.run(pattern)
}

type generator struct {
	fset           *token.FileSet
	customFuncs    map[string]bool
	generatedPairs map[string]bool // tracks already generated conversion pairs
}

func (g *generator) run(pattern string) error {
	pairs, pkgDir, err := g.parse(pattern)
	if err != nil {
		return err
	}

	if len(pairs) == 0 {
		fmt.Println("No conversion pairs found")

		return nil
	}

	fmt.Printf("Found %d conversion pair(s)\n", len(pairs))

	if err := g.detectCustomFuncs(pattern); err != nil {
		return err
	}

	code, err := g.generate(pairs, filepath.Base(pkgDir))
	if err != nil {
		return err
	}

	outputPath := filepath.Join(pkgDir, "generated.go")
	if err := os.WriteFile(outputPath, code, 0o600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("Generated: %s\n", outputPath)

	return nil
}

// --- Parsing ---

type conversionPair struct {
	from, to typeInfo
}

type typeInfo struct {
	pkgPath   string
	pkgName   string
	typeName  string
	typ       types.Type
	isPointer bool
}

func (g *generator) parse(pattern string) ([]conversionPair, string, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedImports | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
		Fset:       g.fset,
		BuildFlags: []string{"-tags=gonverter"},
	}

	pkgs, err := packages.Load(cfg, pattern)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load packages: %w", err)
	}

	if packages.PrintErrors(pkgs) > 0 {
		return nil, "", fmt.Errorf("packages contain errors")
	}

	var pairs []conversionPair

	var pkgDir string

	for _, pkg := range pkgs {
		if len(pkg.GoFiles) > 0 && pkgDir == "" {
			pkgDir = filepath.Dir(pkg.GoFiles[0])
		}

		for _, file := range pkg.Syntax {
			if !hasBuildTag(file, buildTag) {
				continue
			}

			pairs = append(pairs, g.extractPairs(pkg, file)...)
		}
	}

	return pairs, pkgDir, nil
}

func (g *generator) extractPairs(pkg *packages.Package, file *ast.File) []conversionPair {
	var pairs []conversionPair

	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		indexExpr, ok := call.Fun.(*ast.IndexListExpr)
		if !ok || len(indexExpr.Indices) != 2 {
			return true
		}

		callType := g.getRegisterCallType(pkg, indexExpr)
		if callType == registerCallNone {
			return true
		}

		fromType := pkg.TypesInfo.TypeOf(indexExpr.Indices[0])
		toType := pkg.TypesInfo.TypeOf(indexExpr.Indices[1])

		if fromType == nil || toType == nil {
			return true
		}

		// Add forward conversion (From → To)
		pairs = append(pairs, conversionPair{
			from: extractTypeInfo(fromType),
			to:   extractTypeInfo(toType),
		})

		// Add reverse conversion (To → From) for bidirectional registration
		if callType == registerCallBidirectional {
			pairs = append(pairs, conversionPair{
				from: extractTypeInfo(toType),
				to:   extractTypeInfo(fromType),
			})
		}

		return true
	})

	return pairs
}

type registerCallType int

const (
	registerCallNone registerCallType = iota
	registerCallUnidirectional
	registerCallBidirectional
)

func (g *generator) getRegisterCallType(pkg *packages.Package, indexExpr *ast.IndexListExpr) registerCallType {
	sel, ok := indexExpr.X.(*ast.SelectorExpr)
	if !ok {
		return registerCallNone
	}

	// Check for Register or RegisterBidirectional
	if sel.Sel.Name != "Register" && sel.Sel.Name != "RegisterBidirectional" {
		return registerCallNone
	}

	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return registerCallNone
	}

	obj := pkg.TypesInfo.ObjectOf(ident)
	if obj == nil {
		return registerCallNone
	}

	pkgName, ok := obj.(*types.PkgName)
	if !ok || !strings.HasSuffix(pkgName.Imported().Path(), runtimePkgSuffix) {
		return registerCallNone
	}

	if sel.Sel.Name == "RegisterBidirectional" {
		return registerCallBidirectional
	}

	return registerCallUnidirectional
}

func extractTypeInfo(t types.Type) typeInfo {
	info := typeInfo{typ: t}

	if ptr, ok := t.(*types.Pointer); ok {
		info.isPointer = true
		t = ptr.Elem()
	}

	if named, ok := t.(*types.Named); ok {
		obj := named.Obj()
		info.typeName = obj.Name()

		if pkg := obj.Pkg(); pkg != nil {
			info.pkgPath = pkg.Path()
			info.pkgName = pkg.Name()
		}
	}

	return info
}

func hasBuildTag(file *ast.File, tag string) bool {
	for _, cg := range file.Comments {
		for _, c := range cg.List {
			text := strings.TrimSpace(strings.TrimPrefix(c.Text, "//"))
			if (strings.HasPrefix(text, "go:build") || strings.HasPrefix(text, "+build")) &&
				strings.Contains(text, tag) {
				return true
			}
		}
	}

	return false
}

// --- Custom function detection ---

func (g *generator) detectCustomFuncs(pattern string) error {
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedSyntax | packages.NeedFiles,
	}, pattern)
	if err != nil {
		return fmt.Errorf("failed to load packages: %w", err)
	}

	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				if fn, ok := n.(*ast.FuncDecl); ok && strings.HasPrefix(fn.Name.Name, convertPrefix) {
					g.customFuncs[fn.Name.Name] = true
				}

				return true
			})
		}
	}

	return nil
}

// --- Code generation ---

type templateData struct {
	PackageName string
	Imports     []string
	Funcs       []funcData
}

type funcData struct {
	Name         string
	SrcTypeName  string
	DstTypeName  string
	SrcTypeDecl  string
	DstTypeDecl  string
	SrcIsPointer bool
	Mappings     []string
}

func (g *generator) generate(pairs []conversionPair, pkgName string) ([]byte, error) {
	data := templateData{PackageName: pkgName}
	imports := make(map[string]bool)

	// Process pairs including nested structs (use queue to handle discovered nested pairs)
	queue := append([]conversionPair{}, pairs...)

	for len(queue) > 0 {
		pair := queue[0]
		queue = queue[1:]

		// Skip if already generated
		pairKey := g.pairKey(&pair)
		if g.generatedPairs[pairKey] {
			continue
		}

		g.generatedPairs[pairKey] = true

		fd, nestedPairs, err := g.buildFuncDataWithNested(&pair, pkgName, imports)
		if err != nil {
			return nil, err
		}

		data.Funcs = append(data.Funcs, fd)

		// Add discovered nested pairs to queue
		queue = append(queue, nestedPairs...)
	}

	for imp := range imports {
		data.Imports = append(data.Imports, imp)
	}

	sort.Strings(data.Imports)

	tmpl, err := template.New("converter").Parse(converterTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return buf.Bytes(), fmt.Errorf("failed to format: %w", err)
	}

	return formatted, nil
}

func (g *generator) pairKey(pair *conversionPair) string {
	return fmt.Sprintf("%s/%s->%s/%s", pair.from.pkgPath, pair.from.typeName, pair.to.pkgPath, pair.to.typeName)
}

func (g *generator) buildFuncDataWithNested(pair *conversionPair, pkgName string, imports map[string]bool) (funcData, []conversionPair, error) {
	fd := funcData{
		Name:         fmt.Sprintf("Convert%sTo%s", pair.from.typeName, pair.to.typeName),
		SrcTypeName:  pair.from.typeName,
		DstTypeName:  pair.to.typeName,
		SrcTypeDecl:  formatTypeDecl(pair.from, pkgName),
		DstTypeDecl:  formatTypeDecl(pair.to, pkgName),
		SrcIsPointer: pair.from.isPointer,
	}

	// Collect imports
	if pair.from.pkgPath != "" && pair.from.pkgName != pkgName {
		imports[pair.from.pkgPath] = true
	}

	if pair.to.pkgPath != "" && pair.to.pkgName != pkgName {
		imports[pair.to.pkgPath] = true
	}

	// Build mappings and collect nested pairs
	mappings, nestedPairs, err := g.buildMappingsWithNested(pair)
	if err != nil {
		return fd, nil, err
	}

	fd.Mappings = mappings

	return fd, nestedPairs, nil
}

func (g *generator) buildMappingsWithNested(pair *conversionPair) ([]string, []conversionPair, error) {
	fromType := pair.from.typ
	if ptr, ok := fromType.(*types.Pointer); ok {
		fromType = ptr.Elem()
	}

	fromStruct, ok := fromType.Underlying().(*types.Struct)
	if !ok {
		return nil, nil, fmt.Errorf("from type is not a struct: %v", pair.from.typ)
	}

	toType := pair.to.typ
	if ptr, ok := toType.(*types.Pointer); ok {
		toType = ptr.Elem()
	}

	toStruct, ok := toType.Underlying().(*types.Struct)
	if !ok {
		return nil, nil, fmt.Errorf("to type is not a struct: %v", pair.to.typ)
	}

	var mappings []string

	var nestedPairs []conversionPair

	for i := 0; i < toStruct.NumFields(); i++ {
		dstField := toStruct.Field(i)
		if !dstField.Exported() {
			continue
		}

		srcField := findField(fromStruct, dstField.Name())
		mapping, nested := g.createMappingWithNested(pair, srcField, dstField)
		mappings = append(mappings, mapping)

		if nested != nil {
			nestedPairs = append(nestedPairs, *nested)
		}
	}

	return mappings, nestedPairs, nil
}

func (g *generator) createMappingWithNested(pair *conversionPair, srcField, dstField *types.Var) (string, *conversionPair) {
	dstName := dstField.Name()

	// No matching source field -> custom function
	if srcField == nil {
		funcName := g.fieldFuncName(pair.from.typeName, pair.to.typeName, dstName, dstName)

		return fmt.Sprintf("%s(src, dst)", funcName), nil
	}

	srcName := srcField.Name()

	// Same type -> direct assignment or custom if exists
	if types.Identical(srcField.Type(), dstField.Type()) {
		return g.handleIdenticalTypes(pair, srcName, dstName)
	}

	// Check if both fields are slices of structs
	if mapping, nested := g.handleSliceField(pair, srcField, dstField, srcName, dstName); mapping != "" {
		return mapping, nested
	}

	// Check if both fields are maps with struct values
	if mapping, nested := g.handleMapField(pair, srcField, dstField, srcName, dstName); mapping != "" {
		return mapping, nested
	}

	// Check if both fields are structs (nested struct case)
	if mapping, nested := g.handleStructField(pair, srcField, dstField, srcName, dstName); mapping != "" {
		return mapping, nested
	}

	// Different type (non-struct) -> custom function
	funcName := g.fieldFuncName(pair.from.typeName, pair.to.typeName, srcName, dstName)

	return fmt.Sprintf("%s(src, dst)", funcName), nil
}

func (g *generator) handleIdenticalTypes(pair *conversionPair, srcName, dstName string) (string, *conversionPair) {
	funcName := g.fieldFuncName(pair.from.typeName, pair.to.typeName, srcName, dstName)
	if g.customFuncs[funcName] {
		return fmt.Sprintf("%s(src, dst)", funcName), nil
	}

	return fmt.Sprintf("dst.%s = src.%s", dstName, srcName), nil
}

func (g *generator) handleSliceField(pair *conversionPair, srcField, dstField *types.Var, srcName, dstName string) (string, *conversionPair) {
	srcSlice, dstSlice := getSliceElemType(srcField.Type()), getSliceElemType(dstField.Type())
	if srcSlice == nil || dstSlice == nil || !isStructType(srcSlice) || !isStructType(dstSlice) {
		return "", nil
	}

	srcElemInfo := extractTypeInfo(srcSlice)
	dstElemInfo := extractTypeInfo(dstSlice)
	funcName := fmt.Sprintf("Convert%sTo%s", srcElemInfo.typeName, dstElemInfo.typeName)

	// Check if custom function exists
	fieldFuncName := g.fieldFuncName(pair.from.typeName, pair.to.typeName, srcName, dstName)
	if g.customFuncs[fieldFuncName] {
		return fmt.Sprintf("%s(src, dst)", fieldFuncName), nil
	}

	nestedPair := &conversionPair{
		from: typeInfo{
			pkgPath:   srcElemInfo.pkgPath,
			pkgName:   srcElemInfo.pkgName,
			typeName:  srcElemInfo.typeName,
			typ:       srcSlice,
			isPointer: true,
		},
		to: typeInfo{
			pkgPath:   dstElemInfo.pkgPath,
			pkgName:   dstElemInfo.pkgName,
			typeName:  dstElemInfo.typeName,
			typ:       dstSlice,
			isPointer: true,
		},
	}

	return g.createSliceMapping(funcName, srcName, dstName, dstElemInfo.typeName), nestedPair
}

func (g *generator) handleMapField(pair *conversionPair, srcField, dstField *types.Var, srcName, dstName string) (string, *conversionPair) {
	_, srcMapVal := getMapTypes(srcField.Type())
	_, dstMapVal := getMapTypes(dstField.Type())

	if srcMapVal == nil || dstMapVal == nil || !isStructType(srcMapVal) || !isStructType(dstMapVal) {
		return "", nil
	}

	srcValInfo := extractTypeInfo(srcMapVal)
	dstValInfo := extractTypeInfo(dstMapVal)
	funcName := fmt.Sprintf("Convert%sTo%s", srcValInfo.typeName, dstValInfo.typeName)

	// Check if custom function exists
	fieldFuncName := g.fieldFuncName(pair.from.typeName, pair.to.typeName, srcName, dstName)
	if g.customFuncs[fieldFuncName] {
		return fmt.Sprintf("%s(src, dst)", fieldFuncName), nil
	}

	nestedPair := &conversionPair{
		from: typeInfo{
			pkgPath:   srcValInfo.pkgPath,
			pkgName:   srcValInfo.pkgName,
			typeName:  srcValInfo.typeName,
			typ:       srcMapVal,
			isPointer: true,
		},
		to: typeInfo{
			pkgPath:   dstValInfo.pkgPath,
			pkgName:   dstValInfo.pkgName,
			typeName:  dstValInfo.typeName,
			typ:       dstMapVal,
			isPointer: true,
		},
	}

	return g.createMapMapping(funcName, srcName, dstName, srcField.Type(), dstField.Type(), dstValInfo.typeName), nestedPair
}

func (g *generator) handleStructField(pair *conversionPair, srcField, dstField *types.Var, srcName, dstName string) (string, *conversionPair) {
	if !isStructType(srcField.Type()) || !isStructType(dstField.Type()) {
		return "", nil
	}

	srcInfo := extractTypeInfo(srcField.Type())
	dstInfo := extractTypeInfo(dstField.Type())
	funcName := fmt.Sprintf("Convert%sTo%s", srcInfo.typeName, dstInfo.typeName)

	// Check if custom function exists
	fieldFuncName := g.fieldFuncName(pair.from.typeName, pair.to.typeName, srcName, dstName)
	if g.customFuncs[fieldFuncName] {
		return fmt.Sprintf("%s(src, dst)", fieldFuncName), nil
	}

	// Create nested pair for generation (always use pointer for nested struct conversion)
	nestedPair := &conversionPair{
		from: typeInfo{
			pkgPath:   srcInfo.pkgPath,
			pkgName:   srcInfo.pkgName,
			typeName:  srcInfo.typeName,
			typ:       srcInfo.typ,
			isPointer: true,
		},
		to: typeInfo{
			pkgPath:   dstInfo.pkgPath,
			pkgName:   dstInfo.pkgName,
			typeName:  dstInfo.typeName,
			typ:       dstInfo.typ,
			isPointer: true,
		},
	}

	// For pointer fields, need nil check and allocation
	if srcInfo.isPointer || dstInfo.isPointer {
		return g.createPointerFieldMapping(funcName, srcName, dstName, srcInfo.isPointer, dstInfo.isPointer, dstInfo.typeName), nestedPair
	}

	// Determine how to pass the field (with or without &)
	srcExpr := fmt.Sprintf("&src.%s", srcName)
	dstExpr := fmt.Sprintf("&dst.%s", dstName)

	if srcInfo.isPointer {
		srcExpr = fmt.Sprintf("src.%s", srcName)
	}

	if dstInfo.isPointer {
		dstExpr = fmt.Sprintf("dst.%s", dstName)
	}

	return fmt.Sprintf("%s(%s, %s)", funcName, srcExpr, dstExpr), nestedPair
}

// createPointerFieldMapping creates mapping code for pointer struct fields.
func (g *generator) createPointerFieldMapping(funcName, srcName, dstName string, srcIsPtr, dstIsPtr bool, dstTypeName string) string {
	// Both are pointers: if src != nil, allocate dst and convert
	if srcIsPtr && dstIsPtr {
		return fmt.Sprintf(`if src.%s != nil {
		dst.%s = new(%s)
		%s(src.%s, dst.%s)
	}`, srcName, dstName, dstTypeName, funcName, srcName, dstName)
	}

	// Only src is pointer: if src != nil, convert to non-pointer dst
	if srcIsPtr {
		return fmt.Sprintf(`if src.%s != nil {
		%s(src.%s, &dst.%s)
	}`, srcName, funcName, srcName, dstName)
	}

	// Only dst is pointer: allocate dst and convert
	return fmt.Sprintf(`dst.%s = new(%s)
	%s(&src.%s, dst.%s)`, dstName, dstTypeName, funcName, srcName, dstName)
}

// createSliceMapping creates mapping code for slice fields.
func (g *generator) createSliceMapping(funcName, srcName, dstName, dstElemTypeName string) string {
	return fmt.Sprintf(`if src.%s != nil {
		dst.%s = make([]%s, len(src.%s))
		for i := range src.%s {
			%s(&src.%s[i], &dst.%s[i])
		}
	}`, srcName, dstName, dstElemTypeName, srcName, srcName, funcName, srcName, dstName)
}

// createMapMapping creates mapping code for map fields.
func (g *generator) createMapMapping(funcName, srcName, dstName string, srcType, _ types.Type, dstValTypeName string) string {
	// Get key type string
	srcMap, ok := srcType.Underlying().(*types.Map)
	if !ok {
		return fmt.Sprintf("// Error: %s is not a map type", srcName)
	}

	keyTypeStr := srcMap.Key().String()

	return fmt.Sprintf(`if src.%s != nil {
		dst.%s = make(map[%s]%s, len(src.%s))
		for k, v := range src.%s {
			var converted %s
			%s(&v, &converted)
			dst.%s[k] = converted
		}
	}`, srcName, dstName, keyTypeStr, dstValTypeName, srcName, srcName, dstValTypeName, funcName, dstName)
}

// getSliceElemType returns the element type if t is a slice, otherwise nil.
func getSliceElemType(t types.Type) types.Type {
	if slice, ok := t.Underlying().(*types.Slice); ok {
		return slice.Elem()
	}

	return nil
}

// getMapTypes returns (keyType, valueType) if t is a map, otherwise (nil, nil).
func getMapTypes(t types.Type) (types.Type, types.Type) {
	if m, ok := t.Underlying().(*types.Map); ok {
		return m.Key(), m.Elem()
	}

	return nil, nil
}

// isStructType checks if the type is a struct (including named struct types).
func isStructType(t types.Type) bool {
	// Unwrap pointer
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}

	// Check if underlying type is struct
	_, ok := t.Underlying().(*types.Struct)

	return ok
}

func (g *generator) fieldFuncName(srcType, dstType, srcField, dstField string) string {
	return fmt.Sprintf("Convert%s%sTo%s%s", srcType, srcField, dstType, dstField)
}

func findField(s *types.Struct, name string) *types.Var {
	for i := 0; i < s.NumFields(); i++ {
		if f := s.Field(i); f.Name() == name && f.Exported() {
			return f
		}
	}

	return nil
}

func formatTypeDecl(info typeInfo, pkgName string) string {
	ptr := ""
	if info.isPointer {
		ptr = "*"
	}

	if info.pkgName != "" && info.pkgName != pkgName {
		return fmt.Sprintf("%s%s.%s", ptr, info.pkgName, info.typeName)
	}

	return ptr + info.typeName
}
