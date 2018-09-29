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

func main() {
	uri := flag.String("uri", "http://localhost:8080", "the uri to scrape links")
	limit := flag.Int("limit", 10, "the maximum number of links to traverse")
	flag.Parse()

	m := traverse(*uri, *limit)
	for k, v := range m {
		log.Println(k, v)
	}
}

func fetch(href string) ([]byte, error) {
	if href == "" {
		return nil, errors.New("cannot be empty")
	}
	_, err := url.Parse(href)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(href)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func parser(root url.URL, r io.Reader) (result []string) {
	doc, err := html.Parse(r)
	if err != nil {
		log.Println(err)
		return nil
	}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					href := a.Val
					// Most likely a relative path.
					if !strings.HasPrefix(href, "http") {
						root.Path = href
						href = root.String()
					}
					parsed, err := url.PathUnescape(href)
					if err != nil {
						log.Println(err)
						continue
					}
					result = append(result, parsed)
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

func traverse(rootURL string, limit int) map[string]int {
	baseURL, err := url.Parse(rootURL)
	if err != nil {
		log.Fatal(err)
	}
	children := []string{rootURL}
	cache := make(map[string]int)
	for (len(children)) > 0 && len(cache) < limit {
		item := children[0]
		children = children[1:]
		if _, found := cache[item]; found {
			cache[item]++
			continue
		}
		cache[item]++

		body, err := fetch(item)
		if err != nil {
			log.Println(err)
			continue
		}
		links := parser(*baseURL, bytes.NewBuffer(body))
		children = append(children, links...)
	}
	return cache
}
