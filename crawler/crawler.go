// Package crawler implements a simple web crawler which maps a single site.
package crawler

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/jamesroutley/natto/domain"
	"github.com/jamesroutley/natto/parser"
	"github.com/jamesroutley/natto/queue"
)

var (
	indexNew  = map[string]bool{}
	indexLock sync.Mutex
)

// TODO: return dead links from here
func Crawl(startURL *url.URL, workers int) {
	startURL = normaliseURL(startURL)
	q := queue.NewQueue()

	q.Add(&domain.Job{URL: startURL})
	fmt.Println(startURL.String())
	indexNew[startURL.String()] = true

	// TODO: unbuffer
	deadLinks := make(chan string, 1000)

	for i := 0; i < workers; i++ {
		go crawlPageNew(startURL, q, deadLinks)
	}

	q.Wait()

	close(deadLinks)

	for link := range deadLinks {
		fmt.Println(link)
	}
}

func crawlPageNew(startURL *url.URL, q *queue.Queue, deadLinks chan string) {
	for {
		message := q.Next()
		job := message.Job

		fmt.Printf("Crawling %s\n", job.URL)

		body, err := getWebpage(job.URL)
		if err != nil {
			log.Printf("Error reading webpage '%s': %v", job.URL, err)
			return
		}
		details, err := parser.ParseWebpage(startURL, bytes.NewReader(body))
		if err != nil {
			log.Print(err)
			// TODO: log error in job
			q.Delete(message)
			continue
		}
		for _, link := range details.ExternalLinks {
			deadLinks <- link
		}

		for _, link := range details.InternalLinks {

			u, err := url.Parse(link)
			if err != nil {
				fmt.Printf("Error parsing %s: %v", link, err)
				continue
			}
			u = normaliseURL(u)

			indexLock.Lock()
			if seen := indexNew[u.String()]; !seen {
				q.Add(&domain.Job{URL: u})
				indexNew[u.String()] = true
			}
			indexLock.Unlock()
		}

		q.Delete(message)
	}
}

func normaliseURL(u *url.URL) *url.URL {
	u.Fragment = ""
	u.RawQuery = ""

	return u
}

// getWebpage gets and returns the contents of a webpage.
func getWebpage(u *url.URL) ([]byte, error) {
	// log.Printf("Fetching HTML from '%s'", u)
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// TODO: Don't read this in one go
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
