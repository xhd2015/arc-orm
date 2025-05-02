package parse

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

// FieldInfo represents a field in a struct
type FieldInfo struct {
	Name    string
	Type    string
	Tags    map[string]string
	Pointer bool
	Node    ast.Node `json:"-"`
}

// ModelInfo represents a model struct and its fields
type ModelInfo struct {
	Name   string
	Fields []FieldInfo
	Node   ast.Node `json:"-"`
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
	TableName     string
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
	AbsFile string
	Tables  []*TableRelation
	AST     *ast.File `json:"-"`
}

// LoadAndExtractRelations loads Go packages from the specified directory with the given arguments
// and extracts table relations from orm.Bind calls.
// Returns a slice of packages containing files with table relations.
func LoadAndExtractRelations(fset *token.FileSet, dir string, args []string) ([]*Package, error) {
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
						relation, err := extractTableRelation(callExpr, pkg)
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
					AbsFile: filePath,
					Tables:  tables,
					AST:     file,
				})
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

// extractTableRelation extracts a TableRelation from an orm.Bind call expression
func extractTableRelation(callExpr *ast.CallExpr, pkg *packages.Package) (*TableRelation, error) {
	typeInfo := pkg.TypesInfo

	// First, check if we're dealing with an orm.Bind call
	var isBind bool
	var indices []ast.Expr
	var modelNames []string

	// Check if this is an orm.Bind call with type parameters (Go 1.18+)
	if indexListExpr, ok := callExpr.Fun.(*ast.IndexListExpr); ok {
		fn := indexListExpr.X
		if fn == nil {
			return nil, nil // No function expression, skip
		}

		var ident *ast.Ident
		var pkgName string
		if idt, ok := fn.(*ast.Ident); ok {
			ident = idt
		} else if sel, ok := fn.(*ast.SelectorExpr); ok {
			ident = sel.Sel
			if x, ok := sel.X.(*ast.Ident); ok {
				pkgName = x.Name
			}
		} else {
			return nil, nil // Not a recognized function call pattern, skip
		}

		if ident.Name == "Bind" && (pkgName == "" || pkgName == "orm") {
			isBind = true
			indices = indexListExpr.Indices

			// Try to extract model names from type indices
			for _, idx := range indices {
				if id, ok := idx.(*ast.Ident); ok {
					modelNames = append(modelNames, id.Name)
				}
			}
		}
	} else if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
		// Check for orm.Bind (non-generic or older Go versions)
		var pkgName string
		if x, ok := selExpr.X.(*ast.Ident); ok {
			pkgName = x.Name
		}
		if selExpr.Sel.Name == "Bind" {
			if pkgName == "orm" {
				isBind = true
			}
		}
	}

	if !isBind {
		return nil, nil // Not a Bind call, skip
	}

	// For type checking, use the expression type
	exprType := typeInfo.TypeOf(callExpr)
	if exprType == nil {
		// Continue without type information, relying on syntax checks
	} else {
		if !isOrmBind(exprType) {
			// Continue anyway since we've identified it syntactically as orm.Bind
		}
	}

	// Extract table name from the second argument
	if len(callExpr.Args) < 2 {
		return nil, fmt.Errorf("Bind call missing arguments")
	}

	tableArg := callExpr.Args[1]
	tableName := extractTableName(tableArg, typeInfo, pkg)
	if tableName == "" {
		return nil, fmt.Errorf("unable to extract table name")
	}

	// Extract model types
	var model, optModel ModelInfo

	// If we have extracted model names from type indices
	if len(modelNames) >= 2 {
		// Look for matching structs in the package
		model = findModelInfoByName(pkg, modelNames[0])
		optModel = findModelInfoByName(pkg, modelNames[1])
	} else if len(indices) >= 2 {
		// Use type indices with type information
		modelType := indices[0]
		optModelType := indices[1]
		model = extractModelInfo(modelType, typeInfo, pkg)
		optModel = extractModelInfo(optModelType, typeInfo, pkg)
	} else {
		// Try to find model structs based on naming conventions
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				typeSpec, ok := n.(*ast.TypeSpec)
				if !ok {
					return true
				}
				if _, ok := typeSpec.Type.(*ast.StructType); !ok {
					return true
				}

				if strings.HasSuffix(typeSpec.Name.Name, "Optional") {
					// Found optional model
					if optModel.Name == "" {
						optModel = findModelInfoByName(pkg, typeSpec.Name.Name)
					}
				} else {
					// Regular model
					if model.Name == "" {
						model = findModelInfoByName(pkg, typeSpec.Name.Name)
					}
				}
				return true
			})
		}

		if model.Name == "" {
			model = ModelInfo{Name: "Model"}
		}
		if optModel.Name == "" {
			optModel = ModelInfo{Name: "OptionalModel"}
		}
	}

	// Extract field relations
	fields := extractFieldRelations(pkg, tableName)

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
		ast.Inspect(file, func(n ast.Node) bool {
			typeSpec, ok := n.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != name {
				return true
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				return false
			}

			// Set the AST node for the model
			info.Node = typeSpec

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

				// Parse tags
				tags := make(map[string]string)
				if field.Tag != nil {
					// We'd parse tags here in a production implementation
				}

				info.Fields = append(info.Fields, FieldInfo{
					Name:    fieldName,
					Type:    typeName,
					Tags:    tags,
					Pointer: isPointer,
					Node:    field,
				})
			}

			// We found the struct, no need to look further
			return false
		})
	}

	return info
}

func isOrmBind(typ types.Type) bool {
	ptrType, ok := typ.(*types.Pointer)
	if !ok {
		return false
	}

	namedType, ok := ptrType.Elem().(*types.Named)
	if !ok {
		return false
	}

	typeName := namedType.Obj()

	// Check for github.com/xhd2015/ormx/orm.ORM since we're working in the ormx repo
	if typeName.Pkg().Path() == "github.com/xhd2015/ormx/orm" && typeName.Name() == "ORM" {
		return true
	}

	// The original check was for github.com/xhd2015/xgo/orm
	if typeName.Pkg().Path() == "github.com/xhd2015/xgo/orm" && typeName.Name() == "ORM" {
		return true
	}

	return false
}

// extractTableName tries to get the table name from an AST expression
func extractTableName(expr ast.Expr, typeInfo *types.Info, pkg *packages.Package) string {
	// If it's a direct identifier, try to find its value
	if ident, ok := expr.(*ast.Ident); ok {
		// Look for variable declarations in the same package that match this identifier
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
						if name.Name == ident.Name && i < len(valueSpec.Values) {
							// Found the variable declaration
							// If it's created with table.New(), extract the table name
							callExpr, ok := valueSpec.Values[i].(*ast.CallExpr)
							if !ok {
								continue
							}

							fun, ok := callExpr.Fun.(*ast.SelectorExpr)
							if !ok {
								continue
							}

							// Check if it's table.New()
							tableIdent, ok := fun.X.(*ast.Ident)
							if !ok || tableIdent.Name != "table" || fun.Sel.Name != "New" {
								continue
							}

							// Extract table name from string literal
							if len(callExpr.Args) > 0 {
								if lit, ok := callExpr.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
									return strings.Trim(lit.Value, "\"")
								}
							}
						}
					}
				}
			}
		}

		// If we couldn't find a definition, just return the name
		return ident.Name
	}

	// If it's a selector expression, like pkg.Table
	if sel, ok := expr.(*ast.SelectorExpr); ok {
		// Try to find the actual table name if this is a variable
		return sel.Sel.Name
	}

	// For more complex expressions, we'd need to evaluate them
	return ""
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

		// Extract tags
		tags := make(map[string]string)
		if structType.Tag(i) != "" {
			// Parse struct tags like `json:"name" db:"column_name"`
			// This is a simplified version; in a real implementation, you'd use a proper tag parser
			// TODO: parse tags properly
		}

		fieldInfo := FieldInfo{
			Name:    field.Name(),
			Type:    fieldType.String(),
			Tags:    tags,
			Pointer: isPointer,
			Node:    nil,
		}
		info.Fields = append(info.Fields, fieldInfo)
	}

	return info
}

// extractFieldRelations finds field definitions in the package
func extractFieldRelations(pkg *packages.Package, tableName string) []FieldRelation {
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
					if !ok || tableIdent.Name != "Table" {
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
