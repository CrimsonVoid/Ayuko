package url

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"code.google.com/p/go.net/html"
)

func ParseTitle(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check if 200 <= resp < 400
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return "", fmt.Errorf("Response status %v", resp.Status)
	}

	// 'Content-Type' response header is 'text/html'
	contentType := make([]string, 0, 1)
	htmlResp := false

	for key, val := range resp.Header {
		if strings.ToLower(key) == "content-type" {
			contentType = append(contentType, val...)
			break
		}
	}

	for _, typ := range contentType {
		if strings.Contains(typ, "text/html") {
			htmlResp = true
			break
		}
	}

	if !htmlResp {
		return "", errors.New("Content-Type not text/html")
	}

	// Tokenizer
	tkn, err := html.Parse(resp.Body)
	if err != nil {
		return "", err
	}

	return parseTitle(tkn)
}

func parseTitle(n *html.Node) (string, error) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		// Get title from <head> element
		if c.Data == "head" {
			return getTitle(c)
		}

		// Parse child nodes only if they are html.(Document|Element)Nodes
		if fChild := c.FirstChild; fChild != nil &&
			(fChild.Type == html.DocumentNode || fChild.Type == html.ElementNode) {
			return parseTitle(c)
		}
	}

	return "", errors.New("head attribute not found")
}

func getTitle(n *html.Node) (string, error) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		// Only parse (Document|Element)Node's
		// continue if not <title> tag
		if (c.Type != html.DocumentNode && c.Type != html.ElementNode) || c.Data != "title" {
			continue
		}

		if child := c.FirstChild; child != nil && child.Type == html.TextNode {
			return child.Data, nil
		}

		// break if first <title> tag not html.TextNode
		break
	}

	return "", errors.New("No title attribute")
}
