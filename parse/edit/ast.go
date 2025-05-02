package edit

import (
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strings"
)

// ParseStruct extracts a StructDef from an AST struct type
func ParseStruct(structType *ast.StructType, name string) StructDef {
	var fields []FieldDef
	// Extract each field
	if structType != nil {
		fields = make([]FieldDef, 0, len(structType.Fields.List))
		for _, field := range structType.Fields.List {
			// Skip embedded fields or fields without names
			if len(field.Names) == 0 {
				continue
			}
			for _, name := range field.Names {
				fieldName := name.Name
				fieldType := extractFieldType(field.Type)

				var tag string
				if field.Tag != nil && len(field.Tag.Value) >= 2 && strings.HasPrefix(field.Tag.Value, "`") && strings.HasSuffix(field.Tag.Value, "`") {
					// strip the backticks
					tag = field.Tag.Value[1 : len(field.Tag.Value)-1]
				}

				// Extract comment if present
				var comment string
				if field.Comment != nil && len(field.Comment.List) > 0 {
					// Get the comment text
					commentText := field.Comment.List[0].Text
					// Remove comment markers (// or /*...*/)
					commentText = strings.TrimPrefix(commentText, "//")
					commentText = strings.TrimPrefix(commentText, "/*")
					commentText = strings.TrimSuffix(commentText, "*/")
					comment = strings.TrimSpace(commentText)
				}

				fields = append(fields, FieldDef{
					Name:    fieldName,
					Type:    fieldType,
					Tag:     tag,
					Comment: comment,
				})
			}
		}
	}

	return StructDef{
		Name:   name,
		Fields: fields,
	}
}

// findLineComment attempts to find a line comment at the given position
func findLineComment(fset *token.FileSet, pos token.Pos) string {
	// Since we don't have direct access to line comments through the AST node,
	// we'd need a more complex approach to extract them.
	// In a real implementation, you might use the comment map from the parser
	// or scan the file for comments at the right position.
	// This is a simplified approach that would need to be expanded.
	return ""
}

// UpdateAST updates the AST struct type based on a StructDef
func UpdateAST(structType *ast.StructType, structDef StructDef) {
	// Create map of current fields to preserve comments and tags
	currentFields := make(map[string]*ast.Field)
	for _, field := range structType.Fields.List {
		if len(field.Names) > 0 {
			currentFields[field.Names[0].Name] = field
		}
	}

	// Create new field list
	newFields := make([]*ast.Field, 0, len(structDef.Fields))

	// Create each field
	for _, fieldDef := range structDef.Fields {
		var field *ast.Field

		// If field exists, preserve its structure but update the type
		if existing, exists := currentFields[fieldDef.Name]; exists {
			field = existing
			// Update the type
			field.Type = parseTypeString(fieldDef.Type)

			// Update the tag if provided
			if fieldDef.Tag != "" {
				field.Tag = &ast.BasicLit{
					Kind:  token.STRING,
					Value: "`" + fieldDef.Tag + "`",
				}
			}
		} else {
			// Create new field
			field = &ast.Field{
				Names: []*ast.Ident{ast.NewIdent(fieldDef.Name)},
				Type:  parseTypeString(fieldDef.Type),
			}

			// Add tag if provided
			if fieldDef.Tag != "" {
				field.Tag = &ast.BasicLit{
					Kind:  token.STRING,
					Value: "`" + fieldDef.Tag + "`",
				}
			}

			// Add comment if provided
			if fieldDef.Comment != "" {
				field.Comment = &ast.CommentGroup{
					List: []*ast.Comment{
						{
							Text: "// " + fieldDef.Comment,
						},
					},
				}
			}
		}

		newFields = append(newFields, field)
	}

	// Replace the fields list
	structType.Fields.List = newFields
}

// extractFieldType returns the string representation of a field type
func extractFieldType(expr ast.Expr) string {
	// Create a new token file set
	fset := token.NewFileSet()

	// Format the expression
	var buf strings.Builder
	err := format.Node(&buf, fset, expr)
	if err != nil {
		// Return a placeholder if formatting fails
		return "unknown"
	}

	return buf.String()
}

// parseTypeString converts a type string to an ast.Expr
func parseTypeString(typeStr string) ast.Expr {
	// Create a temporary source with the type
	src := "package temp\nvar x " + typeStr

	// Parse the source
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		// Return a placeholder identifier if parsing fails
		return ast.NewIdent("error")
	}

	// Extract the type expression from the variable declaration
	if len(file.Decls) > 0 {
		if genDecl, ok := file.Decls[0].(*ast.GenDecl); ok {
			if len(genDecl.Specs) > 0 {
				if valueSpec, ok := genDecl.Specs[0].(*ast.ValueSpec); ok {
					return valueSpec.Type
				}
			}
		}
	}

	// Return a placeholder if extraction fails
	return ast.NewIdent("unknown")
}
