package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

type Link struct {
	Href    string
	Counter int
	Body    []byte
	Error   error
}

func main() {
	var website = flag.String("website", "http://localhost:8080", "the website to check")
	flag.Parse()

	baseURL, err := url.Parse(*website)
	if err != nil {
		log.Fatal(err)
	}

	cache := make(map[string]Link)

	var recurse func(Link)
	recurse = func(link Link) {
		var (
			href = link.Href
			body = link.Body
		)
		if c, found := cache[href]; found {
			c.Counter++
			return
		}
		cache[href] = link

		out := parser(*baseURL, bytes.NewBuffer(body), cache)
		for k, _ := range out {
			recurse(fetch(k))
		}
	}

	result := fetch(baseURL.String())
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	// cache[result.Href] = result
	recurse(result)
	for _, v := range cache {
		fmt.Printf("%s\t%d\t%v\n", v.Href, v.Counter, v.Error == nil)
	}
}

func fetch(href string) Link {
	_, err := url.Parse(href)
	if err != nil {
		return Link{Error: err}
	}

	resp, err := http.Get(href)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return Link{Error: err}
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Link{Error: err}
	}
	return Link{Body: body, Href: href}
}

func parser(root url.URL, r io.Reader, cache map[string]Link) map[string]int {
	// Parse the html.
	doc, err := html.Parse(r)
	if err != nil {
		return nil
	}
	links := make(map[string]int)
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" && len(a.Val) > 1 {
					href := a.Val
					if !strings.HasPrefix(a.Val, "http") {
						root.Path = a.Val
						href = root.String()
					}
					links[href]++
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return links
}
