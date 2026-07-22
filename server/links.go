package main

import (
	"net/url"
	"regexp"
	"strings"
)

var messageURLPattern = regexp.MustCompile(`https?://[^\s<>{}\[\]"]+`)

type entityLink struct {
	URL        string
	Identifier string
	Route      string
	Collection string
}

func findEntityLinks(message string, config *configuration) []entityLink {
	if config == nil || config.frontendBase == nil {
		return nil
	}
	seen := make(map[string]struct{})
	links := make([]entityLink, 0)
	for _, raw := range messageURLPattern.FindAllString(message, -1) {
		candidate := strings.TrimRight(raw, ".,;:!?)]'")
		parsed, err := url.Parse(candidate)
		if err != nil || !sameOrigin(parsed, config.frontendBase) {
			continue
		}

		relativePath, ok := pathBelowBase(parsed.EscapedPath(), config.frontendBase.EscapedPath())
		if !ok {
			continue
		}
		segments := strings.Split(strings.Trim(relativePath, "/"), "/")
		if len(segments) < 2 {
			continue
		}
		route, err := url.PathUnescape(segments[0])
		if err != nil {
			continue
		}
		collection, exists := entityCollection(route)
		if !exists {
			continue
		}
		identifier, err := url.PathUnescape(segments[1])
		if err != nil || strings.TrimSpace(identifier) == "" {
			continue
		}
		if _, duplicate := seen[candidate]; duplicate {
			continue
		}
		seen[candidate] = struct{}{}
		links = append(links, entityLink{
			URL:        candidate,
			Identifier: identifier,
			Route:      strings.ToLower(route),
			Collection: collection,
		})
		if len(links) == config.MaxPreviewsPerPost {
			break
		}
	}
	return links
}

func sameOrigin(left, right *url.URL) bool {
	return strings.EqualFold(left.Scheme, right.Scheme) && strings.EqualFold(left.Host, right.Host)
}

func pathBelowBase(path, basePath string) (string, bool) {
	base := strings.TrimRight(basePath, "/")
	if base == "" {
		return path, true
	}
	if path == base {
		return "", true
	}
	prefix := base + "/"
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}
	return strings.TrimPrefix(path, base), true
}
