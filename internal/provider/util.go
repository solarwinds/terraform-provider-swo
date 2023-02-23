package provider

import (
	"strconv"
	"strings"
)

func Trim(s string) string {
	return strings.Trim(s, "\"")
}

func GetDataType(s string) string {
	dataType := "string"

	if _, err := strconv.Atoi(s); err == nil {
		dataType = "number"
	}

	return dataType
}
