package output

import (
	"encoding/json"

	"github.com/dkd-dobberkau/squirrel/internal/analyzer"
)

// RenderJSON returns the categorized projects as a JSON string.
func RenderJSON(data analyzer.CategorizedProjects) (string, error) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
