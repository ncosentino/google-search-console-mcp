package searchconsole

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestQuerySearchAnalytics_SearchTypeVideo_SendsTypeVideoUpstream is the
// RED-first test for issue #22: search_type "video" must be forwarded to the
// upstream Search Console API as the request body's "type" field.
func TestQuerySearchAnalytics_SearchTypeVideo_SendsTypeVideoUpstream(t *testing.T) {
	var gotBody apiSearchAnalyticsRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decoding request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"rows":[]}`))
	}))
	defer srv.Close()
	defer SetTestAPIBaseURL(srv.URL)()

	client := NewTestClient(srv.Client())
	_, err := client.QuerySearchAnalytics(
		context.Background(), "devleader.ca", "2025-01-01", "2025-12-31", nil, 10, "video")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotBody.Type != "video" {
		t.Errorf(`request body "type" = %q, want %q`, gotBody.Type, "video")
	}
}

// TestQuerySearchAnalytics_SearchTypeOmitted_PreservesExistingRequest confirms
// that omitting search_type (empty string) sends no "type" field at all,
// preserving the exact request shape that existed before #22 rather than
// silently rewriting every existing caller's request to add "type":"web".
// The response, however, still reports the effective default explicitly.
func TestQuerySearchAnalytics_SearchTypeOmitted_PreservesExistingRequest(t *testing.T) {
	var rawBody map[string]json.RawMessage
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&rawBody); err != nil {
			t.Fatalf("decoding request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"rows":[]}`))
	}))
	defer srv.Close()
	defer SetTestAPIBaseURL(srv.URL)()

	client := NewTestClient(srv.Client())
	resp, err := client.QuerySearchAnalytics(
		context.Background(), "devleader.ca", "2025-01-01", "2025-12-31", nil, 10, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, present := rawBody["type"]; present {
		t.Errorf(`request body unexpectedly contains "type": %s`, rawBody["type"])
	}
	if resp.SearchType != "web" {
		t.Errorf("resp.SearchType = %q, want %q (default)", resp.SearchType, "web")
	}
}

// TestQuerySearchAnalytics_SearchTypeWeb_SendsTypeWebExplicitly confirms an
// explicit "web" is accepted and forwarded as-is -- distinct from omission,
// which sends no "type" field at all (see the preceding test).
func TestQuerySearchAnalytics_SearchTypeWeb_SendsTypeWebExplicitly(t *testing.T) {
	var gotBody apiSearchAnalyticsRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"rows":[]}`))
	}))
	defer srv.Close()
	defer SetTestAPIBaseURL(srv.URL)()

	client := NewTestClient(srv.Client())
	resp, err := client.QuerySearchAnalytics(
		context.Background(), "devleader.ca", "2025-01-01", "2025-12-31", nil, 10, "web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody.Type != "web" {
		t.Errorf(`request body "type" = %q, want %q`, gotBody.Type, "web")
	}
	if resp.SearchType != "web" {
		t.Errorf("resp.SearchType = %q, want %q", resp.SearchType, "web")
	}
}

// TestQuerySearchAnalytics_AllValidSearchTypes_AreAccepted table-drives every
// upstream-supported value (per
// https://developers.google.com/webmaster-tools/v1/searchanalytics/query) to
// guard against silently narrowing the accepted set in a future refactor.
func TestQuerySearchAnalytics_AllValidSearchTypes_AreAccepted(t *testing.T) {
	for _, st := range []string{"web", "image", "video", "news", "discover", "googleNews"} {
		t.Run(st, func(t *testing.T) {
			var gotBody apiSearchAnalyticsRequest
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewDecoder(r.Body).Decode(&gotBody)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"rows":[]}`))
			}))
			defer srv.Close()
			defer SetTestAPIBaseURL(srv.URL)()

			client := NewTestClient(srv.Client())
			resp, err := client.QuerySearchAnalytics(
				context.Background(), "devleader.ca", "2025-01-01", "2025-12-31", nil, 10, st)
			if err != nil {
				t.Fatalf("unexpected error for search_type %q: %v", st, err)
			}
			if gotBody.Type != st {
				t.Errorf(`request body "type" = %q, want %q`, gotBody.Type, st)
			}
			if resp.SearchType != st {
				t.Errorf("resp.SearchType = %q, want %q", resp.SearchType, st)
			}
		})
	}
}

// TestQuerySearchAnalytics_SearchTypeVideo_ComposesWithDimensions covers this
// issue's own motivating example (video-search performance grouped by page):
// search_type and dimensions are independent request fields, but this proves
// they compose correctly on the wire together rather than one silently
// overwriting or suppressing the other.
func TestQuerySearchAnalytics_SearchTypeVideo_ComposesWithDimensions(t *testing.T) {
	var gotBody apiSearchAnalyticsRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"rows":[]}`))
	}))
	defer srv.Close()
	defer SetTestAPIBaseURL(srv.URL)()

	client := NewTestClient(srv.Client())
	resp, err := client.QuerySearchAnalytics(
		context.Background(), "devleader.ca", "2025-01-01", "2025-12-31", []string{"page"}, 10, "video")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotBody.Type != "video" {
		t.Errorf(`request body "type" = %q, want %q`, gotBody.Type, "video")
	}
	if len(gotBody.Dimensions) != 1 || gotBody.Dimensions[0] != "page" {
		t.Errorf(`request body "dimensions" = %v, want ["page"]`, gotBody.Dimensions)
	}
	if resp.SearchType != "video" {
		t.Errorf("resp.SearchType = %q, want %q", resp.SearchType, "video")
	}
}


// an unsupported search_type value is rejected before any network call is
// made, rather than forwarded upstream to fail with a confusing API error.
func TestQuerySearchAnalytics_InvalidSearchType_RejectedWithoutHTTPCall(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"rows":[]}`))
	}))
	defer srv.Close()
	defer SetTestAPIBaseURL(srv.URL)()

	client := NewTestClient(srv.Client())
	_, err := client.QuerySearchAnalytics(
		context.Background(), "devleader.ca", "2025-01-01", "2025-12-31", nil, 10, "youtube")
	if err == nil {
		t.Fatal("expected an error for invalid search_type, got nil")
	}
	if callCount != 0 {
		t.Errorf("expected 0 HTTP calls for an invalid search_type, got %d", callCount)
	}
}
