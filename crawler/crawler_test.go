package crawler

import (
	"fmt"
	"github.com/jamesroutley/natto/parser"
	"log"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"testing"
)

const (
	PORT = 4039
)

// startServer starts a server fixture used in testing.
// Returns a function which can be used to kill the server after testing is
// complete
func startServer(t *testing.T, port int) func() {
	addr := fmt.Sprintf("localhost:%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	fs := http.FileServer(http.Dir("./testdata/www-example"))
	t.Logf("Starting server at 'localhost:%d'", port)
	go http.Serve(listener, fs)

	killServer := func() {
		t.Logf("Killing server")
		listener.Close()
	}
	return killServer
}

func TestCrawl(t *testing.T) {
	addr := fmt.Sprintf("http://localhost:%d", PORT)
	expected := SiteMap{
		Pages: map[string]parser.PageDetails{
			addr: parser.PageDetails{
				InternalLinks: []string{
					addr + "/about.html",
					addr + "/blog.html",
				},
			},
			addr + "/about.html": parser.PageDetails{
				InternalLinks: []string{
					addr + "/blog.html",
				},
				Assets: []string{
					addr + "/css/example.css",
				},
			},
			addr + "/blog.html": parser.PageDetails{
				ExternalLinks: []string{
					"https://google.com",
					"https://bbc.com",
				},
				Assets: []string{
					addr + "/css/example.css",
				},
			},
		},
	}
	killServer := startServer(t, PORT)
	defer killServer()
	u := mustParse(t, addr)
	siteMap := Crawl(u, 3)
	if !reflect.DeepEqual(expected, *siteMap) {
		t.Fatalf("Expected %v, got %v", expected, siteMap)
	}
}

func TestIndexAdd(t *testing.T) {
	testCases := []struct {
		index         index
		toAdd         []*url.URL
		expectedindex index
	}{
		{
			index{},
			[]*url.URL{mustParse(t, "https://example.com")},
			index{"https://example.com": false},
		},
		{
			index{},
			[]*url.URL{mustParse(t, "https://example.com?query=foo#fragment")},
			index{"https://example.com": false},
		},
		{
			index{},
			[]*url.URL{mustParse(t, "https://example.com/path")},
			index{"https://example.com/path": false},
		},
	}
	for _, testCase := range testCases {
		for _, item := range testCase.toAdd {
			testCase.index.add(item)
		}
		if !reflect.DeepEqual(testCase.index, testCase.expectedindex) {
			t.Fatalf("Expected %v, got %v", testCase.index,
				testCase.expectedindex)
		}
	}
}

func TestGetUnvisitedLinks(t *testing.T) {
	testCases := []struct {
		idx      index
		expected []*url.URL
	}{
		{
			index{
				"http://example.com":       true,
				"http://example.com/about": true,
				"http://example.com/blog":  true,
			},
			[]*url.URL{},
		},
		{
			index{
				"http://example.com":       true,
				"http://example.com/about": true,
				"http://example.com/blog":  false,
			},
			[]*url.URL{mustParse(t, "http://example.com/blog")},
		},
		{
			index{
				"http://example.com":       false,
				"http://example.com/about": false,
				"http://example.com/blog":  false,
			},
			[]*url.URL{
				mustParse(t, "http://example.com"),
				mustParse(t, "http://example.com/about"),
				mustParse(t, "http://example.com/blog"),
			},
		},
	}

	for _, testCase := range testCases {
		unvisitedLinks := testCase.idx.getUnvisitedLinks()
		sort.SliceStable(unvisitedLinks, func(i, j int) bool {
			return unvisitedLinks[i].String() < unvisitedLinks[j].String()
		})
		sort.SliceStable(testCase.expected, func(i, j int) bool {
			return unvisitedLinks[i].String() < unvisitedLinks[j].String()
		})
		if !reflect.DeepEqual(unvisitedLinks, testCase.expected) {
			t.Fatalf("Expected %v. got %v", testCase.expected, unvisitedLinks)
		}
	}
}

func mustParse(t *testing.T, rawurl string) *url.URL {
	u, err := url.Parse(rawurl)
	if err != nil {
		t.Fatalf("Could not parse '%s': %v", rawurl, err)
	}
	return u
}

func TestGetWebpage(t *testing.T) {
	addr := fmt.Sprintf("http://localhost:%d", PORT)
	expectedOutput := `<head></head>
<body>
    <h1>My Heading<h1>
    <p>Some text</p>
    <a href="about.html">About Me</a>
    <a href="blog.html">Blog</a>
</body>
`
	killServer := startServer(t, PORT)
	defer killServer()
	u := mustParse(t, addr)
	body, _ := getWebpage(u)
	if string(body) != expectedOutput {
		t.Fatalf("Expected %s. got %s", expectedOutput, string(body))
	}
}
