package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/xgo/support/assert"
)

const base = `
package testorm

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
`

const FullDefiniton = `
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

// Helper function to set up test directory with test files
func setupTestDir(t *testing.T, inputCode string) (dir string, file string) {
	tmpDir, err := os.MkdirTemp("", "ormx-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create a go.mod file
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(`module testormx

go 1.19

require github.com/xhd2015/ormx v0.0.0
`), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create a test file
	testFile := base + inputCode

	file = filepath.Join(tmpDir, "testorm.go")
	err = os.WriteFile(file, []byte(testFile), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	return tmpDir, file
}

func runGen(t *testing.T, inputCode string) (string, error) {
	t.Helper()
	tmpDir, file := setupTestDir(t, inputCode)
	defer os.RemoveAll(tmpDir)

	err := gen([]string{"--dir=" + tmpDir})
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// TestGen_NoChange tests updating existing model fields
func TestGen_NoChange(t *testing.T) {
	code, err := runGen(t, FullDefiniton)
	if err != nil {
		t.Fatalf("Failed to run gen: %v", err)
	}

	expectCode := base + `
type User struct {
	Id int64
	Name string
	Email string
	CreateTime time.Time
	UpdateTime time.Time
}

type UserOptional struct {
	Id *int64
	Name *string
	Email *string
	CreateTime *time.Time
	UpdateTime *time.Time
}
`
	if diff := assert.Diff(expectCode, code); diff != "" {
		t.Error(diff)
	}
}

// TestGen_CreateModel tests creating new models when they don't exist
func TestGen_CreateModel(t *testing.T) {
	// Run with no model definitions
	code, err := runGen(t, "")
	if err != nil {
		t.Fatalf("Failed to run gen: %v", err)
	}

	// Expect the base code plus newly created User and UserOptional models
	expectCode := base + `type User struct {
	Id int64
	Name string
	Email string
	CreateTime time.Time
	UpdateTime time.Time
}
type UserOptional struct {
	Id *int64
	Name *string
	Email *string
	CreateTime *time.Time
	UpdateTime *time.Time
}
`

	// Remove trailing newlines for comparison
	if diff := assert.Diff(expectCode, code); diff != "" {
		t.Error(diff)
	}
}

// TestGen_AddMissingField tests that missing fields in User struct are added
func TestGen_AddMissingField(t *testing.T) {
	// Define User with missing Email field
	incompleteDefinition := `
type User struct {
	Id         int64
	Name       string
	// Email field is missing
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
	code, err := runGen(t, incompleteDefinition)
	if err != nil {
		t.Fatalf("Failed to run gen: %v", err)
	}

	want := base + `
type User struct {
	Id int64
	Name string
	Email string
	CreateTime time.Time
	UpdateTime time.Time
}

type UserOptional struct {
	Id *int64
	Name *string
	Email *string
	CreateTime *time.Time
	UpdateTime *time.Time
}
`

	if diff := assert.Diff(want, code); diff != "" {
		t.Error(diff)
	}
}

// TestGen_RemoveExtraField tests that extra fields not defined in the Table are removed
func TestGen_RemoveExtraField(t *testing.T) {
	// Define User with an extra Age field that is not in the Table definition
	codeWithExtraField := `
type User struct {
	Id         int64
	Name       string
	Email      string
	Age        int    // Extra field not in Table definition
	CreateTime time.Time
	UpdateTime time.Time
}

type UserOptional struct {
	Id         *int64
	Name       *string
	Email      *string
	Age        *int    // Extra field not in Table definition
	CreateTime *time.Time
	UpdateTime *time.Time
}
`
	code, err := runGen(t, codeWithExtraField)
	if err != nil {
		t.Fatalf("Failed to run gen: %v", err)
	}

	// The extra Age field should be removed in the generated code
	// But comments are preserved
	want := base + `
type User struct {
	Id int64
	Name string
	Email string
	CreateTime time.Time
	UpdateTime time.Time
}

type UserOptional struct {
	Id *int64
	Name *string
	Email *string
	CreateTime *time.Time
	UpdateTime *time.Time
}
`

	if diff := assert.Diff(want, code); diff != "" {
		t.Error(diff)
	}
}
