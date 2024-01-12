package main

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

// Parses an HTML document and extracts the words and href attributes
func extract(url string, body []byte, resultChan chan<- ExtractResult) {
	var extractRes ExtractResult
	var skipTags = map[string]bool{
		"style":  true,
		"script": true,
	}
	// Parse the HTML document
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		fmt.Println("Error parsing body")
	}
	var f func(*html.Node)
	f = func(n *html.Node) {
		switch n.Type {
		//If the node is an HTML element node.
		case html.ElementNode:
			if skipTags[n.Data] {
				return
			}
			// If the element is an anchor tag then extract its href attribute
			if n.Data == "a" {
				for _, a := range n.Attr {
					if a.Key == "href" {
						extractRes.hrefs = append(extractRes.hrefs, a.Val)
						break
					}
				}
			}

			// If the element is an img tag then extract its alt and src attributes
			if n.Data == "img" {
				var alt, src string
				for _, a := range n.Attr {
					switch a.Key {
					case "alt":
						alt = a.Val
						extractRes.altWords = append(extractRes.altWords, strings.FieldsFunc(a.Val, func(r rune) bool {
							return !unicode.IsLetter(r) && !unicode.IsNumber(r)
						})...)
					case "src":
						src = a.Val
					}
				}
				extractRes.Images = append(extractRes.Images, alt)
				extractRes.ImageSrcs = append(extractRes.ImageSrcs, src)
			}

			if n.Data == "title" {
				extractRes.title = strings.TrimSpace(titleNodeText(n))
			}
		// If the node is a text node
		case html.TextNode:
			extractRes.words = append(extractRes.words, strings.FieldsFunc(n.Data, func(r rune) bool {
				return !unicode.IsLetter(r) && !unicode.IsNumber(r)
			})...)
		}
		// Recursion to process the child nodes
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)
	resultChan <- ExtractResult{url, extractRes.words, extractRes.hrefs, extractRes.title, extractRes.Images, extractRes.ImageSrcs, extractRes.altWords}
}

// Extract out the title tag from the html content
func titleNodeText(n *html.Node) string {
	var text string

	if n.Type == html.TextNode {
		text = n.Data
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text += titleNodeText(c)
	}

	return text
}
