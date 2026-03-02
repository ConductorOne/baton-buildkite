package connector

import (
	"fmt"
	"strconv"
)

// pageTokenToInt converts a string page token to an integer page number.
// If the token is empty, it returns 0.
// Returns an error if the token cannot be parsed as an integer.
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
