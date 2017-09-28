[![Build Status](https://travis-ci.org/saopayne/gsoup.svg?branch=master)](https://travis-ci.org/saopayne/gsoup)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](http://makeapullrequest.com)

# gsoup

A tiny web scraper written in Go with similar features to jsoup


## Getting started

```
go get github.com/saopayne/gsoup
```

## Initializing the client

## Usage

``` go

import (
    "https://github.com/saopayne/gsoup"
    "fmt"
)

// listing of links given a list of urls
// using goroutines and channels
func main() {
	foundUrls := make(map[string]bool)
	seedUrls := os.Args[1:]

	// Channels to hold the concurrent requests
	chUrls := make(chan string)
	chFinished := make(chan bool)

	// Kick off the crawl process (concurrently) using a goroutine
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

// accessing the DOM elements
func main() {
	resp, _ := gsoup.connect("")
	doc := gsoup.HTMLParse(resp)
	title := doc.Find("div", "id", "id_value").Text()
	image := doc.Find("div", "id", "imageid").Find("img")
	fmt.Println("Text linked to the image :", image.Attrs()["title"])
}

```


## TODO
- [ ] Write unit tests
- [ ] Documentation


## CONTRIBUTING
- Fork the repository, make necessary changes and send the PR.
