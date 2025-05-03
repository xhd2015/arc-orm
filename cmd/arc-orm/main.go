package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/arc-orm/cmd/arc-orm/parse"
	"github.com/xhd2015/less-gen/go/gofmt"
	"github.com/xhd2015/less-gen/go/gostruct"
	"github.com/xhd2015/less-gen/strcase"
	"github.com/xhd2015/xgo/support/edit/goedit"
	"github.com/xhd2015/xgo/support/goinfo"
)

const help = `
Usage: ormx <command>

Commands:
  gen     generate models
  sync    sync models, same as gen

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
		return fmt.Errorf("requires command, run `arc-orm help`")
	}

	switch args[0] {
	case "help":
		fmt.Println(strings.TrimPrefix(help, "\n"))
		return nil
	case "gen", "sync":
		return gen(args[1:])
	}

	return fmt.Errorf("unknown command, run `arc-orm help`")
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
	pkgs, err := parse.ScanRelations(fset, loadDir, loadArgs)
	if err != nil {
		return err
	}

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			code, err := os.ReadFile(file.AbsFile)
			if err != nil {
				return err
			}
			edit := goedit.NewWithBytes(fset, code)
			for i, table := range file.Tables {
				if table.NeedCreateORM {
					// var ORM = orm.Bind[table.Model, table.OptionalModel](nil, table.TableName)
					declare := fmt.Sprintf("\nvar ORM = orm.Bind[%s, %s](nil, %s)", table.Model.Name, table.OptionalModel.Name, table.TablVarName)
					pos, newLine := getMinAppendPos(file, table)
					if newLine {
						declare += "\n"
					}
					edit.Insert(pos, declare)
				}
				if !file.HasGenerate && i == 0 {
					declare := "\n//go:generate go run github.com/xhd2015/arc-orm/cmd/arc-orm@latest sync"
					pos, newLine := getMinAppendPos(file, table)
					if newLine {
						declare += "\n"
					}
					edit.Insert(pos, declare)
				}
				amendModels(edit, file, code, table)
			}
			if !edit.HasEdit() {
				continue
			}
			newCode := edit.Buffer().Bytes()
			newCode = []byte(gofmt.TryFormatCode(string(newCode)))
			err = os.WriteFile(file.AbsFile, newCode, 0644)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func getMinAppendPos(file *parse.File, table *parse.TableRelation) (token.Pos, bool) {
	minDeclPos := token.NoPos
	if table.Model.GenDecl != nil {
		minDeclPos = table.Model.GenDecl.Pos()
		// p = minPos(p, table.Model.GenDecl.Pos())
	}
	if table.OptionalModel.GenDecl != nil {
		minDeclPos = minPos(minDeclPos, table.OptionalModel.GenDecl.Pos())
	}
	if minDeclPos.IsValid() {
		return minDeclPos, true
	}
	return file.AST.End(), false
}

func minPos(a token.Pos, b token.Pos) token.Pos {
	if !a.IsValid() {
		return b
	}
	if !b.IsValid() {
		return a
	}
	if a < b {
		return a
	}
	return b
}

func amendModels(edit *goedit.Edit, file *parse.File, code []byte, table *parse.TableRelation) {
	updateStructFields(edit, file, code, table, table.Model, table.Fields, table.Model.Fields, false)
	updateStructFields(edit, file, code, table, table.OptionalModel, table.Fields, table.OptionalModel.Fields, true)
}

// updateStructFields checks and updates struct fields to match the table field definitions
func updateStructFields(edit *goedit.Edit, file *parse.File, code []byte, table *parse.TableRelation, model parse.ModelInfo, tableFields []parse.FieldRelation, structFields []parse.FieldInfo, asPointer bool) {
	var structTypeName string
	var structType *ast.StructType
	if model.TypeSpec != nil && model.TypeSpec.Name != nil {
		structTypeName = model.TypeSpec.Name.Name
	}
	structType = model.StructType

	current := gostruct.ParseStruct(edit.Fset(), structType, structTypeName)

	// Create desired fields from table fields
	var desiredFields []gostruct.FieldDef
	for _, tableField := range tableFields {
		structType := getStructType(tableField.Type)
		if asPointer {
			structType = "*" + structType
		}
		desiredFields = append(desiredFields, gostruct.FieldDef{
			Name: strcase.SnakeToCamel(tableField.ColumnName),
			Type: structType,
		})
	}

	desired := gostruct.StructDef{
		Name:   model.Name,
		Fields: desiredFields,
	}

	// No fields should be reserved - we want to keep exactly what's in the table definition
	reserveFields := map[string]bool{
		"Count": true,
	}

	// Merge the structs
	result := gostruct.MergeStructs(current, desired, reserveFields)

	if model.TypeSpec != nil {
		edit.Replace(model.TypeSpec.Pos(), model.TypeSpec.End(), result.Format(gostruct.FormatOptions{
			NoPrefixType: true,
		}))
	} else {
		edit.Insert(file.AST.End(), "\n"+result.Format(gostruct.FormatOptions{}))
	}
}

func getStructType(name string) string {
	switch name {
	case "Int64":
		return "int64"
	case "Int32":
		return "int32"
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
