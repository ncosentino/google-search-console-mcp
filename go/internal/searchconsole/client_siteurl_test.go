package searchconsole

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// newTestClient creates a Client with a custom HTTP client for testing (bypasses Google auth).
func newTestClient(hc *http.Client) *Client {
	return &Client{httpClient: hc}
}

// setAPIBaseURL overrides the API base URL for a test and returns a restore function.
func setAPIBaseURL(newURL string) (restore func()) {
	orig := apiBaseURL
	apiBaseURL = newURL
	return func() { apiBaseURL = orig }
}

func TestQuerySearchAnalytics_NormalizesBareInputToSCDomain(t *testing.T) {
	var requestedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"rows":[]}`))
	}))
	defer srv.Close()
	defer setAPIBaseURL(srv.URL)()

	client := newTestClient(srv.Client())
	_, err := client.QuerySearchAnalytics(context.Background(), "devleader.ca", "2025-01-01", "2025-12-31", nil, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The path should contain the sc-domain form, not the bare domain.
	want := "/sites/" + url.PathEscape("sc-domain:devleader.ca") + "/searchAnalytics/query"
	if requestedPath != want {
		t.Errorf("request path = %q, want %q", requestedPath, want)
	}
}

func TestQuerySearchAnalytics_On403_RetriesWithResolvedURL(t *testing.T) {
	callCount := 0

	// The input URL-prefix form triggers a 403; ListSites reveals the sc-domain property.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		urlPrefixPath := "/sites/" + url.PathEscape("https://www.devleader.ca/") + "/searchAnalytics/query"
		scDomainPath := "/sites/" + url.PathEscape("sc-domain:devleader.ca") + "/searchAnalytics/query"
		listSitesPath := "/sites"

		switch {
		case r.URL.EscapedPath() == urlPrefixPath:
			// First attempt: return 403.
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":{"code":403,"message":"forbidden"}}`))
		case r.URL.Path == listSitesPath:
			// ListSites called by ResolveSiteURL.
			resp := apiSiteListResponse{
				SiteEntry: []apiSiteEntry{
					{SiteURL: "sc-domain:devleader.ca", PermissionLevel: "siteFullUser"},
				},
			}
			b, _ := json.Marshal(resp)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(b)
		case r.URL.EscapedPath() == scDomainPath:
			// Retry with resolved URL: succeed.
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"rows":[]}`))
		default:
			t.Errorf("unexpected request path: %s (escaped: %s)", r.URL.Path, r.URL.EscapedPath())
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	defer setAPIBaseURL(srv.URL)()

	client := newTestClient(srv.Client())
	resp, err := client.QuerySearchAnalytics(
		context.Background(), "https://www.devleader.ca/", "2025-01-01", "2025-12-31", nil, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response after retry")
	}
	// Expect: 1 (403) + 1 (ListSites) + 1 (retry) = 3 HTTP calls.
	if callCount != 3 {
		t.Errorf("expected 3 HTTP calls, got %d", callCount)
	}
}

func TestListSitemaps_On403_RetriesWithResolvedURL(t *testing.T) {
	callCount := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		urlPrefixSitemapsPath := "/sites/" + url.PathEscape("https://www.devleader.ca/") + "/sitemaps"
		scDomainSitemapsPath := "/sites/" + url.PathEscape("sc-domain:devleader.ca") + "/sitemaps"
		listSitesPath := "/sites"

		switch {
		case r.URL.EscapedPath() == urlPrefixSitemapsPath:
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":{"code":403}}`))
		case r.URL.Path == listSitesPath:
			resp := apiSiteListResponse{
				SiteEntry: []apiSiteEntry{
					{SiteURL: "sc-domain:devleader.ca", PermissionLevel: "siteFullUser"},
				},
			}
			b, _ := json.Marshal(resp)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(b)
		case r.URL.EscapedPath() == scDomainSitemapsPath:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"sitemap":[]}`))
		default:
			t.Errorf("unexpected request path: %s (escaped: %s)", r.URL.Path, r.URL.EscapedPath())
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	defer setAPIBaseURL(srv.URL)()

	client := newTestClient(srv.Client())
	result, err := client.ListSitemaps(context.Background(), "https://www.devleader.ca/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result after retry")
	}
	if callCount != 3 {
		t.Errorf("expected 3 HTTP calls, got %d", callCount)
	}
}
