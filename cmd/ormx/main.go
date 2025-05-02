package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/less-gen/strcase"
	"github.com/xhd2015/ormx/parse"
	parse_edit "github.com/xhd2015/ormx/parse/edit"
	"github.com/xhd2015/xgo/support/edit/goedit"
	"github.com/xhd2015/xgo/support/goinfo"
)

const help = `
Usage: ormx <command>

Commands:
  gen     generate models

`

func main() {
	err := handle(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires command, run `ormx help`")
	}

	switch args[0] {
	case "help":
		fmt.Println(strings.TrimPrefix(help, "\n"))
		return nil
	case "gen":
		return gen(args[1:])
	}

	return fmt.Errorf("unknown command, run `ormx help`")
}

func gen(args []string) error {
	var dir string
	var remainArgs []string
	n := len(args)
	for i := 0; i < n; i++ {
		arg := args[i]
		if arg == "--dir" {
			if i+1 >= n {
				return fmt.Errorf("%s requires argument", arg)
			}
			dir = args[i+1]
			remainArgs = args[i+2:]
			continue
		} else if strings.HasPrefix(arg, "--dir=") {
			dir = arg[len("--dir="):]
			continue
		}
		if strings.HasPrefix(arg, "-") {
			return fmt.Errorf("unrecognized flag: %s", arg)
		}
		remainArgs = append(remainArgs, arg)
	}

	// if len(args) == 0 {
	// 	return fmt.Errorf("requires table name, run `ormx gen <table_name>`")
	// }
	// TODO:
	var loadDir string
	var loadArgs []string
	if len(remainArgs) == 0 {
		resolveDir := dir
		if dir == "" {
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			resolveDir = wd
		}

		absWd, err := filepath.Abs(resolveDir)
		if err != nil {
			return err
		}

		subPaths, mainModule, err := goinfo.ResolveMainModule(absWd)
		if err != nil {
			return err
		}

		_ = mainModule

		mainDir := absWd
		for i, n := 0, len(subPaths); i < n; i++ {
			mainDir = filepath.Dir(mainDir)
		}

		loadDir = mainDir
		loadArgs = []string{"./..."}
	} else {
		loadDir = dir
		loadArgs = remainArgs
	}

	// Load the packages and extract table relations
	fset := token.NewFileSet()
	pkgs, err := parse.LoadAndExtractRelations(fset, loadDir, loadArgs)
	if err != nil {
		return err
	}

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			content, err := os.ReadFile(file.AbsFile)
			if err != nil {
				return err
			}
			edit := goedit.NewWithBytes(fset, content)
			for _, table := range file.Tables {
				amendModels(edit, file, table)
			}
			if !edit.HasEdit() {
				continue
			}
			newCode := edit.Buffer().Bytes()
			err = os.WriteFile(file.AbsFile, newCode, 0644)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func amendModels(edit *goedit.Edit, file *parse.File, table *parse.TableRelation) {
	updateStructFields(edit, file, table, table.Model, table.Fields, table.Model.Fields, false)
	updateStructFields(edit, file, table, table.OptionalModel, table.Fields, table.OptionalModel.Fields, true)
}

// updateStructFields checks and updates struct fields to match the table field definitions
func updateStructFields(edit *goedit.Edit, file *parse.File, table *parse.TableRelation, model parse.ModelInfo, tableFields []parse.FieldRelation, structFields []parse.FieldInfo, asPointer bool) {
	var structTypeName string
	var structType *ast.StructType
	if typeSpec, ok := model.Node.(*ast.TypeSpec); ok {
		if st, ok := typeSpec.Type.(*ast.StructType); ok {
			structTypeName = typeSpec.Name.Name
			structType = st
		}
	}

	current := parse_edit.ParseStruct(structType, structTypeName)

	// Create desired fields from table fields
	var desiredFields []parse_edit.FieldDef
	for _, tableField := range tableFields {
		structType := getStructType(tableField.Type)
		if asPointer {
			structType = "*" + structType
		}
		desiredFields = append(desiredFields, parse_edit.FieldDef{
			Name: strcase.SnakeToCamel(tableField.ColumnName),
			Type: structType,
		})
	}

	desired := parse_edit.StructDef{
		Name:   model.Name,
		Fields: desiredFields,
	}

	// No fields should be reserved - we want to keep exactly what's in the table definition
	reserveFields := map[string]bool{
		"Count": true,
	}

	// Merge the structs
	result := parse_edit.MergeStructs(current, desired, reserveFields)

	if model.Node != nil {
		edit.Replace(model.Node.Pos(), model.Node.End(), result.Format(parse_edit.FormatOptions{
			NoPrefixType: true,
		}))
	} else {
		edit.Insert(file.AST.End(), "\n"+result.Format(parse_edit.FormatOptions{
			NoPrefixType: false,
		}))
	}
}

func getStructType(name string) string {
	switch name {
	case "Int64":
		return "int64"
	case "Time":
		return "time.Time"
	case "String":
		return "string"
	case "Bool":
		return "bool"
	case "Float64":
		return "float64"
	}
	return "any"
}
