package cachekey

import (
	"strconv"
	"strings"
)

func AppendString(parts []string, name, value, def string) []string {
	if value == def {
		return parts
	}
	return append(parts, name+"="+value)
}

func AppendNonEmptyString(parts []string, name, value string) []string {
	if value == "" {
		return parts
	}
	return append(parts, name+"="+value)
}

func AppendInt(parts []string, name string, value, def int) []string {
	if value == def {
		return parts
	}
	return append(parts, name+"="+strconv.Itoa(value))
}

func AppendBool(parts []string, name string, value bool) []string {
	return append(parts, name+"="+strconv.FormatBool(value))
}

func Join(parts []string) string {
	return strings.Join(parts, ";")
}
