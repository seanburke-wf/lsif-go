package indexer

import (
	"bytes"
	"fmt"
	"go/types"
	"strings"
)

// indent is used to format struct fields.
const indent = "    "

// typeString returns the string representation fo the given object's type.
func typeString(obj types.Object) (signature string, extra string) {
	switch v := obj.(type) {
	case *types.PkgName:
		return fmt.Sprintf("package %s", v.Name()), ""

	case *types.TypeName:
		return formatTypeSignature(v), formatTypeExtra(v)

	case *types.Var:
		if v.IsField() {
			// TODO(efritz) - make this be "(T).F" instead of "struct field F string"
			return fmt.Sprintf("struct %s", obj.String()), ""
		}
	}

	return types.ObjectString(obj, packageQualifier), ""
}

// packageQualifier returns an empty string in order to remove the leading package
// name from all identifiers in the return value of types.ObjectString.
func packageQualifier(*types.Package) string { return "" }

// formatTypeSignature returns a brief description of the given struct or interface type.
func formatTypeSignature(obj *types.TypeName) string {
	switch obj.Type().Underlying().(type) {
	case *types.Struct:
		return fmt.Sprintf("type %s struct", obj.Name())
	case *types.Interface:
		return fmt.Sprintf("type %s interface", obj.Name())
	}

	return ""
}

// formatTypeExtra returns the beautified fields of the given struct or interface type.
//
// The output of `types.TypeString` puts fields of structs and interfaces on a single
// line separated by a semicolon. This method simply expands the fields to reside on
// different lines with the appropriate indentation.
func formatTypeExtra(obj *types.TypeName) string {
	extra := types.TypeString(obj.Type().Underlying(), packageQualifier)

	depth := 0
	buf := bytes.NewBuffer(make([]byte, 0, len(extra)))

	for i := 0; i < len(extra); i++ {
		switch extra[i] {
		case ';':
			buf.WriteString("\n")
			buf.WriteString(strings.Repeat(indent, depth))
			i++ // Skip following ' '

		case '{':
			// Special case empty fields so we don't insert
			// an unnecessary newline.
			if i < len(extra)-1 && extra[i+1] == '}' {
				buf.WriteString("{}")
				i++ // Skip following '}'
			} else {
				depth++
				buf.WriteString(" {\n")
				buf.WriteString(strings.Repeat(indent, depth))
			}

		case '}':
			depth--
			buf.WriteString("\n")
			buf.WriteString(strings.Repeat(indent, depth))
			buf.WriteString("}")

		default:
			buf.WriteByte(extra[i])
		}
	}

	return buf.String()
}