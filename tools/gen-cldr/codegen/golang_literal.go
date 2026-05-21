package codegen

import (
	"fmt"
	"io"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

func EmitLiteral(w io.Writer, value any) error {
	_, err := io.WriteString(w, literal(reflect.ValueOf(value)))
	return err
}

func literal(v reflect.Value) string {
	if !v.IsValid() {
		return "nil"
	}
	if v.Type() == reflect.TypeFor[StringRef]() {
		return v.Interface().(StringRef).GoLiteral()
	}
	switch v.Kind() {
	case reflect.String:
		return strconv.Quote(v.String())
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'g', -1, v.Type().Bits())
	case reflect.Slice, reflect.Array:
		return sliceLiteral(v)
	case reflect.Map:
		return mapLiteral(v)
	case reflect.Struct:
		return structLiteral(v)
	case reflect.Pointer:
		if v.IsNil() {
			return "nil"
		}
		return "&" + literal(v.Elem())
	default:
		return fmt.Sprintf("%#v", v.Interface())
	}
}

func sliceLiteral(v reflect.Value) string {
	var b strings.Builder
	b.WriteString(typeName(v.Type()))
	b.WriteString("{")
	for i := range v.Len() {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(literal(v.Index(i)))
	}
	b.WriteString("}")
	return b.String()
}

func mapLiteral(v reflect.Value) string {
	keys := v.MapKeys()
	slices.SortFunc(keys, func(a, b reflect.Value) int {
		return strings.Compare(literal(a), literal(b))
	})
	var b strings.Builder
	b.WriteString(typeName(v.Type()))
	b.WriteString("{\n")
	for _, key := range keys {
		b.WriteString("\t")
		b.WriteString(literal(key))
		b.WriteString(": ")
		b.WriteString(literal(v.MapIndex(key)))
		b.WriteString(",\n")
	}
	b.WriteString("}")
	return b.String()
}

func structLiteral(v reflect.Value) string {
	var b strings.Builder
	b.WriteString("{")
	wrote := false
	for i := range v.NumField() {
		field := v.Type().Field(i)
		if !field.IsExported() {
			continue
		}
		if wrote {
			b.WriteString(", ")
		}
		b.WriteString(field.Name)
		b.WriteString(": ")
		b.WriteString(literal(v.Field(i)))
		wrote = true
	}
	b.WriteString("}")
	return b.String()
}

func typeName(t reflect.Type) string {
	if t.PkgPath() == "" {
		return t.String()
	}
	return t.String()
}
