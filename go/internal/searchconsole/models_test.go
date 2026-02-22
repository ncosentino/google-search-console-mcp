package searchconsole_test

import (
	"testing"

	"github.com/ncosentino/google-search-console-mcp/go/internal/searchconsole"
)

func TestNewClient_InvalidJSON_ReturnsError(t *testing.T) {
	t.Parallel()
	_, err := searchconsole.NewClient([]byte("not valid json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestNewClient_EmptyJSON_ReturnsError(t *testing.T) {
	t.Parallel()
	_, err := searchconsole.NewClient([]byte("{}"))
	if err == nil {
		t.Fatal("expected error for empty JSON object, got nil")
	}
}

func TestSearchAnalyticsResponse_RowCount(t *testing.T) {
	t.Parallel()
	r := &searchconsole.SearchAnalyticsResponse{
		Rows: []searchconsole.SearchAnalyticsRow{
			{Clicks: 10, Impressions: 100, CTR: 0.1, Position: 3.5},
			{Clicks: 5, Impressions: 50, CTR: 0.1, Position: 5.0},
		},
		RowCount: 2,
	}
	if r.RowCount != len(r.Rows) {
		t.Errorf("RowCount = %d, want %d", r.RowCount, len(r.Rows))
	}
}
