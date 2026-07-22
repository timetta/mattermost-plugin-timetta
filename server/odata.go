package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const maxODataResponseBytes = 1024 * 1024

type odataClient struct {
	httpClient *http.Client

	stateSupportMu   sync.RWMutex
	stateUnsupported map[string]struct{}
}

type odataHTTPError struct {
	statusCode int
	body       string
}

func (e *odataHTTPError) Error() string {
	return fmt.Sprintf("Timetta OData returned HTTP %d: %s", e.statusCode, truncateBytes(strings.TrimSpace(e.body), 300))
}

func newODataClient() *odataClient {
	return &odataClient{
		httpClient: &http.Client{
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		stateUnsupported: make(map[string]struct{}),
	}
}

func (c *odataClient) loadEntity(ctx context.Context, config *configuration, link entityLink) (map[string]any, error) {
	expandState := !c.isStateUnsupported(link.Collection)
	entity, err := c.loadEntityOnce(ctx, config, link, expandState)
	if err == nil || !expandState || !isBadRequest(err) {
		return entity, err
	}

	// Not every OData type exposes State. Retry without the navigation property;
	// remember a successful fallback so later previews use a single request.
	entity, fallbackErr := c.loadEntityOnce(ctx, config, link, false)
	if fallbackErr == nil {
		c.markStateUnsupported(link.Collection)
		return entity, nil
	}
	return nil, err
}

func (c *odataClient) loadEntityOnce(ctx context.Context, config *configuration, link entityLink, expandState bool) (map[string]any, error) {
	requestURL, collectionResponse, err := buildODataURL(config.APIURL, link, expandState)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create OData request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	if token := strings.TrimSpace(config.BearerToken); token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request Timetta OData: %w", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, maxODataResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read Timetta OData response: %w", err)
	}
	if len(body) > maxODataResponseBytes {
		return nil, fmt.Errorf("Timetta OData response exceeds %d bytes", maxODataResponseBytes)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, &odataHTTPError{statusCode: response.StatusCode, body: string(body)}
	}

	if collectionResponse {
		var envelope struct {
			Value []map[string]any `json:"value"`
		}
		if err := json.Unmarshal(body, &envelope); err != nil {
			return nil, fmt.Errorf("decode Timetta OData collection: %w", err)
		}
		if len(envelope.Value) == 0 {
			return nil, fmt.Errorf("entity %q was not found", link.Identifier)
		}
		return envelope.Value[0], nil
	}

	var entity map[string]any
	if err := json.Unmarshal(body, &entity); err != nil {
		return nil, fmt.Errorf("decode Timetta OData entity: %w", err)
	}
	return entity, nil
}

func buildODataURL(apiBase string, link entityLink, expandState bool) (string, bool, error) {
	base, err := url.Parse(strings.TrimRight(apiBase, "/") + "/OData/" + link.Collection)
	if err != nil {
		return "", false, fmt.Errorf("build OData URL: %w", err)
	}
	query := base.Query()
	expansions := entityExpansions(link.Collection, expandState)
	if len(expansions) > 0 {
		query.Set("$expand", strings.Join(expansions, ","))
	}

	if isGUID(link.Identifier) {
		requestURL := base.String() + "(" + link.Identifier + ")"
		if encodedQuery := query.Encode(); encodedQuery != "" {
			requestURL += "?" + encodedQuery
		}
		return requestURL, false, nil
	}
	keyProperty := collectionStringKey(link.Collection)
	if keyProperty == "" {
		return "", false, fmt.Errorf("collection %q only supports Guid identifiers, got %q", link.Collection, link.Identifier)
	}
	query.Set("$filter", fmt.Sprintf("%s eq '%s'", keyProperty, strings.ReplaceAll(link.Identifier, "'", "''")))
	query.Set("$top", "1")
	base.RawQuery = query.Encode()
	return base.String(), true, nil
}

func entityExpansions(collection string, includeState bool) []string {
	expansions := make([]string, 0, 4)
	if includeState {
		expansions = append(expansions, "state($select=name)")
	}
	if collection == "Issues" {
		expansions = append(
			expansions,
			"type($select=name)",
			"priority($select=name)",
			"project($select=name)",
		)
	}
	return expansions
}

func (c *odataClient) isStateUnsupported(collection string) bool {
	c.stateSupportMu.RLock()
	defer c.stateSupportMu.RUnlock()
	_, unsupported := c.stateUnsupported[collection]
	return unsupported
}

func (c *odataClient) markStateUnsupported(collection string) {
	c.stateSupportMu.Lock()
	defer c.stateSupportMu.Unlock()
	c.stateUnsupported[collection] = struct{}{}
}

func isBadRequest(err error) bool {
	httpErr, ok := err.(*odataHTTPError)
	return ok && httpErr.statusCode == http.StatusBadRequest
}

func isGUID(value string) bool {
	if len(value) != 36 {
		return false
	}
	for index, character := range value {
		switch index {
		case 8, 13, 18, 23:
			if character != '-' {
				return false
			}
		default:
			if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f') || (character >= 'A' && character <= 'F')) {
				return false
			}
		}
	}
	return true
}

func truncateBytes(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max] + "…"
}
