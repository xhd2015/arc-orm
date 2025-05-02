package edit

import (
	"fmt"
	"strings"
)

// StructDef represents a struct definition in a simplified form
type StructDef struct {
	Name   string
	Fields []FieldDef
}

// FieldDef represents a struct field in a simplified form
type FieldDef struct {
	Name    string // Field name
	Type    string // Field type as a string
	Tag     string // Field tag (without backticks)
	Comment string // Comment associated with the field
}

// String returns a Go code representation of the struct definition
func (s StructDef) String() string {
	return s.Format(FormatOptions{})
}

type FormatOptions struct {
	NoPrefixType bool
}

func (s StructDef) Format(opts FormatOptions) string {
	var sb strings.Builder

	// Write struct header
	if !opts.NoPrefixType {
		sb.WriteString("type ")
	}
	sb.WriteString(fmt.Sprintf("%s struct {\n", s.Name))

	// Write each field with proper indentation
	for _, field := range s.Fields {
		// Start with field name and type
		fieldLine := fmt.Sprintf("\t%s %s", field.Name, field.Type)

		// Add tag if present
		if field.Tag != "" {
			fieldLine += fmt.Sprintf(" `%s`", field.Tag)
		}

		// Add comment if present
		if field.Comment != "" {
			fieldLine += fmt.Sprintf(" // %s", field.Comment)
		}

		sb.WriteString(fieldLine + "\n")
	}

	// Close struct definition
	sb.WriteString("}")

	return sb.String()
}

// MergeStructs combines two struct definitions, with fields from 'desired'
// taking precedence over 'current' unless they appear in 'reserveFields'
func MergeStructs(current StructDef, desired StructDef, reserveFields map[string]bool) StructDef {
	// Create a map of current fields for quick lookup
	currentFields := make(map[string]FieldDef)
	for _, field := range current.Fields {
		currentFields[field.Name] = field
	}

	// Create a map of desired fields for quick lookup
	desiredFieldMap := make(map[string]FieldDef)
	for _, field := range desired.Fields {
		desiredFieldMap[field.Name] = field
	}

	// Build result using fields from desired struct, but preserve tags and comments
	result := StructDef{Name: desired.Name}

	// Process all desired fields
	for _, field := range desired.Fields {
		// Check if field exists in current struct
		if currentField, exists := currentFields[field.Name]; exists {
			// Copy field with potentially merged properties
			mergedField := field

			// If current field has Tag and desired doesn't, copy it
			if field.Tag == "" && currentField.Tag != "" {
				mergedField.Tag = currentField.Tag
			}

			// If current field has Comment and desired doesn't, copy it
			if field.Comment == "" && currentField.Comment != "" {
				mergedField.Comment = currentField.Comment
			}

			result.Fields = append(result.Fields, mergedField)
		} else {
			// Field doesn't exist in current struct, add as is
			result.Fields = append(result.Fields, field)
		}
	}

	// Check for fields that should be reserved but aren't in desired
	if reserveFields != nil {
		for name, field := range currentFields {
			// If field should be reserved and isn't already in desired fields
			if reserveFields[name] && desiredFieldMap[name].Name == "" {
				result.Fields = append(result.Fields, field)
			}
		}
	}

	return result
}
