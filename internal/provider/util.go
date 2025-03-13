package provider

import (
	"encoding/json"
	"fmt"
	"net/url"
  "strings"
)

func IIf[T any](condition bool, trueValue T, falseValue T) T {
	if condition {
		return trueValue
	}
	return falseValue
}

func convertArray[A, B any](source []A, accumulator func(A) B) []B {
	if source == nil {
		return nil
	}

	var result = []B{}
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

// Removes everything after the domain of a URL.
func StripURLToDomain(rawURL string) (string, error) {
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
