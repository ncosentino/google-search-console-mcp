package searchconsole

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestQuerySearchAnalytics_NormalizesBareInputToSCDomain(t *testing.T) {
	var requestedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"rows":[]}`))
	}))
	defer srv.Close()
	defer SetTestAPIBaseURL(srv.URL)()

	client := NewTestClient(srv.Client())
	_, err := client.QuerySearchAnalytics(context.Background(), "devleader.ca", "2025-01-01", "2025-12-31", nil, 10, "")
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
	defer SetTestAPIBaseURL(srv.URL)()

	client := NewTestClient(srv.Client())
	resp, err := client.QuerySearchAnalytics(
		context.Background(), "https://www.devleader.ca/", "2025-01-01", "2025-12-31", nil, 10, "")
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
	defer SetTestAPIBaseURL(srv.URL)()

	client := NewTestClient(srv.Client())
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

func TestInspectURL_On403_RetriesWithResolvedURL(t *testing.T) {
	callCount := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		if r.Method == http.MethodGet && r.URL.Path == "/sites" {
			resp := apiSiteListResponse{
				SiteEntry: []apiSiteEntry{
					{SiteURL: "sc-domain:devleader.ca", PermissionLevel: "siteFullUser"},
				},
			}
			b, _ := json.Marshal(resp)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(b)
			return
		}

		if r.Method != http.MethodPost || r.URL.Path != "/urlInspection/index:inspect" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var request apiURLInspectionRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		switch request.SiteURL {
		case "https://www.devleader.ca/":
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":{"code":403}}`))
		case "sc-domain:devleader.ca":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"inspectionResult":{"indexStatusResult":{"verdict":"PASS"}}}`))
		default:
			t.Errorf("unexpected siteUrl: %s", request.SiteURL)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	defer SetTestAPIBaseURL(srv.URL)()

	client := NewTestClient(srv.Client())
	result, err := client.InspectURL(
		context.Background(),
		"https://www.devleader.ca/",
		"https://www.devleader.ca/example",
		"",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result after retry")
	}
	if result.SiteURL != "sc-domain:devleader.ca" {
		t.Errorf("result siteUrl = %q, want sc-domain:devleader.ca", result.SiteURL)
	}
	if callCount != 3 {
		t.Errorf("expected 3 HTTP calls, got %d", callCount)
	}
}
