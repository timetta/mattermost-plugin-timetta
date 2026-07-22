package main

import "testing"

func TestBuildAttachmentDisplaysAvailableFields(t *testing.T) {
	t.Parallel()
	link := entityLink{
		URL: "https://app.test/issue/DEV-4608/main", Identifier: "DEV-4608",
		Route: "issue", Collection: "Issues",
	}
	attachment := buildAttachment(link, map[string]any{
		"name":        "Broken login",
		"code":        "DEV-4608",
		"state":       map[string]any{"name": "In progress"},
		"type":        map[string]any{"name": "Bug"},
		"priority":    map[string]any{"name": "High"},
		"project":     map[string]any{"name": "Development"},
		"description": "This must not be displayed",
	})
	if attachment["title"] != "DEV-4608 — Broken login" || attachment["title_link"] != link.URL {
		t.Fatalf("unexpected attachment: %#v", attachment)
	}
	fields, ok := attachment["fields"].([]any)
	if !ok || len(fields) != 4 {
		t.Fatalf("unexpected fields: %#v", attachment["fields"])
	}
	expected := []struct {
		title string
		value string
	}{
		{title: "Состояние", value: "In progress"},
		{title: "Тип", value: "Bug"},
		{title: "Приоритет", value: "High"},
		{title: "Проект", value: "Development"},
	}
	for index, expectedField := range expected {
		field := fields[index].(map[string]any)
		if field["title"] != expectedField.title || field["value"] != expectedField.value {
			t.Fatalf("unexpected field at %d: %#v", index, field)
		}
	}
}

func TestBuildAttachmentOmitsUnavailableOptionalFields(t *testing.T) {
	t.Parallel()
	attachment := buildAttachment(entityLink{Identifier: "id"}, map[string]any{"name": "Name"})
	fields, ok := attachment["fields"].([]any)
	if !ok || len(fields) != 0 {
		t.Fatalf("expected no optional fields, got %#v", attachment["fields"])
	}
	if attachment["title"] != "Name" {
		t.Fatalf("unexpected title without code: %#v", attachment["title"])
	}
}

func TestReplaceTimettaAttachmentsPreservesForeignAttachments(t *testing.T) {
	t.Parallel()
	props := map[string]any{"attachments": []any{
		map[string]any{"title": "foreign"},
		map[string]any{"title": "old", previewMarker: true},
	}}
	replacement := map[string]any{"title": "new", previewMarker: true}
	props = replaceTimettaAttachments(props, []map[string]any{replacement})
	attachments := attachmentSlice(props["attachments"])
	if len(attachments) != 2 {
		t.Fatalf("expected foreign and replacement attachments, got %#v", attachments)
	}
}
