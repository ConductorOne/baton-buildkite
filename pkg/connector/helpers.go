package connector

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"

	"github.com/google/go-querystring/query"
)

// pageTokenToInt converts a string page token to an integer page number.
// If the token is empty, it returns 0.
func pageTokenToInt(token string) (int, error) {
	if token == "" {
		return 0, nil
	}

	page, err := strconv.Atoi(token)
	if err != nil {
		return 0, fmt.Errorf("baton-buildkite: invalid page token: %w", err)
	}

	return page, nil
}

// addOptions adds the parameters in opt as URL query parameters to s.
// opt must be a struct whose fields may contain "url" tags.
func addOptions(s string, opt interface{}) (string, error) {
	v := reflect.ValueOf(opt)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	qs, err := query.Values(opt)
	if err != nil {
		return s, err
	}

	u.RawQuery = qs.Encode()
	return u.String(), nil
}
