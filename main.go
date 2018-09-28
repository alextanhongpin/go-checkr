package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"

	"golang.org/x/net/html"
)

type Link struct {
	Href     string
	Verified bool
	Counter  int64
}

func main() {
	var mu sync.RWMutex
	var links map[string]*Link
	r := bytes.NewBuffer([]byte(`<a href="http://localhost:8080">car</a> <a href="car"></a>`))
	doc, err := html.Parse(r)
	if err != nil {
		log.Fatal(err)
	}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					log.Printf("key=%s value=%s", a.Key, a.Val)
					u, err := url.Parse(a.Val)
					if err != nil {
						log.Println("urlParseError:", err.Error())
						continue
					}
					log.Println("success:", u)

					body, err := fetchHTML(&mu, links, a.Val)
					if err != nil {
						log.Println("err:", err)
						continue
					}
					log.Println("got body:", body)
					// TODO: Scrape recursively.
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
}

func fetchHTML(mu *sync.RWMutex, links map[string]*Link, href string) ([]byte, error) {
	mu.Lock()
	defer mu.Unlock()
	l, found := links[href]
	if found {
		l.Counter++
		return nil, nil
	}

	// Perform fetch over the url.
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
	links[href] = &Link{
		Href:     href,
		Verified: false,
	}
	return body, nil
}
