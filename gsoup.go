package gsoup

import (
	"golang.org/x/net/html"
	"net/http"
	"fmt"
	"strings"
	"log"
	"regexp"
	"errors"
	"io/ioutil"
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
	//foundUrls := make(map[string]bool)
	//seedUrls := os.Args[1:]
	//
	//// Channels to hold the concurrent requests
	//chUrls := make(chan string)
	//chFinished := make(chan bool)
	//
	//// Kick off the crawl process (concurrently)
	//for _, url := range seedUrls {
	//	go listLinks(url, chUrls, chFinished)
	//}
	//
	//// Subscribe to both channels
	//for c := 0; c < len(seedUrls); {
	//	select {
	//	case url := <-chUrls:
	//		foundUrls[url] = true
	//	case <-chFinished:
	//		c++
	//	}
	//}
	//
	//fmt.Sprintf("\nUnique urls found are : %d\n", len(foundUrls))
	//for url := range foundUrls {
	//	fmt.Println(" - " + url)
	//}
	//
	//close(chUrls)

	resp, _ := connect("https://github.com/saopayne/gsoup")
	doc := HTMLParse(resp)
	title := doc.Find("body").Text()
	fmt.Println("Title of the Readme :", title)
	comicImg := doc.Find("h1").Find("p")
	fmt.Println("Description of the project :", comicImg.Attrs()["src"])
}

// Using depth first search to find the first occurrence and return
func FindOnce(n *html.Node, args []string, uni bool) (*html.Node, bool) {
	if uni == true {
		if n.Type == html.ElementNode && n.Data == args[0] {
			if len(args) > 1 && len(args) < 4 {
				for i := 0; i < len(n.Attr); i++ {
					if n.Attr[i].Key == args[1] && n.Attr[i].Val == args[2] {
						return n, true
					}
				}
			} else if len(args) == 1 {
				return n, true
			}
		}
	}
	uni = true
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		p, q := FindOnce(c, args, true)
		if q != false {
			return p, q
		}
	}
	return nil, false
}

// Using depth first search to find all occurrences and return
func FindAllofem(n *html.Node, args []string) []*html.Node {
	var nodeLinks = make([]*html.Node, 0, 10)
	var f func(*html.Node, []string, bool)
	f = func(n *html.Node, args []string, uni bool) {
		if uni == true {
			if n.Data == args[0] {
				if len(args) > 1 && len(args) < 4 {
					for i := 0; i < len(n.Attr); i++ {
						if n.Attr[i].Key == args[1] && n.Attr[i].Val == args[2] {
							nodeLinks = append(nodeLinks, n)
						}
					}
				} else if len(args) == 1 {
					nodeLinks = append(nodeLinks, n)
				}
			}
		}
		uni = true
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c, args, true)
		}
	}
	f(n, args, false)
	return nodeLinks
}

// Returns a key pair value (like a dictionary) for each attribute
func GetKeyValue(attributes []html.Attribute) map[string]string {
	var keyvalues = make(map[string]string)
	for i := 0; i < len(attributes); i++ {
		_, exists := keyvalues[attributes[i].Key]
		if exists == false {
			keyvalues[attributes[i].Key] = attributes[i].Val
		}
	}
	return keyvalues
}

// Catch panics when they occur
func localPanic(fnName string) {
	if r := recover(); r != nil {
		log.Println("Error occurred in", fnName, ":", r)
	}
}

// Root is a structure containing a pointer to an html node, the node value, and an error variable to return an error if occurred
type Root struct {
	Pointer   *html.Node
	NodeValue string
	Error     error
}

var debug = false

func EnableDebug() {
	debug = true
}

func DisableDebug() {
	debug = false
}

// Get returns the HTML returned by the url in string
// Accepts the url and gets the content of the url page
// The connect(String url) method creates a new Connection, and get() fetches and parses a HTML file.
// If an error occurs whilst fetching the URL, it will throw an exception, which you should handle appropriately.
func connect(url string) (string, error) {
	defer localPanic("connect()")
	resp, err := http.Get(url)
	if err != nil {
		if debug {
			panic("Couldn't perform GET request to " + url)
		}

		return "", errors.New("Couldn't perform GET request to " + url)
	}

	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		if debug {
			panic("Unable to read the response body")
		}

		return "", errors.New("Unable to read the response body")
	}

	return string(bytes), nil
}

func HtmlToPlainText(s string){}

// HTMLParse parses the HTML returning a start pointer to the DOM
func HTMLParse(s string) Root {
	defer localPanic("HTMLParse()")
	r, err := html.Parse(strings.NewReader(s))
	if err != nil {
		if debug {
			panic("Unable to parse the provided HTML")
		}

		return Root{nil, "", errors.New("Unable to parse the HTML")}
	}

	// Navigate to find an html.ElementNode
	for r.Type != html.ElementNode {
		switch r.Type {
		case html.DocumentNode:
			r = r.FirstChild
		case html.DoctypeNode:
			r = r.NextSibling
		case html.CommentNode:
			r = r.NextSibling
		}
	}
	return Root{r, r.Data, nil}
}

// Find finds the first occurrence of the given tag name,
// with or without attribute key and value specified,
// and returns a struct with a pointer to it
func (r Root) Find(args ...string) Root {
	defer localPanic("Find()")
	temp, ok := FindOnce(r.Pointer, args, false)
	if ok == false {
		if debug {
			panic("Element `" + args[0] + "` with attributes `" + strings.Join(args[1:], " ") + "` not found")
		}

		return Root{nil, "", errors.New("Element `" + args[0] + "` with attributes `" + strings.Join(args[1:], " ") + "` not found")}
	}

	return Root{temp, temp.Data, nil}
}

// FindAll finds all occurrences of the given tag name,
// with or without key and value specified,
// and returns an array of structs, each having
// the respective pointers
func (r Root) FindAll(args ...string) []Root {
	defer localPanic("FindAll()")
	temp := FindAllofem(r.Pointer, args)
	if len(temp) == 0 {
		if debug {
			panic("Element `" + args[0] + "` with attributes `" + strings.Join(args[1:], " ") + "` not found")
		}

		return []Root{}
	}

	pointers := make([]Root, 0, 10)
	for i := 0; i < len(temp); i++ {
		pointers = append(pointers, Root{temp[i], temp[i].Data, nil})
	}

	return pointers
}

// FindNextSibling finds the next sibling of the pointer in the DOM
// returning a struct with a pointer to it
func (r Root) FindNextSibling() Root {
	defer localPanic("FindNextSibling()")
	nextSibling := r.Pointer.NextSibling
	if nextSibling == nil {
		if debug {
			panic("No next sibling found")
		}

		return Root{nil, "", errors.New("No next sibling found")}
	}

	return Root{nextSibling, nextSibling.Data, nil}
}

// FindParent finds the parent of the pointer in the DOM
// returning a struct with a pointer to it
func (r Root) FindParent() Root {
	defer localPanic("FindNextSibling()")
	parent := r.Pointer.Parent
	if parent == nil {
		if debug {
			panic("No next sibling found")
		}

		return Root{nil, "", errors.New("No next sibling found")}
	}

	return Root{parent, parent.Data, nil}
}

// FindFirstChild finds the first child of the pointer in the DOM
// returning a struct with a pointer to it
func (r Root) FindFirstChild() Root {
	defer localPanic("FindNextSibling()")
	child := r.Pointer.FirstChild
	if child == nil {
		if debug {
			panic("No next sibling found")
		}

		return Root{nil, "", errors.New("No next sibling found")}
	}

	return Root{child, child.Data, nil}
}

// FindLastChild finds the last child of the pointer in the DOM
// returning a struct with a pointer to it
func (r Root) FindLastChild() Root {
	defer localPanic("FindNextSibling()")
	child := r.Pointer.LastChild
	if child == nil {
		if debug {
			panic("No next sibling found")
		}

		return Root{nil, "", errors.New("No next sibling found")}
	}

	return Root{child, child.Data, nil}
}

// FindPrevSibling finds the previous sibling of the pointer in the DOM
// returning a struct with a pointer to it
func (r Root) FindPrevSibling() Root {
	defer localPanic("FindPrevSibling()")
	prevSibling := r.Pointer.PrevSibling
	if prevSibling == nil {
		if debug {
			panic("No previous sibling found")
		}

		return Root{nil, "", errors.New("No previous sibling found")}
	}

	return Root{prevSibling, prevSibling.Data, nil}
}

// FindNextElementSibling finds the next element sibling of the pointer in the DOM
// returning a struct with a pointer to it
func (r Root) FindNextElementSibling() Root {
	defer localPanic("FindNextElementSibling()")
	nextSibling := r.Pointer.NextSibling
	if nextSibling == nil {
		if debug {
			panic("No next element sibling found")
		}

		return Root{nil, "", errors.New("No next element sibling found")}
	}

	if nextSibling.Type == html.ElementNode {
		return Root{nextSibling, nextSibling.Data, nil}
	}

	p := Root{nextSibling, nextSibling.Data, nil}
	return p.FindNextElementSibling()
}

// FindPrevElementSibling finds the previous element sibling of the pointer in the DOM
// returning a struct with a pointer to it
func (r Root) FindPrevElementSibling() Root {
	defer localPanic("FindPrevElementSibling()")
	prevSibling := r.Pointer.PrevSibling
	if prevSibling == nil {
		if debug {
			panic("No previous element sibling found")
		}

		return Root{nil, "", errors.New("No previous element sibling found")}
	}

	if prevSibling.Type == html.ElementNode {
		return Root{prevSibling, prevSibling.Data, nil}
	}

	p := Root{prevSibling, prevSibling.Data, nil}
	return p.FindPrevElementSibling()
}

// Attrs returns a map containing all attributes
func (r Root) Attrs() map[string]string {
	defer localPanic("Attrs()")
	if r.Pointer.Type != html.ElementNode {
		if debug {
			panic("Not an ElementNode")
		}

		return nil
	}

	if len(r.Pointer.Attr) == 0 {
		return nil
	}

	return GetKeyValue(r.Pointer.Attr)
}

// Text returns the string inside a non-nested element
func (r Root) Text() string {
	defer localPanic("Text()")
	k := r.Pointer.FirstChild
checkNode:
	if k.Type != html.TextNode {
		k = k.NextSibling
		if k == nil {
			if debug {
				panic("No text node found")
			}
			return ""
		}

		goto checkNode
	}

	if k != nil {
		r, _ := regexp.Compile(`^\s+$`)
		if ok := r.MatchString(k.Data); ok {
			k = k.NextSibling
			if k == nil {
				if debug {
					panic("No text node found")
				}

				return ""
			}

			goto checkNode
		}
		return k.Data
	}

	return ""
}