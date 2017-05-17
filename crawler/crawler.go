// Package crawler implements a simple web crawler which maps a single site.
package crawler

import (
	// "fmt"
	"bytes"
	"github.com/jamesroutley/natto/parser"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	// "time"
)

// SiteMap describes the properties of a website.
type SiteMap struct {
	Pages map[string]parser.PageDetails `json:"pages"`
}

// namedPageDetails stores the url of a webpage, along with its details.
type namedPageDetails struct {
	url     *url.URL
	details parser.PageDetails
}

// Crawl returns the SiteMap of a website.
func Crawl(startURL *url.URL, concur int) *SiteMap {
	index := make(map[string]bool)
	siteMap := &SiteMap{Pages: make(map[string]parser.PageDetails)}
	count := counter{}

	// Create worker goroutines
	pagesToVisit := make(chan *url.URL, concur)
	results := make(chan namedPageDetails, 100)//concur)
	for i := 0; i < concur; i++ {
		go crawlPage(startURL, pagesToVisit, results)
	}

	pagesToVisit <- startURL
	index[startURL.String()] = true
	count.incr()
	for result := range results {
		log.Println("reading result from ", result.url.String())
		siteMap.Pages[result.url.String()] = result.details
		for _, link := range result.details.InternalLinks {
			linkURL, err := url.Parse(link)
			if err != nil {
				log.Fatal("Could not parse url '%s': %v", link, err)
			}
			linkURL.Fragment = ""
			linkURL.RawQuery = ""
			if !index[linkURL.String()] {
				log.Println("Adding", linkURL.String(), "to pagesToVisit")
				log.Println("len(pagesToVisit)", len(pagesToVisit))
				pagesToVisit <- linkURL
				index[linkURL.String()] = true
				count.incr()
			}
		}
		count.decr()
		if count.val == 0 {
			break
		}
		log.Println("looping round: len(results)", len(results))
	}
	return siteMap
}

// crawlPage visits and parses pages sent to 'urls', writing results to 'results'
func crawlPage(startURL *url.URL, urls <-chan *url.URL, results chan<- namedPageDetails) {
	for u := range urls {
		body, err := getWebpage(u)
		if err != nil {
			log.Printf("Error reading webpage '%s': %v", u, err)
			return
		}
		details := parser.ParseWebpage(startURL, bytes.NewReader(body))
		namedDetails := namedPageDetails{
			url:     u,
			details: details,
		}
		log.Println("returning results from", u.String())
		log.Println("len(results)", len(results))
		results <- namedDetails
	}
}

// counter counts items.
type counter struct {
	val int
}

// incr increments the counter by 1.
func (c *counter) incr() {
	c.val++
	log.Println("incremented count:", c.val)
}

// decr decrements the counter by 1. Panics if val goes below 0.
func (c *counter) decr() {
	nextVal := c.val - 1
	if nextVal < 0 {
		panic("Counter's val cannot go below 0.")
	}
	c.val = nextVal
	log.Println("decremented count:", c.val)
}

// index is a map of internal links found on the website.
// A link's value indicates whether the link has been visited.
type index map[string]bool

// add adds a new url to the index.
// URL query and fragment are stripped
func (i index) add(u *url.URL) {
	// Remove fragment and query from URL. Index stores a list of visited
	// pages. It is assumed that the contents of 'url.com' are the same as
	// 'url.com?key=val#section'
	u.Fragment = ""
	u.RawQuery = ""
	if _, ok := i[u.String()]; !ok {
		log.Printf("Adding %v to index", u)
		i[u.String()] = false
	}
}

func (i index) markVisited(u *url.URL) {
	u.Fragment = ""
	u.RawQuery = ""
	log.Printf("Marking %v as visited", u)
	i[u.String()] = true
}

// getUnvisitedLinks returns the unvisited links in the index.
func (i index) getUnvisitedLinks() []*url.URL {
	unvisitedLinks := []*url.URL{}
	for k, v := range i {
		if !v {
			u, _ := url.Parse(k)
			unvisitedLinks = append(unvisitedLinks, u)
		}
	}
	return unvisitedLinks
}

// getWebpage gets and returns the contents of a webpage.
func getWebpage(u *url.URL) ([]byte, error) {
	// log.Printf("Fetching HTML from '%s'", u)
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
