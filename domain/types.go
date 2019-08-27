package domain

import "net/url"

type Job struct {
	URL    *url.URL
	Errors []string
}
