package shortcode

import (
	"errors"
	"net/url"
	"sort"
)

const (
	ModeFull     = "full"
	ModePathOnly = "path_only"
)

func Normalize(raw, mode string, blacklist []string) (string, error) {
	if mode != ModeFull && mode != ModePathOnly {
		return "", errors.New("mode must be full or path_only")
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return "", errors.New("url must be an absolute http or https URL")
	}
	if mode == ModePathOnly {
		u.RawQuery = ""
		return u.String(), nil
	}
	blocked := make(map[string]struct{}, len(blacklist))
	for _, key := range blacklist {
		blocked[key] = struct{}{}
	}
	values := u.Query()
	for key := range values {
		if _, found := blocked[key]; found {
			delete(values, key)
			continue
		}
		sort.Strings(values[key])
	}
	u.RawQuery = values.Encode()
	return u.String(), nil
}
