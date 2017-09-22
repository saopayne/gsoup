package main

import (
	"golang.org/x/net/html"
	"net/http"
	"fmt"
	"strings"
	"os"
)

// Function to pull the href attribute from an anchor token from the html tokenizer
func getHrefFromAnchorTag(t html.Token) (ok bool, href string) {
	// Iterate over all of the Token's attributes until we find an "href"
	for _, a := range t.Attr {
		if a.Key == "href" {
			href = a.Val
			ok = true
		}
	}
	// "bare" return will return the variables (ok, href) as defined in
	// the function definition
	return
}

// Extract all http links from a given web page
// ->passing a list of urls to it via the channel causes the function to
// ->list all the links on each page for each url item
// A link is contained within <a href=""> </a>, we can select any tag with that
//
func listLinks(url string, ch chan string, chanExhausted chan bool) {

	resp, err := http.Get(url)

	// since defer gets executed last regardless
	// notify that the links listing has finished
	defer func() {
		chanExhausted <- true
	}()

	if err != nil {
		fmt.Sprintf("ERROR: Failed to get the links for the url: {\"%s\"}", url)
		return
	}

	b := resp.Body

	defer b.Close()

	z := html.NewTokenizer(b)

	for {
		nextToken := z.Next()

		switch {
		case nextToken == html.ErrorToken:
			// End of the document
			return
		case nextToken == html.StartTagToken:
			// example <a> <p> <span>
			t := z.Token()
			// Check if the token is an <a> tag
			isAnchor := t.Data == "a"
			if !isAnchor {
				continue
			}
			// Extract the href value from , if there is one
			ok, url := getHrefFromAnchorTag(t)
			if !ok {
				continue
			}
			// Ensure the url starts with http**
			hasProto := strings.Index(url, "http") == 0
			if hasProto {
				ch <- url
			}
		}
	}
}

func main() {
	foundUrls := make(map[string]bool)
	seedUrls := os.Args[1:]

	// Channels to hold the concurrent requests
	chUrls := make(chan string)
	chFinished := make(chan bool)

	// Kick off the crawl process (concurrently)
	for _, url := range seedUrls {
		go listLinks(url, chUrls, chFinished)
	}

	// Subscribe to both channels
	for c := 0; c < len(seedUrls); {
		select {
		case url := <-chUrls:
			foundUrls[url] = true
		case <-chFinished:
			c++
		}
	}

	fmt.Sprintf("\nUnique urls found are : %d\n", len(foundUrls))
	for url := range foundUrls {
		fmt.Println(" - " + url)
	}

	close(chUrls)
}