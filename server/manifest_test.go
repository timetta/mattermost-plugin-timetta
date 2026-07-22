package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
)

func TestPluginManifestIsValid(t *testing.T) {
	t.Parallel()
	contents, err := os.ReadFile("../plugin.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest model.Manifest
	if err := json.Unmarshal(contents, &manifest); err != nil {
		t.Fatal(err)
	}
	if err := manifest.IsValid(); err != nil {
		t.Fatalf("invalid Mattermost plugin manifest: %v", err)
	}
}
