package parser

import (
	"net/url"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestParse(t *testing.T) {
	rawInput := `
<!doctype html>
<head>
	<link rel="stylesheet" href="/css/example.css">
	<link rel="canonical" href="https://example.com">
</head>
<body>
	<a href="/about">About Me</a>
	<a href="https://example.com/blog"></a>
	<a href="https://google.com">Search</a>
</body>
`
	input := strings.NewReader(rawInput)
	expectedDetails := PageDetails{
		Assets: []string{
			"https://example.com/css/example.css",
			"https://example.com",
		},
		ExternalLinks: []string{
			"https://google.com",
		},
		InternalLinks: []string{
			"https://example.com/about",
			"https://example.com/blog",
		},
	}
	rawurl := "https://example.com"
	url, err := url.Parse(rawurl)
	if err != nil {
		t.Fatalf("Couldn't parse url '%s'", rawurl)
	}
	details := ParseWebpage(url, input)

	if !reflect.DeepEqual(details, expectedDetails) {
		t.Fatalf("Expected %v, got %v", expectedDetails, details)
	}
}

func TestGetAttribute(t *testing.T) {
	testCases := []struct {
		Attr          []html.Attribute
		Key           string
		ExpectedValue string
		ExpectedOk    bool
	}{
		{
			Attr: []html.Attribute{
				html.Attribute{
					Namespace: "", Key: "href", Val: "https://example.com",
				},
			},
			Key:           "href",
			ExpectedValue: "https://example.com",
			ExpectedOk:    true,
		},
		{
			Attr: []html.Attribute{
				html.Attribute{
					Namespace: "", Key: "foo", Val: "bar",
				},
			},
			Key:           "href",
			ExpectedValue: "",
			ExpectedOk:    false,
		},
	}

	for _, testCase := range testCases {
		value, ok := getAttribute(testCase.Attr, testCase.Key)
		if value != testCase.ExpectedValue {
			t.Fatalf("Expected %s. got %s", testCase.ExpectedValue, value)
		}
		if ok != testCase.ExpectedOk {
			t.Fatalf("Expected %t. got %t", testCase.ExpectedOk, ok)
		}
	}
}
