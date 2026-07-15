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
	"strings"
	"time"

	"golang.org/x/oauth2/google"
)

const (
	gscScope          = "https://www.googleapis.com/auth/webmasters.readonly"
	httpTimeout       = 30 * time.Second
	apiErrorBodyLimit = 300

	// defaultSearchType is the effective search type when the caller omits
	// search_type, matching the upstream API's own documented default.
	defaultSearchType   = "web"
	defaultLanguageCode = "en-US"
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

var (
	// apiBaseURL serves the Sites, Sitemaps, and Search Analytics resources.
	apiBaseURL = "https://www.googleapis.com/webmasters/v3"

	// urlInspectionAPIBaseURL serves the URL Inspection resource.
	urlInspectionAPIBaseURL = "https://searchconsole.googleapis.com/v1"
)

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

// SetTestAPIBaseURL overrides both Search Console API base URLs, returning a
// function that restores the original values. Exported solely for package tests.
func SetTestAPIBaseURL(newURL string) (restore func()) {
	origAPI := apiBaseURL
	origInspectionAPI := urlInspectionAPIBaseURL
	apiBaseURL = newURL
	urlInspectionAPIBaseURL = newURL
	return func() {
		apiBaseURL = origAPI
		urlInspectionAPIBaseURL = origInspectionAPI
	}
}

func withResolvedSiteURL[T any](
	ctx context.Context,
	client *Client,
	input string,
	operation func(string) (T, error),
) (T, error) {
	resolved := NormalizeSiteURL(input)
	result, err := operation(resolved)
	if err == nil {
		return result, nil
	}

	var apiErr *apiRequestError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusForbidden {
		var zero T
		return zero, err
	}

	slog.Info("site URL returned 403, attempting property resolution", "input", input, "tried", resolved)
	resolvedURL, resolveErr := ResolveSiteURL(ctx, client, input)
	if resolveErr != nil {
		var zero T
		return zero, err
	}

	slog.Info("retrying with resolved property", "resolvedURL", resolvedURL)
	return operation(resolvedURL)
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

	return withResolvedSiteURL(ctx, c, siteURL, func(resolved string) (*SearchAnalyticsResponse, error) {
		return c.querySearchAnalyticsWithURL(
			ctx, resolved, startDate, endDate, dimensions, rowLimit, searchType)
	})
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
		return nil, &apiRequestError{StatusCode: resp.StatusCode, Body: truncateAPIErrorBody(string(body))}
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
		return nil, &apiRequestError{StatusCode: resp.StatusCode, Body: truncateAPIErrorBody(string(body))}
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
	return withResolvedSiteURL(ctx, c, siteURL, func(resolved string) (*SitemapList, error) {
		return c.listSitemapsWithURL(ctx, resolved)
	})
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
		return nil, &apiRequestError{StatusCode: resp.StatusCode, Body: truncateAPIErrorBody(string(body))}
	}

	var raw apiSitemapListResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing sitemaps response: %w", err)
	}

	sitemaps := make([]Sitemap, len(raw.Sitemap))
	for i, s := range raw.Sitemap {
		sm := Sitemap{
			Path:            s.Path,
			IsPending:       s.IsPending,
			IsSitemapsIndex: s.IsSitemapsIndex,
			Type:            s.Type,
			Warnings:        s.Warnings.Value,
			Errors:          s.Errors.Value,
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

// InspectURL returns Google's indexed status and available per-URL enhancement
// information for one URL under the given Search Console property.
func (c *Client) InspectURL(
	ctx context.Context,
	siteURL string,
	inspectionURL string,
	languageCode string,
) (*URLInspectionResponse, error) {
	inspectionURL = strings.TrimSpace(inspectionURL)
	if inspectionURL == "" {
		return nil, errors.New("inspection_url is required")
	}

	languageCode = strings.TrimSpace(languageCode)
	if languageCode == "" {
		languageCode = defaultLanguageCode
	}

	return withResolvedSiteURL(ctx, c, siteURL, func(resolved string) (*URLInspectionResponse, error) {
		return c.inspectURLWithSiteURL(ctx, resolved, inspectionURL, languageCode)
	})
}

func (c *Client) inspectURLWithSiteURL(
	ctx context.Context,
	siteURL string,
	inspectionURL string,
	languageCode string,
) (*URLInspectionResponse, error) {
	requestBody := apiURLInspectionRequest{
		SiteURL:       siteURL,
		InspectionURL: inspectionURL,
		LanguageCode:  languageCode,
	}
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("marshalling request body: %w", err)
	}

	endpoint := urlInspectionAPIBaseURL + "/urlInspection/index:inspect"
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
		return nil, &apiRequestError{StatusCode: resp.StatusCode, Body: truncateAPIErrorBody(string(body))}
	}

	var raw apiURLInspectionResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing URL inspection response: %w", err)
	}
	inspectionResult := bytes.TrimSpace(raw.InspectionResult)
	if len(inspectionResult) == 0 || bytes.Equal(inspectionResult, []byte("null")) {
		return nil, fmt.Errorf("parsing URL inspection response: inspectionResult is missing")
	}

	return &URLInspectionResponse{
		SiteURL:          siteURL,
		InspectionURL:    inspectionURL,
		LanguageCode:     languageCode,
		InspectionResult: raw.InspectionResult,
		QueriedAt:        time.Now().UTC(),
	}, nil
}

func truncateAPIErrorBody(s string) string {
	if len(s) <= apiErrorBodyLimit {
		return s
	}
	return s[:apiErrorBodyLimit] + "..."
}
