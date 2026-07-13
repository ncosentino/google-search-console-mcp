// Package searchconsole provides a client for the Google Search Console API.
package searchconsole

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2/google"
)

const (
	gscScope    = "https://www.googleapis.com/auth/webmasters.readonly"
	httpTimeout = 30 * time.Second

	// defaultSearchType is the effective search type when the caller omits
	// search_type, matching the upstream API's own documented default.
	defaultSearchType = "web"
)

// validSearchTypes are the upstream Search Console API's supported values for
// the search analytics "type" field. See
// https://developers.google.com/webmaster-tools/v1/searchanalytics/query.
var validSearchTypes = map[string]bool{
	"web":        true,
	"image":      true,
	"video":      true,
	"news":       true,
	"discover":   true,
	"googleNews": true,
}

// apiBaseURL is the base URL for the Google Search Console API.
// It is a variable so tests can override it to point at a local test server.
var apiBaseURL = "https://www.googleapis.com/webmasters/v3"

// apiRequestError is returned when the Search Console API responds with a non-2xx status.
type apiRequestError struct {
	StatusCode int
	Body       string
}

func (e *apiRequestError) Error() string {
	return fmt.Sprintf("Search Console API returned HTTP %d: %s", e.StatusCode, e.Body)
}

// Client calls the Google Search Console API.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a Client authenticated with the provided service account JSON.
func NewClient(serviceAccountJSON []byte) (*Client, error) {
	cfg, err := google.JWTConfigFromJSON(serviceAccountJSON, gscScope)
	if err != nil {
		return nil, fmt.Errorf("parsing service account JSON: %w", err)
	}
	base := cfg.Client(context.Background())
	base.Timeout = httpTimeout
	return &Client{httpClient: base}, nil
}

// NewTestClient is exported solely for use in package-level tests, including tests in
// other packages that need a Client backed by a fake HTTP server instead of real
// Google OAuth2/Search Console endpoints.
func NewTestClient(httpClient *http.Client) *Client {
	return &Client{httpClient: httpClient}
}

// SetTestAPIBaseURL overrides the package-level Search Console API base URL, returning
// a function that restores the original value. Exported solely for use in
// package-level tests, including tests in other packages.
func SetTestAPIBaseURL(newURL string) (restore func()) {
	orig := apiBaseURL
	apiBaseURL = newURL
	return func() { apiBaseURL = orig }
}

// QuerySearchAnalytics queries search analytics data for the given site.
// siteURL accepts any of: bare domain ("example.com"), URL ("https://example.com"),
// or canonical GSC form ("sc-domain:example.com", "https://example.com/").
// searchType filters results to one upstream-supported type ("web", "image",
// "video", "news", "discover", "googleNews"); an empty string omits the field
// from the outbound request entirely, which upstream defaults to "web".
func (c *Client) QuerySearchAnalytics(
	ctx context.Context,
	siteURL string,
	startDate, endDate string,
	dimensions []string,
	rowLimit int,
	searchType string,
) (*SearchAnalyticsResponse, error) {
	if searchType != "" && !validSearchTypes[searchType] {
		return nil, fmt.Errorf(
			"invalid search_type %q: must be one of web, image, video, news, discover, googleNews", searchType)
	}
	resolved := NormalizeSiteURL(siteURL)
	result, err := c.querySearchAnalyticsWithURL(ctx, resolved, startDate, endDate, dimensions, rowLimit, searchType)
	if err != nil {
		var apiErr *apiRequestError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusForbidden {
			slog.Info("site URL returned 403, attempting property resolution", "input", siteURL, "tried", resolved)
			resolvedURL, resolveErr := ResolveSiteURL(ctx, c, siteURL)
			if resolveErr != nil {
				return nil, err // return original 403 error
			}
			slog.Info("retrying with resolved property", "resolvedURL", resolvedURL)
			return c.querySearchAnalyticsWithURL(ctx, resolvedURL, startDate, endDate, dimensions, rowLimit, searchType)
		}
		return nil, err
	}
	return result, nil
}

func (c *Client) querySearchAnalyticsWithURL(
	ctx context.Context,
	siteURL string,
	startDate, endDate string,
	dimensions []string,
	rowLimit int,
	searchType string,
) (*SearchAnalyticsResponse, error) {
	if rowLimit <= 0 {
		rowLimit = 1000
	}
	reqBody := apiSearchAnalyticsRequest{
		StartDate:  startDate,
		EndDate:    endDate,
		Dimensions: dimensions,
		Type:       searchType,
		RowLimit:   rowLimit,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshalling request body: %w", err)
	}

	endpoint := fmt.Sprintf("%s/sites/%s/searchAnalytics/query",
		apiBaseURL, url.PathEscape(siteURL))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &apiRequestError{StatusCode: resp.StatusCode, Body: truncate(string(body), 300)}
	}

	var raw apiSearchAnalyticsResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing search analytics response: %w", err)
	}

	rows := make([]SearchAnalyticsRow, len(raw.Rows))
	for i, r := range raw.Rows {
		rows[i] = SearchAnalyticsRow(r)
	}

	effectiveSearchType := searchType
	if effectiveSearchType == "" {
		effectiveSearchType = defaultSearchType
	}

	return &SearchAnalyticsResponse{
		SiteURL:    siteURL,
		StartDate:  startDate,
		EndDate:    endDate,
		Dimensions: dimensions,
		SearchType: effectiveSearchType,
		RowCount:   len(rows),
		Rows:       rows,
		QueriedAt:  time.Now().UTC(),
	}, nil
}

// ListSites returns all Search Console properties accessible to the service account.
func (c *Client) ListSites(ctx context.Context) (*SiteList, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiBaseURL+"/sites", nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &apiRequestError{StatusCode: resp.StatusCode, Body: truncate(string(body), 300)}
	}

	var raw apiSiteListResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing sites response: %w", err)
	}

	sites := make([]Site, len(raw.SiteEntry))
	for i, s := range raw.SiteEntry {
		sites[i] = Site(s)
	}

	return &SiteList{Sites: sites, QueriedAt: time.Now().UTC()}, nil
}

// ListSitemaps returns submitted sitemaps for the given site.
// siteURL accepts any of: bare domain ("example.com"), URL ("https://example.com"),
// or canonical GSC form ("sc-domain:example.com", "https://example.com/").
func (c *Client) ListSitemaps(ctx context.Context, siteURL string) (*SitemapList, error) {
	resolved := NormalizeSiteURL(siteURL)
	result, err := c.listSitemapsWithURL(ctx, resolved)
	if err != nil {
		var apiErr *apiRequestError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusForbidden {
			slog.Info("site URL returned 403, attempting property resolution", "input", siteURL, "tried", resolved)
			resolvedURL, resolveErr := ResolveSiteURL(ctx, c, siteURL)
			if resolveErr != nil {
				return nil, err
			}
			slog.Info("retrying with resolved property", "resolvedURL", resolvedURL)
			return c.listSitemapsWithURL(ctx, resolvedURL)
		}
		return nil, err
	}
	return result, nil
}

func (c *Client) listSitemapsWithURL(ctx context.Context, siteURL string) (*SitemapList, error) {
	endpoint := fmt.Sprintf("%s/sites/%s/sitemaps", apiBaseURL, url.PathEscape(siteURL))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &apiRequestError{StatusCode: resp.StatusCode, Body: truncate(string(body), 300)}
	}

	var raw apiSitemapListResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing sitemaps response: %w", err)
	}

	sitemaps := make([]Sitemap, len(raw.Sitemap))
	for i, s := range raw.Sitemap {
		warnings, warningsDiagnostic := parseSitemapCount("warnings", s.Warnings)
		errorsCount, errorsDiagnostic := parseSitemapCount("errors", s.Errors)
		diagnostics := make([]SitemapFieldDiagnostic, 0, 2)
		if warningsDiagnostic != nil {
			diagnostics = append(diagnostics, *warningsDiagnostic)
		}
		if errorsDiagnostic != nil {
			diagnostics = append(diagnostics, *errorsDiagnostic)
		}
		sm := Sitemap{
			Path:            s.Path,
			IsPending:       s.IsPending,
			IsSitemapsIndex: s.IsSitemapsIndex,
			Type:            s.Type,
			Warnings:        warnings,
			Errors:          errorsCount,
			Diagnostics:     diagnostics,
		}
		if t, err := time.Parse(time.RFC3339, s.LastSubmitted); err == nil {
			sm.LastSubmitted = t
		}
		if t, err := time.Parse(time.RFC3339, s.LastDownloaded); err == nil {
			sm.LastDownloaded = t
		}
		sitemaps[i] = sm
	}

	return &SitemapList{SiteURL: siteURL, Sitemaps: sitemaps, QueriedAt: time.Now().UTC()}, nil
}

func parseSitemapCount(field string, raw json.RawMessage) (*int64, *SitemapFieldDiagnostic) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return nil, nil
	}

	if value, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
		return &value, nil
	}

	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		if value, parseErr := strconv.ParseInt(strings.TrimSpace(text), 10, 64); parseErr == nil {
			return &value, nil
		}
	}

	return nil, &SitemapFieldDiagnostic{
		Field:    field,
		RawValue: trimmed,
		Warning:  "value is not an integer or numeric string",
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
