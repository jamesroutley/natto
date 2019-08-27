// Package main implements Natto's CLI.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"

	"github.com/jamesroutley/natto/crawler"
)

var concur = flag.Int("concurrency", 10, "Number of concurrent requests.")

var debug = flag.Bool("debug", true, "Enable debug logging.")

var noIndent = flag.Bool("no-indent", false,
	"Print site map without indentation.")

func init() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr,
			"Usage: natto [-concurrency] [-debug] [-no-indent] URL")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
	}
}

func setupLogging(debug bool) {
	if !debug {
		log.SetOutput(ioutil.Discard)
	}
}

func throw(message string) {
	flag.Usage()
	fmt.Fprintln(os.Stderr, "Error:", message)
	os.Exit(1)
}

func main() {
	flag.Parse()
	setupLogging(*debug)
	rawurl := flag.Arg(0)
	u, err := url.ParseRequestURI(rawurl)
	if err != nil {
		message := fmt.Sprintf(
			"Could not validate url '%s'.\n%v.\n", rawurl, err)
		throw(message)
	}
	// c := crawler.New(u, *concur)
	crawler.Crawl(u, *concur)
}
