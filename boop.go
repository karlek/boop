// Package xfinger implements functions for retrieving image URLs of mugshots
// of KTH people.
package xfinger

import (
	"fmt"
	"net/url"

	"code.google.com/p/go.net/html"
	"github.com/mewkiz/pkg/errutil"
	"github.com/mewkiz/pkg/httputil"
)

const (
	xfingerFmt = "http://www.csc.kth.se/hacks/new/xfinger/results.php?freetext=%s"
	xfingerImg = "http://www.csc.kth.se/hacks/new/xfinger/"
)

// Lookup takes a name and returns the potential images of that person.
func Lookup(name string) ([]*url.URL, error) {
	// Retrieve the HTML DOM tree.
	doc, err := httputil.GetDoc(fmt.Sprintf(xfingerFmt, name))
	if err != nil {
		return nil, errutil.Err(err)
	}

	// Find all image elements and append them to the imgUrls slice.
	var imgUrls []*url.URL

	// f is the recursive function to traverse the DOM tree.
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			for _, attr := range n.Attr {
				if attr.Key != "src" {
					continue
				}
				u, err := url.Parse(xfingerImg + attr.Val)
				if err != nil {
					continue
				}
				imgUrls = append(imgUrls, u)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return imgUrls, nil
}

// Single takes a name and returns the potential images of that person.
func Single(name string) (*url.URL, error) {
	urls, err := Lookup(name)
	if err != nil {
		return nil, errutil.Err(err)
	}
	if urls == nil {
		return nil, errutil.NewNoPosf("no match found for: %s.", name)
	}
	return urls[0], nil
}
