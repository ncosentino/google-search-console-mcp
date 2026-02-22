// Package searchconsole provides types for the Google Search Console API.
package searchconsole

import "time"

// SearchAnalyticsRow is a single row from a search analytics query.
type SearchAnalyticsRow struct {
	Keys        []string `json:"keys,omitempty"`
	Clicks      float64  `json:"clicks"`
	Impressions float64  `json:"impressions"`
	CTR         float64  `json:"ctr"`
	Position    float64  `json:"position"`
}

// SearchAnalyticsResponse is the parsed result of a search analytics query.
type SearchAnalyticsResponse struct {
	SiteURL     string               `json:"siteUrl"`
	StartDate   string               `json:"startDate"`
	EndDate     string               `json:"endDate"`
	Dimensions  []string             `json:"dimensions,omitempty"`
	RowCount    int                  `json:"rowCount"`
	Rows        []SearchAnalyticsRow `json:"rows"`
	QueriedAt   time.Time            `json:"queriedAt"`
}

// Site represents a Search Console property.
type Site struct {
	SiteURL        string `json:"siteUrl"`
	PermissionLevel string `json:"permissionLevel"`
}

// SiteList is the result of listing Search Console properties.
type SiteList struct {
	Sites     []Site    `json:"sites"`
	QueriedAt time.Time `json:"queriedAt"`
}

// Sitemap represents a submitted sitemap.
type Sitemap struct {
	Path          string    `json:"path"`
	LastSubmitted time.Time `json:"lastSubmitted,omitempty"`
	IsPending     bool      `json:"isPending"`
	IsSitemapsIndex bool    `json:"isSitemapsIndex"`
	Type          string    `json:"type"`
	LastDownloaded time.Time `json:"lastDownloaded,omitempty"`
	Warnings      int64     `json:"warnings"`
	Errors        int64     `json:"errors"`
}

// SitemapList is the result of listing sitemaps for a property.
type SitemapList struct {
	SiteURL   string    `json:"siteUrl"`
	Sitemaps  []Sitemap `json:"sitemaps"`
	QueriedAt time.Time `json:"queriedAt"`
}

// --- Search Console API raw response types ---

type apiSiteEntry struct {
	SiteURL        string `json:"siteUrl"`
	PermissionLevel string `json:"permissionLevel"`
}

type apiSiteListResponse struct {
	SiteEntry []apiSiteEntry `json:"siteEntry"`
}

type apiSitemapEntry struct {
	Path           string `json:"path"`
	LastSubmitted  string `json:"lastSubmitted"`
	IsPending      bool   `json:"isPending"`
	IsSitemapsIndex bool  `json:"isSitemapsIndex"`
	Type           string `json:"type"`
	LastDownloaded string `json:"lastDownloaded"`
	Warnings       int64  `json:"warnings"`
	Errors         int64  `json:"errors"`
}

type apiSitemapListResponse struct {
	Sitemap []apiSitemapEntry `json:"sitemap"`
}

type apiSearchAnalyticsRequest struct {
	StartDate  string   `json:"startDate"`
	EndDate    string   `json:"endDate"`
	Dimensions []string `json:"dimensions,omitempty"`
	RowLimit   int      `json:"rowLimit,omitempty"`
}

type apiSearchAnalyticsRow struct {
	Keys        []string `json:"keys"`
	Clicks      float64  `json:"clicks"`
	Impressions float64  `json:"impressions"`
	CTR         float64  `json:"ctr"`
	Position    float64  `json:"position"`
}

type apiSearchAnalyticsResponse struct {
	Rows []apiSearchAnalyticsRow `json:"rows"`
}
