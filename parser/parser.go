// Package parser implements an HTML parser which finds a webpage's static
// assets and internal and external links.
package parser

import (
	"io"
	"net/url"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// PageDetails describes certain properties of an HTML page.
type PageDetails struct {
	InternalLinks []string `json:"internal_links"`
	ExternalLinks []string `json:"external_links"`
	Assets        []string `json:"assets"`
}

// ParseWebpage parses the HTML webpage.
func ParseWebpage(pageURL *url.URL, webpage io.Reader) (PageDetails, error) {
	details := PageDetails{}
	tokenizer := html.NewTokenizer(webpage)
	for {
		if tokenizer.Next() == html.ErrorToken {
			// TODO: handle this error
			return details, nil
		}
		token := tokenizer.Token()
		// The attributes we're interested in all exist in HTML opening tags.
		if token.Type != html.StartTagToken {
			continue
		}
		switch token.DataAtom {
		case atom.Link:
			rawurl, ok := getAttribute(token.Attr, "href")
			if !ok {
				continue
			}
			u, _ := url.Parse(rawurl)
			resolvedURL := pageURL.ResolveReference(u)
			details.Assets = append(details.Assets, resolvedURL.String())
		case atom.A:
			rawurl, ok := getAttribute(token.Attr, "href")
			if !ok {
				continue
			}
			u, err := url.Parse(rawurl)
			if err != nil {
				return PageDetails{}, err
			}
			resolvedURL := pageURL.ResolveReference(u)
			if resolvedURL.Host == pageURL.Host {
				details.InternalLinks = append(details.InternalLinks,
					resolvedURL.String())
			} else {
				details.ExternalLinks = append(details.ExternalLinks,
					resolvedURL.String())
			}
		}
	}
}

// getAttribute returns the value of first attribute named k.
// If no attribute is found, the second return value returns false.
// If more than one attribute share the Key k, the value associated with the
// first attribute in the slice is returned.
func getAttribute(a []html.Attribute, k string) (string, bool) {
	for _, attr := range a {
		if attr.Key == k {
			return attr.Val, true
		}
	}
	return "", false
}
