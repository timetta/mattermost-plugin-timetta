package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestLoadEntityByStringKey(t *testing.T) {
	t.Parallel()
	var receivedFilter string
	var receivedExpand string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer secret-token" {
			t.Errorf("unexpected authorization header: %q", request.Header.Get("Authorization"))
		}
		receivedFilter = request.URL.Query().Get("$filter")
		receivedExpand = request.URL.Query().Get("$expand")
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"value":[{"name":"Broken login","description":"Details","state":{"name":"In progress"}}]}`))
	}))
	defer server.Close()

	config := &configuration{APIURL: server.URL, BearerToken: "secret-token"}
	link := entityLink{
		URL: "https://app.test/issue/DEV-4608/main", Identifier: "DEV-4608",
		Route: "issue", Collection: "Issues",
	}
	entity, err := newODataClient().loadEntity(context.Background(), config, link)
	if err != nil {
		t.Fatal(err)
	}
	if receivedFilter != "key eq 'DEV-4608'" {
		t.Fatalf("unexpected OData filter: %q", receivedFilter)
	}
	if receivedExpand != "state($select=name),type($select=name),priority($select=name),project($select=name)" {
		t.Fatalf("unexpected OData expand: %q", receivedExpand)
	}
	if entity["name"] != "Broken login" || entity["description"] != "Details" {
		t.Fatalf("unexpected entity: %#v", entity)
	}
}

func TestEntityExpansionsForNonIssue(t *testing.T) {
	t.Parallel()
	expansions := entityExpansions("Projects", true)
	if len(expansions) != 1 || expansions[0] != "state($select=name)" {
		t.Fatalf("unexpected expansions: %#v", expansions)
	}
}

func TestBuildODataURLForGUID(t *testing.T) {
	t.Parallel()
	link := entityLink{Identifier: "1fbd9446-3a05-4cf9-bb29-252fc8721c10", Collection: "Issues"}
	requestURL, collection, err := buildODataURL("https://api.timetta.com", link, true)
	if err != nil {
		t.Fatal(err)
	}
	if collection || !strings.Contains(requestURL, "/OData/Issues(1fbd9446-3a05-4cf9-bb29-252fc8721c10)") {
		t.Fatalf("unexpected request URL: %s", requestURL)
	}
	if !strings.Contains(requestURL, "%24expand=state") {
		t.Fatalf("expected State expansion in URL: %s", requestURL)
	}
}

func TestLoadEntityRemembersCollectionWithoutState(t *testing.T) {
	t.Parallel()
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requests.Add(1)
		if request.URL.Query().Has("$expand") {
			http.Error(writer, `property 'State' does not exist`, http.StatusBadRequest)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"name":"Ada Lovelace","description":"Engineer"}`))
	}))
	defer server.Close()

	client := newODataClient()
	config := &configuration{APIURL: server.URL}
	link := entityLink{Identifier: "1fbd9446-3a05-4cf9-bb29-252fc8721c10", Collection: "Users"}
	for range 2 {
		entity, err := client.loadEntity(context.Background(), config, link)
		if err != nil {
			t.Fatal(err)
		}
		if entity["name"] != "Ada Lovelace" {
			t.Fatalf("unexpected entity: %#v", entity)
		}
	}
	if actual := requests.Load(); actual != 3 {
		t.Fatalf("expected one fallback and one cached request, got %d requests", actual)
	}
}
