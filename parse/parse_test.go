package parse

import (
	"encoding/json"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/goinfo"
	"golang.org/x/tools/go/packages"
)

func TestExtractTableName(t *testing.T) {
	// Setup a temporary directory with test files
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	// Create a test package
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedDeps | packages.NeedName | packages.NeedImports | packages.NeedTypesInfo,
		Dir:  tmpDir,
	}, "./...")
	if err != nil {
		t.Fatalf("Failed to load packages: %v", err)
	}

	if len(pkgs) == 0 {
		t.Fatal("No packages loaded")
	}

	pkg := pkgs[0]

	tableVar, _ := findVarDef(pkg, "Table")
	if tableVar == nil {
		t.Fatal("Table variable not found in test files")
	}
	tableVarDefObj, ok := pkg.TypesInfo.Defs[tableVar]
	if !ok {
		t.Fatal("Table variable not found in test files")
	}
	tableVarDef, ok := tableVarDefObj.(*types.Var)
	if !ok {
		t.Fatal("Table variable is not a types.Var")
	}

	// Test extractTableName
	tableName := extractTableFromVar(pkg, tableVarDef)
	expected := "test_users"
	if tableName != expected {
		t.Errorf("Expected table name %q, got %q", expected, tableName)
	}
}

func TestExtractModelInfo(t *testing.T) {
	// Setup a temporary directory with test files
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	// Create a test package
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedDeps | packages.NeedName | packages.NeedImports | packages.NeedTypesInfo,
		Dir:  tmpDir,
	}, "./...")
	if err != nil {
		t.Fatalf("Failed to load packages: %v", err)
	}

	if len(pkgs) == 0 {
		t.Fatal("No packages loaded")
	}

	pkg := pkgs[0]
	typeInfo := pkg.TypesInfo

	// Find the User type in the AST
	var userExpr ast.Expr
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if typeSpec.Name.Name == "User" {
					userExpr = typeSpec.Name
					break
				}
			}
		}
	}

	if userExpr == nil {
		t.Fatal("User type not found in test files")
	}

	// Test extractModelInfo
	modelInfo := extractModelInfo(userExpr, typeInfo, pkg)
	if modelInfo.Name != "User" {
		t.Errorf("Expected model name 'User', got %q", modelInfo.Name)
	}
	if len(modelInfo.Fields) != 5 {
		t.Errorf("Expected 5 fields, got %d", len(modelInfo.Fields))
	}

	// Check for specific fields
	fieldNames := map[string]bool{
		"Id":         false,
		"Name":       false,
		"Email":      false,
		"CreateTime": false,
		"UpdateTime": false,
	}

	for _, field := range modelInfo.Fields {
		fieldNames[field.Name] = true
	}

	for name, found := range fieldNames {
		if !found {
			t.Errorf("Expected field %q not found", name)
		}
	}
}

func TestExtractFieldRelations(t *testing.T) {
	// Setup a temporary directory with test files
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	// Create a test package
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedDeps | packages.NeedName | packages.NeedImports | packages.NeedTypesInfo,
		Dir:  tmpDir,
	}, "./...")
	if err != nil {
		t.Fatalf("Failed to load packages: %v", err)
	}

	if len(pkgs) == 0 {
		t.Fatal("No packages loaded")
	}

	pkg := pkgs[0]

	tableVar, _ := findVarDef(pkg, "Table")
	if tableVar == nil {
		t.Fatal("Table variable not found in test files")
	}
	tableVarDefObj, ok := pkg.TypesInfo.Defs[tableVar]
	if !ok {
		t.Fatal("Table variable not found in test files")
	}
	tableVarDef, ok := tableVarDefObj.(*types.Var)
	if !ok {
		t.Fatal("Table variable is not a types.Var")
	}

	// Test extractFieldRelations
	fieldRelations := extractFieldRelations(pkg, tableVarDef)
	if len(fieldRelations) != 5 {
		t.Errorf("Expected 5 field relations, got %d", len(fieldRelations))
	}

	// Check for specific field relations
	expectedFields := map[string]string{
		"ID":         "id",
		"Name":       "name",
		"Email":      "email",
		"CreateTime": "create_time",
		"UpdateTime": "update_time",
	}

	for _, rel := range fieldRelations {
		expectedCol, exists := expectedFields[rel.FieldName]
		if !exists {
			t.Errorf("Unexpected field relation: %s -> %s", rel.FieldName, rel.ColumnName)
			continue
		}
		if rel.ColumnName != expectedCol {
			t.Errorf("Expected column name %q for field %q, got %q", expectedCol, rel.FieldName, rel.ColumnName)
		}
		delete(expectedFields, rel.FieldName)
	}

	for field := range expectedFields {
		t.Errorf("Expected field relation for %q not found", field)
	}
}

func TestIsOrmBind(t *testing.T) {
	// This test would require a more complex setup to create a types.Type
	// that represents an ORM binding. Skipping for now.
	t.Skip("Needs complex setup to test isOrmBind")
}

// Helper function to set up test directory with test files
func setupTestDir(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "ormx-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	absWd, err := filepath.Abs(wd)
	if err != nil {
		t.Fatalf("Failed to get absolute working directory: %v", err)
	}
	subPaths, _, err := goinfo.ResolveMainModule(absWd)
	if err != nil {
		t.Fatalf("Failed to resolve main module: %v", err)
	}
	projectRoot := absWd
	for i, n := 0, len(subPaths); i < n; i++ {
		projectRoot = filepath.Dir(projectRoot)
	}

	// Create a go.mod file
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(`module testormx

go 1.19

require github.com/xhd2015/ormx v0.0.0

replace github.com/xhd2015/ormx => `+projectRoot+`
`), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create a test file
	testFile := `package testorm

import (
	"time"

	"github.com/xhd2015/ormx/orm"
	"github.com/xhd2015/ormx/table"
)

// Table is the test_users table
var Table = table.New("test_users")

// Field definitions
var (
	ID         = Table.Int64("id")
	Name       = Table.String("name")
	Email      = Table.String("email")
	CreateTime = Table.Time("create_time")
	UpdateTime = Table.Time("update_time")
)

var ORM = orm.Bind[User, UserOptional](nil, Table)

type User struct {
	Id         int64
	Name       string
	Email      string
	CreateTime time.Time
	UpdateTime time.Time
}

type UserOptional struct {
	Id         *int64
	Name       *string
	Email      *string
	CreateTime *time.Time
	UpdateTime *time.Time
}
`

	err = os.WriteFile(filepath.Join(tmpDir, "testorm.go"), []byte(testFile), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	err = cmd.Dir(tmpDir).Run("go", "mod", "tidy")
	if err != nil {
		t.Fatalf("Failed to tidy go.mod: %v", err)
	}

	return tmpDir
}

// TestLoadAndExtractRelations tests the LoadAndExtractRelations function
func TestLoadAndExtractRelations(t *testing.T) {
	// Setup a temporary directory with test files
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	// Call LoadAndExtractRelations
	fset := token.NewFileSet()
	pkgResults, err := LoadAndExtractRelations(fset, tmpDir, []string{"./..."})
	if err != nil {
		t.Fatalf("LoadAndExtractRelations failed: %v", err)
	}

	// Debug info
	t.Logf("Packages: %d", len(pkgResults))

	// Check each package
	for _, pkg := range pkgResults {
		t.Logf("Processing package: %s", pkg.PkgPath)

		// Check each file
		for _, file := range pkg.Files {
			t.Logf("Processing file: %s", file.AbsFile)
			if !strings.HasSuffix(file.AbsFile, "testorm.go") {
				t.Errorf("Unexpected file path: %s", file.AbsFile)
			}

			// Now we can directly marshal the tables with AST nodes excluded
			tablesJSON, err := json.Marshal(file.Tables)
			if err != nil {
				t.Fatalf("Failed to marshal tables: %v", err)
			}
			t.Logf("Tables: %s", string(tablesJSON))

			// Check the table relations
			if len(file.Tables) == 0 {
				t.Error("No table relations found in file")
				continue
			}

			// Verify the table relation
			rel := file.Tables[0]
			if rel.TableName != "test_users" {
				t.Errorf("Expected table name 'test_users', got %q", rel.TableName)
			}

			if rel.Model.Name != "User" {
				t.Errorf("Expected model name 'User', got %q", rel.Model.Name)
			}

			if len(rel.Model.Fields) != 5 {
				t.Errorf("Expected 5 model fields, got %d", len(rel.Model.Fields))
			}

			if rel.OptionalModel.Name != "UserOptional" {
				t.Errorf("Expected optional model name 'UserOptional', got %q", rel.OptionalModel.Name)
			}

			if len(rel.Fields) == 0 {
				t.Error("No field relations found")
			}

			// Verify the AST nodes are set
			if rel.Model.Node == nil {
				t.Error("Model Node is nil")
			}

			for i, field := range rel.Model.Fields {
				if field.Node == nil {
					t.Errorf("Model field[%d] Node is nil", i)
				}
			}
		}
	}
}
