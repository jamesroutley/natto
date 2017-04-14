# Natto

Natto is a simple command-line web crawler. Given a starting URL, Natto will crawl that website and print a JSON map of that site's pages, listing each page's static assets and internal and external links.

## Example

```shell
$ natto http://tomblomfield.com
{
  "pages": {
    "http://tomblomfield.com": {
      "internal_links": [
        "http://tomblomfield.com/",
        "http://tomblomfield.com/about",
        "http://tomblomfield.com/archive",
        ...
      ],
      "external_links": [...],
      "assets": [...]
    },
    ...
  }
}
```

## API

```shell
Usage: natto [-concurrency] [-debug] [-no-indent] URL
  -concurrency int
    	Number of concurrent requests. (default 10)
  -debug
    	Enable debug logging.
  -no-indent
    	Print site map without indentation.
```

## Install

```shell
$ make install
```


## Implementation Notes

- Internal links defined as the value of `href` attributes of `a` tags with a scheme and host matching the start URL.
- External links defined as the value of `href` attributes of `a` tags with a scheme or host not matching the start URL.
- Assets defined as the value of `href` attributes of `link` tags.
- Natto assuemes ignores a URL's query and fragment when looking for distinct internal pages. For example, `http://example.com` and `http://example.com#section1` are considered to be the same. Single page applications which use the fragment to differentiate between pages will not be crawled correctly.
- Natto crawls multiple pages concurrently. The number of concurrent requests made can be changed with the flag `-concurrency`. Increasing concurrency can offer a significant speed boost, but puts more strain on the website so it may be best to limit this value.

## Concurrency benchmark

The following section shows the output of the `time` command for various concurrency values. Note that the absolute values will depend on internet speed, so only the relative times should be considered. For http://tomblomfield.com, the increase in speed slows down after ten concurrent requests. This number is specific to the particular website and cannot be generalised.

```
natto -concurrency=1 http://tomblomfield.com  0.22s user 0.09s system 1% cpu 30.792 total
natto -concurrency=2 http://tomblomfield.com  0.24s user 0.10s system 2% cpu 15.035 total
natto -concurrency=5 http://tomblomfield.com  0.25s user 0.11s system 4% cpu 7.405 total
natto -concurrency=10 http://tomblomfield.com  0.25s user 0.10s system 8% cpu 4.348 total
natto -concurrency=20 http://tomblomfield.com  0.26s user 0.11s system 9% cpu 4.024 total
natto -concurrency=50 http://tomblomfield.com  0.30s user 0.14s system 12% cpu 3.738 total
```

## Future work

- A plugin system for the Parser. It would be elegant to be able to define parser rules separately from the core parser code.
