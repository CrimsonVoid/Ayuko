package url

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"code.google.com/p/go.net/html"
	"github.com/crimsonvoid/irclib/styles"
)

func Parse(url string) (string, error) {
	for _, parser := range parseMap {
		if !parser.re.MatchString(url) {
			continue
		}

		out, err := parser.fn(parser.re, url)
		if err == nil {
			return out, err
		}

		Module.Logger.Errorf("Parse(%v) %v", url, err)

		break
	}

	// return genericParser(url)
	return "", errors.New("No match")
}

func ytVidParser(re *regexp.Regexp, uri string) (string, error) {
	groups, err := matchGroups(re, uri)
	if err != nil {
		return "", err
	}

	jData := ytVidJSON{}
	if err = decodeJSON(fmt.Sprintf(ytVidAPI, groups["id"]), &jData); err != nil {
		return "", err
	}
	data := jData.Data

	timeQuery := ""
	if u, err := url.Parse(uri); err == nil {
		vals := u.Query()

		if t := vals.Get("t"); t != "" {
			timeQuery = "?t=" + t
		}
	}

	duration, err := time.ParseDuration(fmt.Sprintf("%ds", data.Duration))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("[https://youtu.be/%v%v] %v - %v (%v)",
		data.Id, timeQuery,
		data.Uploader,
		styles.Bold.Paint("%v", data.Title),
		duration), nil
}

func ytPLParser(re *regexp.Regexp, url string) (string, error) {
	groups, err := matchGroups(re, url)
	if err != nil {
		return "", err
	}

	jData := ytPLJSON{}
	if err = decodeJSON(fmt.Sprintf(ytPLAPI, groups["id"]), &jData); err != nil {
		return "", err
	}
	data := jData.Data

	return fmt.Sprintf("[https://youtube.com/playlist?list=%v] %v - %v (%v videos)",
		data.Id,
		data.Author,
		styles.Bold.Paint("%v", data.Title),
		data.TotalItems), nil
}

func githubParser(re *regexp.Regexp, url string) (string, error) {
	groups, err := matchGroups(re, url)
	if err != nil {
		return "", err
	}

	api := fmt.Sprintf(githubAPI, groups["user"], groups["repo"])

	jData := githubJSON{}

	// Get basic repo info
	if err := decodeJSON(api, &jData); err != nil {
		return "", err
	}

	// Get issue info
	if id := groups["id"]; id != "" {
		api += fmt.Sprintf(gitIssueAppend, id)
		if err := decodeJSON(api, &jData); err != nil {
			return "", err
		}
	}

	var lang string
	switch jData.State { // `State` will only be present for issues
	case "open":
		lang = styles.LightGreen.Fg("%v", jData.Language)
	case "closed":
		lang = styles.LightRed.Fg("%v", jData.Language)
	default: // case "":
		lang = styles.LightBlue.Fg("%v", jData.Language)
	}

	desc, homepage := jData.Title, ""
	if desc == "" {
		desc, homepage = jData.Description, jData.Homepage
	}

	if len(desc) > maxContentLen {
		desc = desc[:maxContentLen] + "..."
	}

	return fmt.Sprintf("[%v] <%v> %v %v",
		jData.Html_url,
		lang,
		desc,
		homepage,
	), nil
}

func vimeoParser(re *regexp.Regexp, url string) (string, error) {
	groups, err := matchGroups(re, url)
	if err != nil {
		return "", err
	}

	jData := make([]vimeoJSON, 0, 1)
	if err = decodeJSON(fmt.Sprintf(vimeoAPI, groups["id"]), &jData); err != nil {
		return "", err
	}
	if len(jData) == 0 {
		return "", fmt.Errorf("No data found")
	}

	data := jData[0]

	duration, err := time.ParseDuration(fmt.Sprintf("%ds", data.Duration))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("[http://vimeo.com/%v] %v - %v (%v)",
		data.Id,
		data.Username,
		styles.Bold.Paint("%v", data.Title),
		duration), nil
}

func steamParser(re *regexp.Regexp, url string) (string, error) {
	groups, err := matchGroups(re, url)
	if err != nil {
		return "", err
	}

	return parseTitle(fmt.Sprintf("http://store.steampowered.com/app/%v",
		groups["id"]))
}

func hnParser(re *regexp.Regexp, url string) (string, error) {
	groups, err := matchGroups(re, url)
	if err != nil {
		return "", err
	}

	return parseTitle(fmt.Sprintf("https://news.ycombinator.com/item?id=%v",
		groups["id"]))
}

func parseTitle(url string) (string, error) {
	title, err := genericParser(url)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("[%v] %v",
		url,
		styles.Bold.Paint("%v", title),
	), nil
}

func decodeJSON(url string, data interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := respOkay(resp); err != nil {
		return err
	}

	dec := json.NewDecoder(resp.Body)
	if err = dec.Decode(data); err != nil {
		return err
	}

	return nil
}

// Returns the <title> of `url`
func genericParser(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if err := respOkay(resp); err != nil {
		return "", err
	}

	// Tokenizer
	tkn, err := html.Parse(resp.Body)
	if err != nil {
		return "", err
	}

	return findTitle(tkn)
}

func findTitle(n *html.Node) (string, error) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		// Get title from <head> element
		if c.Data == "head" {
			return getTitle(c)
		}

		// Parse child nodes only if they are html.(Document|Element)Nodes
		if fChild := c.FirstChild; fChild != nil &&
			(fChild.Type == html.DocumentNode || fChild.Type == html.ElementNode) {
			return findTitle(c)
		}
	}

	return "", errors.New("head attribute not found")
}

// Find "<title>" tag, returning an error if the tag is not found
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

// Check the response is okay and "content-type" is plain text
func respOkay(resp *http.Response) error {
	// 200 < resp <= 400
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("Response status %v", resp.Status)
	}

	// 'Content-Type' response header contains (text|json)
	contentType := make([]string, 0, 1)

	for key, val := range resp.Header {
		if strings.ToLower(key) == "content-type" {
			contentType = append(contentType, val...)
			break
		}
	}

	// Return true if (text|json)
	for _, cntType := range contentType {
		if strings.Contains(cntType, "text") || strings.Contains(cntType, "json") {
			return nil
		}
	}

	return errors.New("Content-Type not text|json")
}
