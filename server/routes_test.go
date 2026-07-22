package main

import "testing"

func TestEntityCollections(t *testing.T) {
	t.Parallel()
	expected := map[string]string{
		"issue":         "Issues",
		"projects":      "Projects",
		"deals":         "Deals",
		"employees":     "Users",
		"timesheets":    "TimeSheets",
		"time-off":      "TimeOffRequests",
		"expenses":      "ExpenseRequests",
		"organizations": "Organizations",
	}
	if len(entityCollections) != len(expected) {
		t.Fatalf("expected %d routes, got %d", len(expected), len(entityCollections))
	}
	for route, collection := range expected {
		actual, exists := entityCollection(route)
		if !exists || actual != collection {
			t.Fatalf("route %q: expected %q, got %q", route, collection, actual)
		}
	}
}
