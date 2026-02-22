// Package searchconsole provides a client for the Google Search Console API.
package searchconsole

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2/google"
)

const (
	baseURL    = "https://www.googleapis.com/webmasters/v3"
	gscScope   = "https://www.googleapis.com/auth/webmasters.readonly"
	httpTimeout = 30 * time.Second
)

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

// QuerySearchAnalytics queries search analytics data for the given site.
func (c *Client) QuerySearchAnalytics(
	ctx context.Context,
	siteURL string,
	startDate, endDate string,
	dimensions []string,
	rowLimit int,
) (*SearchAnalyticsResponse, error) {
	if rowLimit <= 0 {
		rowLimit = 1000
	}
	reqBody := apiSearchAnalyticsRequest{
		StartDate:  startDate,
		EndDate:    endDate,
		Dimensions: dimensions,
		RowLimit:   rowLimit,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshalling request body: %w", err)
	}

	endpoint := fmt.Sprintf("%s/sites/%s/searchAnalytics/query",
		baseURL, url.PathEscape(siteURL))
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
		return nil, fmt.Errorf("Search Console API returned HTTP %d: %s",
			resp.StatusCode, truncate(string(body), 300))
	}

	var raw apiSearchAnalyticsResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing search analytics response: %w", err)
	}

	rows := make([]SearchAnalyticsRow, len(raw.Rows))
	for i, r := range raw.Rows {
		rows[i] = SearchAnalyticsRow(r)
	}

	return &SearchAnalyticsResponse{
		SiteURL:     siteURL,
		StartDate:   startDate,
		EndDate:     endDate,
		Dimensions:  dimensions,
		RowCount:    len(rows),
		Rows:        rows,
		QueriedAt:   time.Now().UTC(),
	}, nil
}

// ListSites returns all Search Console properties accessible to the service account.
func (c *Client) ListSites(ctx context.Context) (*SiteList, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/sites", nil)
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
		return nil, fmt.Errorf("Search Console API returned HTTP %d: %s",
			resp.StatusCode, truncate(string(body), 300))
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
func (c *Client) ListSitemaps(ctx context.Context, siteURL string) (*SitemapList, error) {
	endpoint := fmt.Sprintf("%s/sites/%s/sitemaps", baseURL, url.PathEscape(siteURL))
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
		return nil, fmt.Errorf("Search Console API returned HTTP %d: %s",
			resp.StatusCode, truncate(string(body), 300))
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
			Warnings:        s.Warnings,
			Errors:          s.Errors,
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

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
