package main

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

type Status struct {
	Count int
	Error error
	Code  int
}

func main() {
	var (
		uri   = flag.String("uri", "http://localhost:8080", "the uri to scrape links")
		limit = flag.Int("limit", 10, "the maximum number of links to traverse")
	)
	flag.Parse()

	m := traverse(*uri, *limit)
	if len(m) == 0 {
		log.Println("no results")
		return
	}
	for k, v := range m {
		log.Println(k, v.Code, v.Error, v.Code)
	}
}

func fetch(href string) ([]byte, int, error) {
	code := -1
	if href == "" {
		return nil, code, errors.New("cannot be empty")
	}
	_, err := url.Parse(href)
	if err != nil {
		return nil, code, err
	}
	resp, err := http.Get(href)
	if resp != nil {
		code = resp.StatusCode
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, code, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, code, err
	}
	return body, code, nil
}

func parser(r io.Reader) (result []string) {
	doc, err := html.Parse(r)
	if err != nil || doc == nil {
		log.Println(err)
		return nil
	}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n == nil {
			return
		}
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" && len(a.Val) > 0 {
					result = append(result, a.Val)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return
}

func traverse(rootURL string, limit int) map[string]*Status {
	baseURL, err := url.Parse(rootURL)
	if err != nil {
		log.Fatal(err)
	}
	children := []string{rootURL}
	cache := make(map[string]*Status)
	for (len(children)) > 0 && len(cache) < limit {
		item := children[0]
		children = children[1:]
		if c, found := cache[item]; found {
			c.Count++
			continue
		}
		log.Println("fetching", item)
		body, status, err := fetch(item)
		cache[item] = &Status{Count: 1, Error: err, Code: status}
		if err != nil || body == nil || status != 200 {
			continue
		}
		links := parser(bytes.NewBuffer(body))

		if len(links) == 0 {
			continue
		}

		for k, v := range links {
			absURL := mapRelativeURL(*baseURL, v)
			parsed, err := url.PathUnescape(absURL)
			if err != nil {
				log.Println(err)
				continue
			}
			links[k] = parsed
		}
		children = append(children, links...)
	}
	return cache
}

func mapRelativeURL(root url.URL, path string) string {
	// Most likely a relative path.
	if !strings.HasPrefix(path, "http") {
		root.Path = path
		return root.String()

	}
	return path
}
