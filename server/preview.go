package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

const previewMarker = "timetta_link_preview"

func buildAttachment(link entityLink, entity map[string]any) map[string]any {
	name := stringProperty(entity, "name")
	if name == "" {
		name = link.Identifier
	}
	title := name
	if code := stringProperty(entity, "code"); code != "" {
		title = code + " — " + name
	}

	fields := make([]any, 0, 4)
	if state := stateName(entity); state != "" {
		fields = append(fields, attachmentField("Состояние", state))
	}
	if link.Collection == "Issues" {
		for _, field := range []struct {
			property string
			title    string
		}{
			{property: "type", title: "Тип"},
			{property: "priority", title: "Приоритет"},
			{property: "project", title: "Проект"},
		} {
			if value := relatedName(entity, field.property); value != "" {
				fields = append(fields, attachmentField(field.title, value))
			}
		}
	}

	return map[string]any{
		"fallback":    fmt.Sprintf("Timetta: %s", title),
		"color":       "#3155a4",
		"author_name": "Timetta",
		"title":       title,
		"title_link":  link.URL,
		"fields":      fields,
		"footer":      "Timetta",
		previewMarker: true,
	}
}

func stringProperty(entity map[string]any, property string) string {
	value, exists := jsonProperty(entity, property)
	if !exists || value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return strings.TrimSpace(text)
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func stateName(entity map[string]any) string {
	return relatedName(entity, "state")
}

func relatedName(entity map[string]any, property string) string {
	value, exists := jsonProperty(entity, property)
	if !exists || value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return strings.TrimSpace(text)
	}
	if object, ok := value.(map[string]any); ok {
		return stringProperty(object, "name")
	}
	return ""
}

func attachmentField(title, value string) map[string]any {
	return map[string]any{"title": title, "value": value, "short": true}
}

func jsonProperty(entity map[string]any, property string) (any, bool) {
	if value, exists := entity[property]; exists {
		return value, true
	}
	for key, value := range entity {
		if strings.EqualFold(key, property) {
			return value, true
		}
	}
	return nil, false
}

func replaceTimettaAttachments(props map[string]any, replacements []map[string]any) map[string]any {
	if props == nil {
		props = make(map[string]any)
	}
	existing := attachmentSlice(props["attachments"])
	filtered := make([]any, 0, len(existing)+len(replacements))
	for _, attachment := range existing {
		if object, ok := attachment.(map[string]any); ok {
			if marked, _ := object[previewMarker].(bool); marked {
				continue
			}
		}
		filtered = append(filtered, attachment)
	}
	for _, attachment := range replacements {
		filtered = append(filtered, attachment)
	}
	if len(filtered) == 0 {
		delete(props, "attachments")
	} else {
		props["attachments"] = filtered
	}
	return props
}

func attachmentSlice(value any) []any {
	if value == nil {
		return nil
	}
	if values, ok := value.([]any); ok {
		return values
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	var decoded []any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		return nil
	}
	return decoded
}

func (p *Plugin) loadAttachments(config *configuration, links []entityLink) []map[string]any {
	attachments := make([]map[string]any, len(links))
	var wait sync.WaitGroup
	for index, link := range links {
		wait.Add(1)
		go func() {
			defer wait.Done()
			ctx, cancel := context.WithTimeout(context.Background(), config.requestTimeout())
			defer cancel()
			entity, err := p.odata.loadEntity(ctx, config, link)
			if err != nil {
				p.API.LogWarn("Unable to create Timetta link preview", "url", link.URL, "error", err.Error())
				return
			}
			attachments[index] = buildAttachment(link, entity)
		}()
	}
	wait.Wait()

	result := make([]map[string]any, 0, len(attachments))
	for _, attachment := range attachments {
		if attachment != nil {
			result = append(result, attachment)
		}
	}
	return result
}
