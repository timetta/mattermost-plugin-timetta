package main

import "strings"

// entityCollections maps the first frontend URL segment to an OData collection.
// Add new supported Timetta entity routes here.
var entityCollections = map[string]string{
	"issue":         "Issues",
	"projects":      "Projects",
	"deals":         "Deals",
	"employees":     "Users",
	"timesheets":    "TimeSheets",
	"time-off":      "TimeOffRequests",
	"expenses":      "ExpenseRequests",
	"organizations": "Organizations",
}

// collectionStringKeys is only used when the frontend identifier is not a Guid.
// Collections absent from this map support Guid identifiers only.
var collectionStringKeys = map[string]string{
	"Issues":        "key",
	"Projects":      "key",
	"Deals":         "key",
	"Users":         "code",
	"Organizations": "key",
}

func entityCollection(route string) (string, bool) {
	collection, exists := entityCollections[strings.ToLower(route)]
	return collection, exists
}

func collectionStringKey(collection string) string {
	return collectionStringKeys[collection]
}
