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
	"os"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"

	"golang.org/x/net/html"
)

type Link struct {
	Href    string
	Counter int
	Body    []byte
	Error   error
}

const LIMIT = 10

func main() {
	var uri = flag.String("uri", "http://localhost:8080", "the uri to scrape for links")
	flag.Parse()

	cache := make(map[string]Link)

	var recurse func(url.URL, Link)

	recurse = func(baseURL url.URL, link Link) {
		if len(cache) > LIMIT {
			return
		}
		var (
			href = link.Href
			body = link.Body
		)
		if c, found := cache[href]; found {
			c.Counter++
			return
		}
		log.Println("found:", link.Href)
		cache[href] = link

		out := parser(baseURL, bytes.NewBuffer(body), cache)
		var wg sync.WaitGroup
		wg.Add(len(out))
		for k, _ := range out {
			go func(uri string) {
				defer wg.Done()
				recurse(baseURL, fetch(uri))
			}(k)
		}
		wg.Wait()
	}

	baseURL, err := url.Parse(*uri)
	if err != nil {
		log.Fatal(err)
	}
	recurse(*baseURL, increment(fetch(baseURL.String())))

	format := "%s\t%s\t%v\n"
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 4, 4, 4, ' ', tabwriter.Debug)
	fmt.Fprintf(w, format, "URL", "Frequency", "Success")
	fmt.Fprintf(w, format, "---", "---", "---")
	for _, v := range cache {
		fmt.Fprintf(w, format, v.Href, strconv.Itoa(v.Counter+1), v.Error == nil)
	}
	w.Flush()
}

func increment(link Link) Link {
	link.Counter++
	return link
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
	doc, err := html.Parse(r)
	if err != nil {
		log.Println(err)
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
