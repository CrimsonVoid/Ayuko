package url

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"code.google.com/p/go.net/html"
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

func ytVidParser(re *regexp.Regexp, url string) (string, error) {
	groups, err := matchGroups(re, url)
	if err != nil {
		return "", err
	}

	jData := ytVidJSON{}
	if err = getBody(fmt.Sprintf(ytVidAPI, groups["id"]), &jData); err != nil {
		return "", err
	}
	data := jData.Data

	duration, err := time.ParseDuration(fmt.Sprintf("%ds", data.Duration))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("[https://youtu.be/%v] %v - %v (%v)\n",
		data.Id, data.Uploader, data.Title, duration), nil
}

func ytPLParser(re *regexp.Regexp, url string) (string, error) {
	groups, err := matchGroups(re, url)
	if err != nil {
		return "", err
	}

	jData := ytPLJSON{}
	if err = getBody(fmt.Sprintf(ytPLAPI, groups["id"]), &jData); err != nil {
		return "", err
	}
	data := jData.Data

	return fmt.Sprintf("[https://youtube.com/playlist?list=%v] %v - %v (%v videos)\n",
		data.Id, data.Author, data.Title, data.TotalItems), nil
}

func githubParser(re *regexp.Regexp, url string) (string, error) {
	groups, err := matchGroups(re, url)
	if err != nil {
		return "", err
	}

	jData := githubJSON{}
	if err := getBody(fmt.Sprintf(githubAPI, groups["user"], groups["repo"]), &jData); err != nil {
		return "", err
	}

	descLen, elip := len(jData.Description), ""
	if descLen > maxContentLen {
		descLen, elip = maxContentLen, "..."
	}

	return fmt.Sprintf("[%v] <%v> %v%v %v\n",
		jData.Html_url, jData.Language, jData.Description[:descLen],
		elip, jData.Homepage), nil
}

func fourChParser(re *regexp.Regexp, url string) (string, error) {
	groups, err := matchGroups(re, url)
	if err != nil {
		return "", err
	}

	jData := fourChJSON{}
	if err = getBody(fmt.Sprintf(fourChAPI, groups["board"], groups["thread"]), &jData); err != nil {
		return "", err
	}

	op, post := jData.Posts[0], ""

	if postID, err := strconv.Atoi(groups["post"]); err == nil && postID > op.No {
		for _, fChData := range jData.Posts {
			if fChData.No != postID {
				continue
			}

			op = fChData
			post = fmt.Sprintf("#p%v", fChData.No)

			break
		}
	}

	subj, elip := op.Sub, ""

	if subj == "" {
		if subj = op.Com; subj == "" {
			subj = "{No comment}"
		}
	}

	if len(subj) > maxContentLen {
		subj = subj[:maxContentLen]
	}

	subj = html.UnescapeString(subj)

	return fmt.Sprintf("[https://boards.4chan.org/%v/res/%v%v] %v%v\n",
		groups["board"], jData.Posts[0].No, post, subj, elip), nil
}

func genericParser(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if err := respStatus(resp); err != nil {
		return "", err
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

func getBody(url string, data interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := respStatus(resp); err != nil {
		return err
	}

	if dec := json.NewDecoder(resp.Body); dec.Decode(data) != nil {
		return err
	}

	return nil
}

func respStatus(resp *http.Response) error {
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
