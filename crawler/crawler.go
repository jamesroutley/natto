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
	"time"
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

// Crawler implements a simple web crawler. Should be instantiated with the
// New function.
type Crawler struct {
	startURL *url.URL
	siteMap  *SiteMap
	idx      index
	concur int
}

// New instantiates and returns a new Crawler.
func New(u *url.URL, concur int) *Crawler {
	siteMap := &SiteMap{Pages: make(map[string]parser.PageDetails)}
	idx := make(index)
	idx.add(u)
	c := &Crawler{u, siteMap, idx, concur}
	return c
}

// Crawl returns the SiteMap of a website.
func (c *Crawler) Crawl() *SiteMap {
	count := counter{}
	pagesToVisit := make(chan *url.URL, c.concur)
	results := make(chan namedPageDetails, c.concur)
	for i := 0; i < c.concur; i++ {
		go crawlPage(c.startURL, pagesToVisit, results)
	}
	for {
		select {
		case result := <-results:
			for _, link := range result.details.InternalLinks {
				linkURL, _ := url.Parse(link)
				c.idx.add(linkURL)
			}
			c.siteMap.Pages[result.url.String()] = result.details
			count.decr()
		default:
			unvisitedLinks := c.idx.getUnvisitedLinks()
			numUnvisitedLinks := len(unvisitedLinks)
			if numUnvisitedLinks == 0 && count.val == 0 {
				return c.siteMap
			} else if numUnvisitedLinks > 0 {
				l := unvisitedLinks[0]
				c.idx.markVisited(l)
				pagesToVisit <- l
				count.incr()
			} else {
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

// crawlPage visits and parses pages sent to 'urls', writing results to 'results'
func crawlPage(startURL *url.URL, urls <-chan *url.URL, results chan<- namedPageDetails) {
	for {
		select {
		case u := <-urls:
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
			results <- namedDetails
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}
}

// counter counts items.
type counter struct {
	val int
}

// incr increments the counter by 1.
func (c *counter) incr() {
	c.val++
}

// decr decrements the counter by 1. Panics if val goes below 0.
func (c *counter) decr() {
	nextVal := c.val - 1
	if nextVal < 0 {
		panic("Counter's val cannot go below 0.")
	}
	c.val = nextVal
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
	log.Printf("Fetching HTML from '%s'", u)
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
