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
	"github.com/jamesroutley/natto/queue"
)

var (
	indexNew  = map[string]bool{}
	indexLock sync.Mutex
)

// TODO: return dead links from here
func Crawl(url string, workers int) {
	startURL, err := normaliseURL(url, url)
	if err != nil {
		log.Fatal(err)
	}
	q := queue.NewQueue()

	q.Add(&domain.Job{URL: startURL})
	indexNew[startURL] = true

	// TODO: unbuffer
	results := make(chan *domain.Job, 1000)

	for i := 0; i < workers; i++ {
		go crawl(startURL, q, results)
	}

	q.Wait()

	close(results)

	for job := range results {
		fmt.Println(job.URL)
	}
}

func crawl(startURL string, q *queue.Queue, results chan *domain.Job) {
	for {
		message := q.Next()
		job := message.Job

		fmt.Println(job.URL)

		body, err := getWebpage(job.URL)
		if err != nil {
			log.Printf("Error reading webpage '%s': %v", job.URL, err)
			return
		}
		links, err := parseWebpage(bytes.NewReader(body))
		if err != nil {
			log.Print(err)
			// TODO: log error in job
			q.Delete(message)
			continue
		}

		for _, link := range links {
			var err error
			link, err = normaliseURL(startURL, link)
			if err != nil {
				log.Printf("Could not normalise link '%s', skipping: %v", link, err)
				continue
			}

			indexLock.Lock()
			if seen := indexNew[link]; !seen {
				q.Add(&domain.Job{URL: link})
				indexNew[link] = true
			}
			indexLock.Unlock()
		}

		results <- job

		q.Delete(message)
	}
}

func normaliseURL(baseURL, link string) (string, error) {
	b, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	u, err := url.Parse(link)
	if err != nil {
		return "", err
	}
	u.Fragment = ""
	u.RawQuery = ""

	if u.Host == "" {
		u.Host = b.Host
		u.Scheme = b.Scheme
	}

	return u.String(), nil
}

func isInternalURL(baseURL, link string) (bool, error) {
	b, err := url.Parse(baseURL)
	if err != nil {
		return false, err
	}
	u, err := url.Parse(link)
	if err != nil {
		return false, err
	}

	return b.Host == u.Host, nil
}

// getWebpage gets and returns the contents of a webpage.
func getWebpage(url string) ([]byte, error) {
	// log.Printf("Fetching HTML from '%s'", u)
	resp, err := http.Get(url)
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
