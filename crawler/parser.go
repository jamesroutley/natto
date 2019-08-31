package crawler

import (
	"io"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func parseWebpage(webpage io.Reader) ([]string, error) {
	tokenizer := html.NewTokenizer(webpage)
	var urls []string
	for {
		if tokenizer.Next() == html.ErrorToken {
			err := tokenizer.Err()
			// End of the file - return the URLs we've found
			if err == io.EOF {
				return urls, nil
			}
			// Otherwise, return the error
			return nil, err
		}
		token := tokenizer.Token()

		// The attributes we're interested in all exist in HTML opening tags.
		if token.Type != html.StartTagToken {
			continue
		}

		switch token.DataAtom {
		case atom.A, atom.Link:
			url := getAttribute(token.Attr, "href")
			if url != "" {
				urls = append(urls, url)
			}
		}
	}
}

// getAttribute returns the value of first attribute named k.
// If more than one attribute share the Key k, the value associated with the
// first attribute in the slice is returned.
func getAttribute(a []html.Attribute, k string) string {
	for _, attr := range a {
		if attr.Key == k {
			return attr.Val
		}
	}
	return ""
}
