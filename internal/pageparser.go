package internal

import (
	"net/url"
	"strings"

	"github.com/DrakkarStorm/deadlinkr/logger"
	"github.com/DrakkarStorm/deadlinkr/model"
	"github.com/PuerkitoBio/goquery"
)

// PageParserService implements the PageParser interface
type PageParserService struct {
	LinkChecker     LinkChecker // Exported for stats access
	urlProcessor    URLProcessor
	excludeHtmlTags string
	onlyInternal    bool
}

// NewPageParserService creates a new PageParserService
func NewPageParserService(linkChecker LinkChecker, urlProcessor URLProcessor, excludeHtmlTags string, onlyInternal bool) *PageParserService {
	return &PageParserService{
		LinkChecker:     linkChecker,
		urlProcessor:    urlProcessor,
		excludeHtmlTags: excludeHtmlTags,
		onlyInternal:    onlyInternal,
	}
}

// ParsePage fetches and parses a web page
func (pp *PageParserService) ParsePage(pageURL string) (*goquery.Document, error) {
	retry := 3
	resp, err := pp.LinkChecker.FetchWithRetry(pageURL, retry)

	if err != nil {
		logger.Errorf("Failed to fetch %s after %d retries: %s", pageURL, retry, err)
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Errorf("Error closing response body for %s: %s", pageURL, err)
		}
	}()

	if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		return nil, nil
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		logger.Errorf("Error parsing HTML from %s: %s", pageURL, err)
		return nil, err
	}

	return doc, nil
}

// ExtractLinks extracts links from a parsed document
func (pp *PageParserService) ExtractLinks(baseUrlParsed *url.URL, pageURL string, doc *goquery.Document) []model.LinkResult {
	pageLinks := []model.LinkResult{}

	doc.Find("body a[href]").Not(pp.excludeHtmlTags).Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists || href == "" || strings.HasPrefix(href, "#") {
			logger.Debugf("Skipping link due to missing href or #: %s", href)
			return
		}

		linkURL := pp.resolveAndFilterURL(baseUrlParsed, pageURL, href)
		if linkURL == nil {
			logger.Debugf("Skipping link due to invalid URL resolution: %s", href)
			return
		}

		isExternal := baseUrlParsed.Hostname() != linkURL.Hostname()

		if pp.onlyInternal && isExternal {
			return
		}

		if pp.urlProcessor.ShouldSkipURL(baseUrlParsed, linkURL) {
			logger.Debugf("Skipping link due to pattern match: %s", href)
			return
		}

		status, errMsg := pp.LinkChecker.CheckLink(linkURL.String())

		linkResult := model.LinkResult{
			SourceURL:  pageURL,
			TargetURL:  linkURL.String(),
			Status:     status,
			Error:      errMsg,
			IsExternal: isExternal,
		}

		pageLinks = append(pageLinks, linkResult)
	})

	return pageLinks
}

// resolveAndFilterURL resolves and filters a URL
func (pp *PageParserService) resolveAndFilterURL(baseUrlParsed *url.URL, pageURL, href string) *url.URL {
	linkURL, err := pp.urlProcessor.ResolveURL(pageURL, href)
	if err != nil || pp.urlProcessor.ShouldSkipURL(baseUrlParsed, linkURL) {
		return nil
	}
	return linkURL
}

// SetConfig updates the parser configuration
func (pp *PageParserService) SetConfig(excludeHtmlTags string, onlyInternal bool) {
	pp.excludeHtmlTags = excludeHtmlTags
	pp.onlyInternal = onlyInternal
}