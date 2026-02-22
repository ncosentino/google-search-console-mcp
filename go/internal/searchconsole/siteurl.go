// Package searchconsole provides a client for the Google Search Console API.
package searchconsole

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// SiteLister is the minimal interface required by ResolveSiteURL to look up accessible
// Search Console properties. *Client satisfies this interface automatically.
type SiteLister interface {
	ListSites(ctx context.Context) (*SiteList, error)
}

// NormalizeSiteURL converts a user-supplied site reference into its canonical GSC format.
//
// Normalization rules:
//   - Already "sc-domain:*" -- returned unchanged.
//   - URL with scheme and no trailing slash at root (e.g. "https://example.com") -- converted to "sc-domain:<apex>".
//   - URL with trailing slash at root (e.g. "https://example.com/") or a non-root path -- treated as a URL-prefix
//     property; trailing slash is added if absent.
//   - Bare domain (no scheme, no sc-domain prefix) -- converted to "sc-domain:<apex>".
func NormalizeSiteURL(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return input
	}

	// Already in canonical domain-property format.
	if strings.HasPrefix(input, "sc-domain:") {
		return input
	}

	// Input has a URL scheme.
	if strings.Contains(input, "://") {
		u, err := url.Parse(input)
		if err == nil {
			hasNonRootPath := u.Path != "" && u.Path != "/"
			hasTrailingSlash := strings.HasSuffix(input, "/")

			if hasNonRootPath || hasTrailingSlash {
				// Caller signalled URL-prefix intent -- keep as-is, ensure trailing slash.
				if !hasTrailingSlash {
					return input + "/"
				}
				return input
			}
			// Root URL without trailing slash -- infer sc-domain.
			return "sc-domain:" + extractApexDomain(u.Host)
		}
	}

	// Bare domain (e.g. "devleader.ca").
	return "sc-domain:" + input
}

// ResolveSiteURL resolves a user-supplied site reference against the accessible GSC
// properties by querying the API. It is used as a fallback after the normalized form
// fails with a 403.
//
// Resolution preference: sc-domain property > URL-prefix property for the same apex domain.
// If no match is found the error message lists all accessible properties.
func ResolveSiteURL(ctx context.Context, lister SiteLister, input string) (string, error) {
	sites, err := lister.ListSites(ctx)
	if err != nil {
		return "", fmt.Errorf("resolving site URL: listing accessible properties: %w", err)
	}

	apex := extractApexFromInput(input)

	// Prefer sc-domain property for the apex domain.
	for _, s := range sites.Sites {
		if strings.EqualFold(s.SiteURL, "sc-domain:"+apex) {
			return s.SiteURL, nil
		}
	}

	// Fall back to URL-prefix property whose hostname maps to the same apex.
	for _, s := range sites.Sites {
		if !strings.HasPrefix(s.SiteURL, "http") {
			continue
		}
		u, parseErr := url.Parse(s.SiteURL)
		if parseErr != nil {
			continue
		}
		if extractApexDomain(u.Host) == apex {
			return s.SiteURL, nil
		}
	}

	accessible := make([]string, len(sites.Sites))
	for i, s := range sites.Sites {
		accessible[i] = s.SiteURL
	}
	return "", fmt.Errorf(
		"no matching GSC property found for %q -- accessible properties: %v",
		input, accessible)
}

// extractApexFromInput returns the apex domain from any supported input format.
func extractApexFromInput(input string) string {
	input = strings.TrimSpace(input)
	if strings.HasPrefix(input, "sc-domain:") {
		return strings.TrimPrefix(input, "sc-domain:")
	}
	if strings.Contains(input, "://") {
		if u, err := url.Parse(input); err == nil {
			return extractApexDomain(u.Host)
		}
	}
	// Bare domain -- strip any path component.
	if idx := strings.Index(input, "/"); idx != -1 {
		input = input[:idx]
	}
	return input
}

// extractApexDomain strips the www. prefix and port from a hostname.
func extractApexDomain(host string) string {
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}
	return strings.TrimPrefix(host, "www.")
}
