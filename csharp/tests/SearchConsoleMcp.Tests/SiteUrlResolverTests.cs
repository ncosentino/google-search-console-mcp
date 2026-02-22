using SearchConsoleMcp.SearchConsole;
using Xunit;

namespace SearchConsoleMcp.Tests;

public sealed class SiteUrlResolverTests
{
    [Theory]
    [InlineData("sc-domain:devleader.ca",        "sc-domain:devleader.ca")]
    [InlineData("devleader.ca",                  "sc-domain:devleader.ca")]
    [InlineData("  devleader.ca  ",              "sc-domain:devleader.ca")]
    [InlineData("https://devleader.ca",          "sc-domain:devleader.ca")]
    [InlineData("https://www.devleader.ca",      "sc-domain:devleader.ca")]
    [InlineData("http://devleader.ca",           "sc-domain:devleader.ca")]
    [InlineData("http://www.devleader.ca",       "sc-domain:devleader.ca")]
    [InlineData("https://www.devleader.ca/",     "https://www.devleader.ca/")]
    [InlineData("https://devleader.ca/",         "https://devleader.ca/")]
    [InlineData("https://www.devleader.ca/blog/","https://www.devleader.ca/blog/")]
    [InlineData("https://www.devleader.ca/blog", "https://www.devleader.ca/blog/")]
    public void Normalize_ReturnsExpectedForm(string input, string expected)
    {
        var result = SiteUrlResolver.Normalize(input);
        Assert.Equal(expected, result);
    }

    [Fact]
    public void FindBestMatch_PrefersDomainProperty_WhenAvailable()
    {
        var sites = new[]
        {
            new SiteEntry("sc-domain:devleader.ca", "siteFullUser"),
            new SiteEntry("https://www.devleader.ca/", "siteOwner"),
        };

        var result = SiteUrlResolver.FindBestMatch("devleader.ca", sites);
        Assert.Equal("sc-domain:devleader.ca", result);
    }

    [Fact]
    public void FindBestMatch_FallsBackToUrlPrefix_WhenNoDomainProperty()
    {
        var sites = new[]
        {
            new SiteEntry("https://www.devleader.ca/", "siteOwner"),
        };

        var result = SiteUrlResolver.FindBestMatch("devleader.ca", sites);
        Assert.Equal("https://www.devleader.ca/", result);
    }

    [Fact]
    public void FindBestMatch_PrefersDomainProperty_OverMultipleUrlPrefixes()
    {
        var sites = new[]
        {
            new SiteEntry("https://www.devleader.ca/", "siteOwner"),
            new SiteEntry("https://devleader.ca/", "siteOwner"),
            new SiteEntry("sc-domain:devleader.ca", "siteFullUser"),
        };

        var result = SiteUrlResolver.FindBestMatch("https://www.devleader.ca", sites);
        Assert.Equal("sc-domain:devleader.ca", result);
    }

    [Fact]
    public void FindBestMatch_ReturnsNull_WhenNoMatchFound()
    {
        var sites = new[]
        {
            new SiteEntry("sc-domain:other.com", "siteFullUser"),
        };

        var result = SiteUrlResolver.FindBestMatch("devleader.ca", sites);
        Assert.Null(result);
    }

    [Fact]
    public void FindBestMatch_WithUrlInput_ExtractsApexDomain()
    {
        var sites = new[]
        {
            new SiteEntry("sc-domain:devleader.ca", "siteFullUser"),
        };

        var result = SiteUrlResolver.FindBestMatch("https://www.devleader.ca/", sites);
        Assert.Equal("sc-domain:devleader.ca", result);
    }
}
