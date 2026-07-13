package searchconsole

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListSitemaps_NormalizesStringNullAndMalformedCounts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sitemap": [
				{"path":"https://example.com/sitemap.xml","warnings":"2","errors":0},
				{"path":"https://example.com/news.xml","warnings":null,"errors":"invalid"},
				{"path":"https://example.com/images.xml"}
			]
		}`))
	}))
	defer srv.Close()
	defer SetTestAPIBaseURL(srv.URL)()

	client := NewTestClient(srv.Client())
	result, err := client.ListSitemaps(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("ListSitemaps returned an error: %v", err)
	}
	if len(result.Sitemaps) != 3 {
		t.Fatalf("sitemap count = %d, want 3", len(result.Sitemaps))
	}

	if result.Sitemaps[0].Warnings == nil || *result.Sitemaps[0].Warnings != 2 {
		t.Fatalf("numeric string warnings = %v, want 2", result.Sitemaps[0].Warnings)
	}
	if result.Sitemaps[0].Errors == nil || *result.Sitemaps[0].Errors != 0 {
		t.Fatalf("numeric errors = %v, want 0", result.Sitemaps[0].Errors)
	}
	if result.Sitemaps[1].Warnings != nil {
		t.Errorf("null warnings = %v, want nil", result.Sitemaps[1].Warnings)
	}
	if result.Sitemaps[1].Errors != nil {
		t.Errorf("malformed errors = %v, want nil", result.Sitemaps[1].Errors)
	}
	if len(result.Sitemaps[1].Diagnostics) != 1 {
		t.Fatalf("diagnostics count = %d, want 1", len(result.Sitemaps[1].Diagnostics))
	}
	diagnostic := result.Sitemaps[1].Diagnostics[0]
	if diagnostic.Field != "errors" || diagnostic.RawValue != `"invalid"` {
		t.Errorf("diagnostic = %+v, want errors with preserved raw value", diagnostic)
	}
	if result.Sitemaps[2].Warnings != nil || result.Sitemaps[2].Errors != nil {
		t.Errorf("missing counters should remain nil: %+v", result.Sitemaps[2])
	}
}
