package util

import (
	"net/url"
	"path"
	"strings"
)

func NormalizeURL(rawURL string) (*url.URL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	// 1. Lowercase Scheme and Host
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)

	// 2. Normalize Path (optional: strip trailing slash or clean relative elements)
	u.Path = path.Clean(u.Path)
	if u.Path == "." {
		u.Path = ""
	}

	// 3. Sort query parameters deterministically
	q := u.Query()
	u.RawQuery = q.Encode() // Encode() sorts keys alphabetically

	return u, nil
}

// CompareURLs checks whether two URL strings are functionally equivalent.
func CompareURLs(urlStr1, urlStr2 string) (bool, error) {
	u1, err := NormalizeURL(urlStr1)
	if err != nil {
		return false, err
	}

	u2, err := NormalizeURL(urlStr2)
	if err != nil {
		return false, err
	}

	// Compare normalized string representations or struct fields
	return u1.String() == u2.String(), nil
}
