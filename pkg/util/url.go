package util

import (
	"fmt"
	"net/url"
	"strings"
)

// ParseDocsURL parses a google docs url and tries to suss out a docid
func ParseDocsURL(s string) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", fmt.Errorf("Unable to parse url: %v", err)
	}

	parts := strings.Split(u.Path, "/")
	if len(parts) < 4 {
		return "", fmt.Errorf("Invalid URL, cannot parse docID properly")
	}

	return parts[3], nil
}
