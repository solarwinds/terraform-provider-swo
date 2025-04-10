package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"net/url"
	"strings"
)

func convertArray[A, B any](source []A, accumulator func(A) B) []B {
	if source == nil {
		return nil
	}

	var result []B
	for _, x := range source {
		result = append(result, accumulator(x))
	}
	return result
}

func convertObject[T any](from any) (*T, error) {
	b, err := json.Marshal(&from)
	if err != nil {
		return nil, err
	}

	var result T
	err = json.Unmarshal(b, &result)
	if err != nil {
		return nil, err
	}

	return &result, err
}

// stripURLToDomain Removes everything after the domain of a URL.
func stripURLToDomain(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host), nil
}

func findCaseInsensitiveMatch(slice []string, target string) string {
	for _, str := range slice {
		if strings.EqualFold(str, target) {
			return str
		}
	}
	return ""
}

func lowerCaseSlice(input []string) []string {
	for i, v := range input {
		input[i] = strings.ToLower(v)
	}
	return input
}

func sliceToStringList[T any](
	items []T,
	mapFn func(T) string,
) types.List {
	if items == nil || len(items) == 0 {
		return types.ListNull(types.StringType)
	}

	ctx := context.Background()
	elements := make([]attr.Value, len(items))

	for i, item := range items {
		elements[i] = types.StringValue(mapFn(item))
	}

	list, _ := types.ListValueFrom(ctx, types.StringType, elements)
	return list
}

func attrValueToString(val attr.Value) string {
	if val == nil {
		return ""
	}
	return strings.Trim(val.String(), "\"")
}
