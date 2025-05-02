# parse/edit

This package provides utilities for manipulating Go struct definitions in AST (Abstract Syntax Tree) form.

## Main Functions

### UpdateStructFields

```go
func UpdateStructFields(
    structType *ast.StructType,
    name string,
    fields []FieldDef,
    reserveFields map[string]bool,
) StructDef
```

Parses an AST struct and merges it with the desired fields, returning the merged struct definition.

- Extracts current fields from the AST struct
- Creates a struct definition with desired fields
- Preserves specified fields even if not in the desired list
- Returns the merged struct definition without modifying the AST

### MergeStructs

```go
func MergeStructs(current StructDef, desired StructDef, reserveFields map[string]bool) StructDef
```

Merges two struct definitions, with fields from `desired` taking precedence over `current` unless they appear in `reserveFields`.

### ParseStruct

```go
func ParseStruct(structType *ast.StructType, name string) StructDef
```

Converts an AST struct type to a simplified `StructDef` representation.

### UpdateAST

```go
func UpdateAST(structType *ast.StructType, structDef StructDef)
```

Updates an AST struct type based on a `StructDef`. This can be used to apply the merged struct back to the AST.

## Types

### StructDef

```go
type StructDef struct {
    Name   string
    Fields []FieldDef
}

// String returns a Go code representation of the struct definition
func (s StructDef) String() string
```

A simplified representation of a struct definition with a String method that returns the struct as Go code.

### FieldDef

```go
type FieldDef struct {
    Name    string // Field name
    Type    string // Field type as a string
    Tag     string // Field tag (without backticks)
    Comment string // Comment associated with the field
}
```

A simplified representation of a struct field, including name, type, tag and comment.

## Example Usage

```go
// Get AST struct from source code
fset := token.NewFileSet()
file, _ := parser.ParseFile(fset, "example.go", src, parser.ParseComments)

// Find struct type
var structType *ast.StructType
var structName string
ast.Inspect(file, func(n ast.Node) bool {
    if typeSpec, ok := n.(*ast.TypeSpec); ok {
        if st, ok := typeSpec.Type.(*ast.StructType); ok {
            structType = st
            structName = typeSpec.Name.Name
            return false
        }
    }
    return true
})

// Define desired fields with comments and tags
fields := []FieldDef{
    {Name: "ID", Type: "int64", Tag: `json:"id"`, Comment: "Primary key"},
    {Name: "Name", Type: "string", Tag: `json:"name"`, Comment: "User's name"},
    {Name: "Email", Type: "string", Tag: `json:"email,omitempty"`, Comment: "User's email address"},
}

// Get merged struct definition
mergedStruct := UpdateStructFields(structType, structName, fields, nil)

// Get the struct as Go code
structCode := mergedStruct.String()
fmt.Println(structCode)

// If you want to update the AST:
UpdateAST(structType, mergedStruct)

// Print updated AST
var buf bytes.Buffer
printer.Fprint(&buf, fset, file)
fmt.Println(buf.String())
``` 