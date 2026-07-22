package main

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"
)

type configuration struct {
	FrontendURL           string
	APIURL                string
	BearerToken           string
	RequestTimeoutSeconds int
	MaxPreviewsPerPost    int

	frontendBase *url.URL
}

func (c *configuration) clone() *configuration {
	if c == nil {
		return nil
	}
	clone := *c
	if c.frontendBase != nil {
		base := *c.frontendBase
		clone.frontendBase = &base
	}
	return &clone
}

func (c *configuration) prepare() error {
	if strings.TrimSpace(c.FrontendURL) == "" {
		c.FrontendURL = "https://app.timetta.com"
	}
	if strings.TrimSpace(c.APIURL) == "" {
		c.APIURL = "https://api.timetta.com"
	}
	if c.RequestTimeoutSeconds == 0 {
		c.RequestTimeoutSeconds = 5
	}
	if c.MaxPreviewsPerPost == 0 {
		c.MaxPreviewsPerPost = 5
	}
	if c.RequestTimeoutSeconds < 1 || c.RequestTimeoutSeconds > 30 {
		return fmt.Errorf("RequestTimeoutSeconds must be between 1 and 30")
	}
	if c.MaxPreviewsPerPost < 1 || c.MaxPreviewsPerPost > 10 {
		return fmt.Errorf("MaxPreviewsPerPost must be between 1 and 10")
	}

	frontend, err := parseBaseURL(c.FrontendURL, "FrontendURL")
	if err != nil {
		return err
	}
	apiURL, err := parseBaseURL(c.APIURL, "APIURL")
	if err != nil {
		return err
	}
	c.FrontendURL = strings.TrimRight(frontend.String(), "/")
	c.APIURL = strings.TrimRight(apiURL.String(), "/")
	c.frontendBase = frontend
	return nil
}

func parseBaseURL(value, settingName string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("%s must be an absolute HTTP(S) URL", settingName)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("%s must use HTTP or HTTPS", settingName)
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, fmt.Errorf("%s must not contain credentials, a query, or a fragment", settingName)
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	parsed.RawPath = ""
	return parsed, nil
}

func (c *configuration) requestTimeout() time.Duration {
	return time.Duration(c.RequestTimeoutSeconds) * time.Second
}

type configurationStore struct {
	mu     sync.RWMutex
	config *configuration
}

func (s *configurationStore) get() *configuration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.clone()
}

func (s *configurationStore) set(config *configuration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = config
}
