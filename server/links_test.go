package main

import "testing"

func TestFindEntityLinks(t *testing.T) {
	t.Parallel()
	config := &configuration{
		FrontendURL:           "https://app.timetta.com",
		APIURL:                "https://api.timetta.com",
		MaxPreviewsPerPost:    5,
		RequestTimeoutSeconds: 5,
	}
	if err := config.prepare(); err != nil {
		t.Fatal(err)
	}
	message := "See [issue](https://app.timetta.com/issue/DEV-4608/main?navigation=my.dev), not https://other.test/issue/DEV-1"
	links := findEntityLinks(message, config)
	if len(links) != 1 {
		t.Fatalf("expected one link, got %d: %#v", len(links), links)
	}
	if links[0].Identifier != "DEV-4608" || links[0].Collection != "Issues" {
		t.Fatalf("unexpected link: %#v", links[0])
	}
}

func TestFindEntityLinksHonorsFrontendBasePath(t *testing.T) {
	t.Parallel()
	config := &configuration{
		FrontendURL:           "https://example.test/timetta",
		APIURL:                "https://api.example.test",
		MaxPreviewsPerPost:    5,
		RequestTimeoutSeconds: 5,
	}
	if err := config.prepare(); err != nil {
		t.Fatal(err)
	}
	links := findEntityLinks("https://example.test/timetta/issue/ABC-1/main https://example.test/issue/ABC-2", config)
	if len(links) != 1 || links[0].Identifier != "ABC-1" {
		t.Fatalf("unexpected links: %#v", links)
	}
}
