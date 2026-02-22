package searchconsole_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ncosentino/google-search-console-mcp/go/internal/searchconsole"
)

func TestNormalizeSiteURL(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input string
		want  string
	}{
		// Already canonical -- pass through unchanged.
		{"sc-domain:devleader.ca", "sc-domain:devleader.ca"},
		// Bare domain -- infer sc-domain.
		{"devleader.ca", "sc-domain:devleader.ca"},
		{"  devleader.ca  ", "sc-domain:devleader.ca"},
		// HTTPS URL without trailing slash -- infer sc-domain.
		{"https://devleader.ca", "sc-domain:devleader.ca"},
		{"https://www.devleader.ca", "sc-domain:devleader.ca"},
		// HTTP URL without trailing slash -- infer sc-domain.
		{"http://devleader.ca", "sc-domain:devleader.ca"},
		{"http://www.devleader.ca", "sc-domain:devleader.ca"},
		// URL with trailing slash at root -- treat as URL-prefix.
		{"https://www.devleader.ca/", "https://www.devleader.ca/"},
		{"https://devleader.ca/", "https://devleader.ca/"},
		// URL with non-root path -- treat as URL-prefix; ensure trailing slash.
		{"https://www.devleader.ca/blog/", "https://www.devleader.ca/blog/"},
		{"https://www.devleader.ca/blog", "https://www.devleader.ca/blog/"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			got := searchconsole.NormalizeSiteURL(tc.input)
			if got != tc.want {
				t.Errorf("NormalizeSiteURL(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// mockSiteLister implements SiteLister for testing.
type mockSiteLister struct {
	sites []searchconsole.Site
	err   error
}

func (m *mockSiteLister) ListSites(_ context.Context) (*searchconsole.SiteList, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &searchconsole.SiteList{Sites: m.sites}, nil
}

func TestResolveSiteURL_DomainProperty_Found(t *testing.T) {
	t.Parallel()
	lister := &mockSiteLister{sites: []searchconsole.Site{
		{SiteURL: "sc-domain:devleader.ca", PermissionLevel: "siteFullUser"},
	}}
	got, err := searchconsole.ResolveSiteURL(context.Background(), lister, "https://www.devleader.ca")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "sc-domain:devleader.ca" {
		t.Errorf("got %q, want %q", got, "sc-domain:devleader.ca")
	}
}

func TestResolveSiteURL_URLPrefixFallback_WhenNoDomainProperty(t *testing.T) {
	t.Parallel()
	lister := &mockSiteLister{sites: []searchconsole.Site{
		{SiteURL: "https://www.devleader.ca/", PermissionLevel: "siteFullUser"},
	}}
	got, err := searchconsole.ResolveSiteURL(context.Background(), lister, "devleader.ca")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "https://www.devleader.ca/" {
		t.Errorf("got %q, want %q", got, "https://www.devleader.ca/")
	}
}

func TestResolveSiteURL_PrefersDomainPropertyOverURLPrefix(t *testing.T) {
	t.Parallel()
	lister := &mockSiteLister{sites: []searchconsole.Site{
		{SiteURL: "https://www.devleader.ca/", PermissionLevel: "siteFullUser"},
		{SiteURL: "sc-domain:devleader.ca", PermissionLevel: "siteFullUser"},
	}}
	got, err := searchconsole.ResolveSiteURL(context.Background(), lister, "devleader.ca")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "sc-domain:devleader.ca" {
		t.Errorf("got %q, want %q", got, "sc-domain:devleader.ca")
	}
}

func TestResolveSiteURL_InputWithTrailingSlash_FindsDomainProperty(t *testing.T) {
	t.Parallel()
	lister := &mockSiteLister{sites: []searchconsole.Site{
		{SiteURL: "sc-domain:devleader.ca", PermissionLevel: "siteFullUser"},
	}}
	got, err := searchconsole.ResolveSiteURL(context.Background(), lister, "https://www.devleader.ca/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "sc-domain:devleader.ca" {
		t.Errorf("got %q, want %q", got, "sc-domain:devleader.ca")
	}
}

func TestResolveSiteURL_NoMatch_ReturnsErrorWithAccessibleSites(t *testing.T) {
	t.Parallel()
	lister := &mockSiteLister{sites: []searchconsole.Site{
		{SiteURL: "sc-domain:other.com", PermissionLevel: "siteFullUser"},
	}}
	_, err := searchconsole.ResolveSiteURL(context.Background(), lister, "devleader.ca")
	if err == nil {
		t.Fatal("expected error for no matching property, got nil")
	}
}

func TestResolveSiteURL_ListError_Propagates(t *testing.T) {
	t.Parallel()
	lister := &mockSiteLister{err: fmt.Errorf("network error")}
	_, err := searchconsole.ResolveSiteURL(context.Background(), lister, "devleader.ca")
	if err == nil {
		t.Fatal("expected error from ListSites, got nil")
	}
}
