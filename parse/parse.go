package parse

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
	"strings"

	"github.com/xhd2015/less-gen/go/gogen"
	"github.com/xhd2015/less-gen/strcase"
	"golang.org/x/tools/go/packages"
)

// FieldInfo represents a field in a struct
type FieldInfo struct {
	Name    string
	Type    string
	Tags    string
	Pointer bool
	Node    ast.Node `json:"-"`
}

// ModelInfo represents a model struct and its fields
type ModelInfo struct {
	Name       string
	Fields     []FieldInfo
	GenDecl    *ast.GenDecl    `json:"-"`
	TypeSpec   *ast.TypeSpec   `json:"-"`
	StructType *ast.StructType `json:"-"`
}

// FieldRelation represents a relation between a field and a column
type FieldRelation struct {
	FieldName  string
	ColumnName string
	Type       string
	IsPrimary  bool
	IsIndex    bool
	IsUnique   bool
}

// TableRelation represents a relation between a table and its models
type TableRelation struct {
	TablVarName   string
	TableName     string
	NeedCreateORM bool
	Model         ModelInfo
	OptionalModel ModelInfo
	Fields        []FieldRelation
}

// Package represents a Go package containing files with ORM table relations
type Package struct {
	PkgPath string
	Files   []*File
}

// File represents a Go file containing ORM table relations
type File struct {
	AbsFile     string
	HasGenerate bool
	Tables      []*TableRelation
	AST         *ast.File `json:"-"`
}

// ScanRelations loads Go packages from the specified directory with the given arguments
// and extracts table relations from orm.Bind calls.
// Returns a slice of packages containing files with table relations.
func ScanRelations(fset *token.FileSet, dir string, args []string) ([]*Package, error) {
	// Load packages
	pkgs, err := packages.Load(&packages.Config{
		Fset: fset,
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedDeps | packages.NeedName | packages.NeedImports | packages.NeedTypesInfo,
		Dir:  dir,
	}, args...)
	if err != nil {
		return nil, err
	}

	// Create the result structure
	var result []*Package

	// Process each package
	for _, pkg := range pkgs {
		var files []*File

		// Process each file in the package
		for _, file := range pkg.Syntax {
			filePath := pkg.Fset.Position(file.Pos()).Filename
			var tables []*TableRelation
			var hasGenerate bool

			// check go:generate
			genCmds := gogen.FindGoGenerate(file.Comments, []string{"go run github.com/xhd2015/arc-orm/cmd/arc-orm@latest gen", "go run github.com/xhd2015/arc-orm/cmd/arc-orm gen", "arc-orm gen"})
			if len(genCmds) > 0 {
				hasGenerate = true
			}

			// Process each declaration in the file
			for _, decl := range file.Decls {
				genDecl, ok := decl.(*ast.GenDecl)
				if !ok || genDecl.Tok != token.VAR {
					continue
				}

				// Process each variable specification
				for _, spec := range genDecl.Specs {
					varDecl, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}

					// Process each value in the variable declaration
					for _, value := range varDecl.Values {
						callExpr, ok := value.(*ast.CallExpr)
						if !ok {
							continue
						}

						// Extract table relation from the call expression
						relation, err := tryExtractORMTableRelation(callExpr, pkg)
						if err != nil {
							// Skip this call if there was an error
							continue
						}

						// Add the relation to our collection if it was extracted successfully
						if relation != nil {
							tables = append(tables, relation)
						}
					}
				}
			}

			// Only add file to the result if it has any relations
			if len(tables) > 0 {
				files = append(files, &File{
					AbsFile:     filePath,
					HasGenerate: hasGenerate,
					Tables:      tables,
					AST:         file,
				})
			} else {
				// no ORM found, find table
				firstTable, ident, _ := findFirstTableDef(pkg, file)
				if firstTable != "" {
					varDef := pkg.TypesInfo.Defs[ident]
					if varDef == nil {
						continue
					}
					tableVar, ok := varDef.(*types.Var)
					if !ok {
						continue
					}

					modelName := strcase.SnakeToCamel(pkg.Name)
					optionalModelName := modelName + "Optional"

					model := findModelInfoByName(pkg, modelName)
					optModel := findModelInfoByName(pkg, optionalModelName)

					// Extract field relations
					fields := extractFieldRelations(pkg, tableVar)

					files = append(files, &File{
						AbsFile:     filePath,
						HasGenerate: hasGenerate,
						Tables: []*TableRelation{{
							TablVarName:   ident.Name,
							TableName:     firstTable,
							Model:         model,
							OptionalModel: optModel,
							Fields:        fields,
							NeedCreateORM: true,
						}},
						AST: file,
					})
				}
			}
		}

		// Only add package to the result if it has any files with relations
		if len(files) > 0 {
			result = append(result, &Package{
				PkgPath: pkg.PkgPath,
				Files:   files,
			})
		}
	}

	return result, nil
}

// tryExtractORMTableRelation extracts a TableRelation from an orm.Bind call expression
func tryExtractORMTableRelation(callExpr *ast.CallExpr, pkg *packages.Package) (*TableRelation, error) {
	typeInfo := pkg.TypesInfo

	// First, check if we're dealing with an orm.Bind call
	var indices []ast.Expr
	var modelNames []string

	indexListExpr, ok := callExpr.Fun.(*ast.IndexListExpr)
	if !ok {
		return nil, nil
	}
	// Check if this is an orm.Bind call with type parameters (Go 1.18+)
	if indexListExpr.X == nil {
		return nil, nil
	}
	fn := indexListExpr.X

	var ident *ast.Ident
	if idt, ok := fn.(*ast.Ident); ok {
		ident = idt
	} else if sel, ok := fn.(*ast.SelectorExpr); ok {
		ident = sel.Sel
	}

	if ident == nil || ident.Name != "Bind" {
		return nil, nil
	}
	use := pkg.TypesInfo.Uses[ident]

	fnType, ok := use.(*types.Func)
	if !ok {
		return nil, nil // Not a function, skip
	}
	if fnType.Name() != "Bind" || fnType.Pkg().Path() != "github.com/xhd2015/arc-orm/orm" {
		return nil, nil // Not an orm.Bind call, skip
	}

	indices = indexListExpr.Indices
	// Try to extract model names from type indices
	for _, idx := range indices {
		if id, ok := idx.(*ast.Ident); ok {
			modelNames = append(modelNames, id.Name)
		}
	}

	// For type checking, use the expression type
	exprType := typeInfo.TypeOf(callExpr)
	if !isPtrToORM(exprType) {
		return nil, nil // Not a ptr to orm.ORM, skip
	}

	// Extract table name from the second argument
	if len(callExpr.Args) < 2 {
		return nil, nil
	}

	tableArg := callExpr.Args[1]
	tableVar, tableName := extractRefTableName(pkg, tableArg)
	if tableName == "" {
		return nil, nil
	}

	// Extract model types
	var model, optModel ModelInfo
	// If we have extracted model names from type indices
	if len(modelNames) >= 2 {
		// Look for matching structs in the package
		model = findModelInfoByName(pkg, modelNames[0])
		optModel = findModelInfoByName(pkg, modelNames[1])
	}

	// Extract field relations
	fields := extractFieldRelations(pkg, tableVar)

	// Create and return the table relation
	return &TableRelation{
		TableName:     tableName,
		Model:         model,
		OptionalModel: optModel,
		Fields:        fields,
	}, nil
}

// findModelInfoByName searches for a struct type with the given name in the package
func findModelInfoByName(pkg *packages.Package, name string) ModelInfo {
	info := ModelInfo{Name: name}

	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || typeSpec.Name.Name != name {
					continue
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				// Set the AST node for the model
				info.GenDecl = genDecl
				info.TypeSpec = typeSpec
				info.StructType = structType

				// Extract fields
				for _, field := range structType.Fields.List {
					if len(field.Names) == 0 {
						continue // Skip embedded fields
					}

					fieldName := field.Names[0].Name
					var typeName string
					var isPointer bool

					// Get the type
					switch t := field.Type.(type) {
					case *ast.Ident:
						typeName = t.Name
					case *ast.SelectorExpr:
						if x, ok := t.X.(*ast.Ident); ok {
							typeName = x.Name + "." + t.Sel.Name
						}
					case *ast.StarExpr:
						isPointer = true
						if ident, ok := t.X.(*ast.Ident); ok {
							typeName = ident.Name
						} else if sel, ok := t.X.(*ast.SelectorExpr); ok {
							if x, ok := sel.X.(*ast.Ident); ok {
								typeName = x.Name + "." + sel.Sel.Name
							}
						}
					}
					var tag string
					if field.Tag != nil {
						tagVal := field.Tag.Value
						if tagVal != "" && len(tagVal) >= 2 && strings.HasPrefix(tagVal, "`") && strings.HasSuffix(tagVal, "`") {
							tag = tagVal[1 : len(tagVal)-1]
						}
					}

					info.Fields = append(info.Fields, FieldInfo{
						Name:    fieldName,
						Type:    typeName,
						Tags:    tag,
						Pointer: isPointer,
						Node:    field,
					})
				}
			}
		}
	}

	return info
}

func isPtrToORM(typ types.Type) bool {
	ptrType, ok := typ.(*types.Pointer)
	if !ok {
		return false
	}

	namedType, ok := ptrType.Elem().(*types.Named)
	if !ok {
		return false
	}

	typeName := namedType.Obj()
	if typeName.Name() != "ORM" {
		return false
	}

	if typeName.Pkg().Path() != "github.com/xhd2015/arc-orm/orm" {
		return false
	}

	return true
}

// extractRefTableName tries to get the table name from an AST expression
func extractRefTableName(pkg *packages.Package, expr ast.Expr) (*types.Var, string) {
	var ident *ast.Ident
	if idt, ok := expr.(*ast.Ident); ok {
		ident = idt
	} else if sel, ok := expr.(*ast.SelectorExpr); ok {
		ident = sel.Sel
	}
	if ident == nil {
		return nil, ""
	}

	use := pkg.TypesInfo.Uses[ident]
	if use == nil {
		return nil, ""
	}
	useVar, ok := use.(*types.Var)
	if !ok {
		return nil, ""
	}
	tableName := extractTableFromVar(pkg, useVar)
	if tableName == "" {
		return nil, ""
	}
	return useVar, tableName
}

func extractTableFromVar(pkg *packages.Package, tableVar *types.Var) string {
	// typeInfo.Defs[]
	name := tableVar.Name()
	tablePkg := tableVar.Pkg()
	if tablePkg.Path() != pkg.PkgPath {
		return ""
	}

	_, value := findVarDef(pkg, name)
	return extractTableFromVarDef(pkg.TypesInfo, value)
}

func extractTableFromVarDef(typeInfo *types.Info, expr ast.Expr) string {
	if expr == nil {
		return ""
	}
	callExpr, ok := expr.(*ast.CallExpr)
	if !ok {
		return ""
	}
	if len(callExpr.Args) == 0 {
		return ""
	}

	var ident *ast.Ident
	if idt, ok := callExpr.Fun.(*ast.Ident); ok {
		ident = idt
	} else if sel, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
		ident = sel.Sel
	}
	if ident == nil || ident.Name != "New" {
		return ""
	}
	useFn := typeInfo.Uses[ident]
	if useFn == nil || useFn.Name() != "New" || useFn.Pkg().Path() != "github.com/xhd2015/arc-orm/table" {
		return ""
	}

	arg0 := callExpr.Args[0]
	// Extract table name from string literal
	if lit, ok := arg0.(*ast.BasicLit); ok && lit.Kind == token.STRING {
		unq, _ := strconv.Unquote(lit.Value)
		return unq
	}
	return ""
}

func forEachVarDef(pkg *packages.Package, fn func(file *ast.File, spec *ast.ValueSpec, name *ast.Ident, value ast.Expr) bool) {
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.VAR {
				continue
			}

			for _, spec := range genDecl.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}

				for i, valName := range valueSpec.Names {
					var value ast.Expr
					if i < len(valueSpec.Values) {
						value = valueSpec.Values[i]
					}
					fn(file, valueSpec, valName, value)
				}
			}
		}
	}
}

func findVarDef(pkg *packages.Package, name string) (*ast.Ident, ast.Expr) {
	var ident *ast.Ident
	var expr ast.Expr
	forEachVarDef(pkg, func(file *ast.File, spec *ast.ValueSpec, varName *ast.Ident, value ast.Expr) bool {
		if varName.Name == name {
			ident = varName
			expr = value
			return true
		}
		return false
	})
	return ident, expr
}

func findFirstTableDef(pkg *packages.Package, wantFile *ast.File) (string, *ast.Ident, ast.Expr) {
	var tableName string
	var ident *ast.Ident
	var expr ast.Expr
	forEachVarDef(pkg, func(file *ast.File, spec *ast.ValueSpec, varName *ast.Ident, value ast.Expr) bool {
		if file != wantFile {
			return false
		}
		resolvedName := extractTableFromVarDef(pkg.TypesInfo, value)
		if resolvedName != "" {
			tableName = resolvedName
			ident = varName
			expr = value
			return true
		}
		return false
	})
	return tableName, ident, expr
}

// extractModelInfo extracts information about a model struct from its type expression
func extractModelInfo(expr ast.Expr, typeInfo *types.Info, pkg *packages.Package) ModelInfo {
	info := ModelInfo{}

	// Get the type information
	typ := typeInfo.TypeOf(expr)
	if typ == nil {
		return info
	}

	// Get the named type
	var named *types.Named
	if ptr, ok := typ.(*types.Pointer); ok {
		if n, ok := ptr.Elem().(*types.Named); ok {
			named = n
		}
	} else if n, ok := typ.(*types.Named); ok {
		named = n
	}

	if named == nil {
		return info
	}

	// Get the struct type
	info.Name = named.Obj().Name()
	structType, ok := named.Underlying().(*types.Struct)
	if !ok {
		return info
	}

	// Extract fields
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		fieldType := field.Type()

		// Check if it's a pointer type
		isPointer := false
		if _, ok := fieldType.(*types.Pointer); ok {
			isPointer = true
		}

		fieldInfo := FieldInfo{
			Name:    field.Name(),
			Type:    fieldType.String(),
			Tags:    structType.Tag(i),
			Pointer: isPointer,
			Node:    nil,
		}
		info.Fields = append(info.Fields, fieldInfo)
	}

	return info
}

// extractFieldRelations finds field definitions in the package
func extractFieldRelations(pkg *packages.Package, tableVar *types.Var) []FieldRelation {
	var fields []FieldRelation

	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.VAR {
				continue
			}

			for _, spec := range genDecl.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}

				for i, name := range valueSpec.Names {
					if i >= len(valueSpec.Values) {
						continue
					}

					// Check if this is a field definition (like ID = Table.Int64("id"))
					callExpr, ok := valueSpec.Values[i].(*ast.CallExpr)
					if !ok {
						continue
					}

					// Check if the call is on a Table object
					selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
					if !ok {
						continue
					}

					// Check if the base is a reference to our table
					tableIdent, ok := selExpr.X.(*ast.Ident)
					if !ok {
						continue
					}
					useVar := pkg.TypesInfo.Uses[tableIdent]
					if useVar == nil {
						continue
					}
					if useVar != tableVar {
						continue
					}

					// Check if there's an argument for the column name
					if len(callExpr.Args) == 0 {
						continue
					}

					// Extract column name from string literal
					columnName := ""
					if lit, ok := callExpr.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
						columnName = strings.Trim(lit.Value, "\"")
					}

					if columnName == "" {
						continue
					}

					// Create field relation
					field := FieldRelation{
						FieldName:  name.Name,
						ColumnName: columnName,
						Type:       selExpr.Sel.Name,
						// For simplicity, we're not determining these yet
						IsPrimary: columnName == "id",
						IsIndex:   false,
						IsUnique:  false,
					}
					fields = append(fields, field)
				}
			}
		}
	}

	return fields
}
