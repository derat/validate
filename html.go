// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package validate

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// Text included in https://validator.w3.org/nu/ results pages on success.
const htmlSuccess = "The document validates according to the specified schema(s)."

// HTML reads an HTML document from r and validates it using https://validator.w3.org/nu/.
// issues describes individual issues in the page. Each entry is typically a multiline string.
// out contains the raw HTML page returned by the validation service.
// If err is non-nil, an issue occurred in the validation process.
func HTML(ctx context.Context, r io.Reader) (issues []string, out []byte, err error) {
	body, err := post(ctx, "https://validator.w3.org/nu/",
		map[string]string{"action": "check"},
		[]fileInfo{{field: "uploaded_file", name: "page.html", r: r}})
	if err != nil {
		return nil, nil, err
	}
	defer body.Close()

	out, err = ioutil.ReadAll(body)
	if err != nil {
		return nil, nil, err
	}
	node, err := html.Parse(bytes.NewReader(out))
	if err != nil {
		return nil, out, fmt.Errorf("failed to parse response: %v", err)
	}
	issues = extractHTMLIssues(node)

	// To avoid reporting success falsely if/when the results page changes,
	// check that that we found either issues or the success message.
	if len(issues) == 0 {
		if !strings.Contains(string(out), htmlSuccess) {
			return nil, out, errors.New("didn't find any issues or success message")
		}
	} else {
		if strings.Contains(string(out), htmlSuccess) {
			return issues, out, errors.New("found both issuess and success message")
		}
	}
	return issues, out, nil
}

var spaces = regexp.MustCompile(`\s+`)
var spacesAroundLines = regexp.MustCompile(`\s*\n\s*`)

// extractHTMLIssues recursively walks n and returns an array of slices describing validation issues.
// n is all or part of a document returned by https://validator.w3.org/nu/, where errors are denoted
// by <li class="error">.
func extractHTMLIssues(n *html.Node) []string {
	if n.Type == html.ElementNode && n.Data == "li" && getAttr(n, "class") == "error" {
		msg := strings.TrimSpace(getText(n))
		msg = spacesAroundLines.ReplaceAllString(msg, "\n")
		return []string{msg}
	}

	var errors []string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		for _, e := range extractHTMLIssues(c) {
			errors = append(errors, e)
		}
	}
	return errors
}

// getAttr returns the first named attribute from n, or an empty string if the attribute doesn't exist.
func getAttr(n *html.Node, name string) string {
	for _, a := range n.Attr {
		if a.Key == name {
			return a.Val
		}
	}
	return ""
}

// getText recursively walks n and concatenates the contents of all nodes of type html.TextNode.
// Repeated spaces are compressed and <p> elements are converted to newlines.
func getText(n *html.Node) string {
	if n.Type == html.TextNode {
		return spaces.ReplaceAllString(n.Data, " ")
	}

	var s string
	if n.Type == html.ElementNode && n.Data == "p" {
		s += "\n"
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		s += getText(c)
	}
	return s
}
