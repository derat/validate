// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package validate

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// Text included in https://jigsaw.w3.org/css-validator/ results pages on success.
const cssSuccess = "<!-- NO ERRORS -->"

// CSS reads an HTML or CSS document from r and validates its CSS content using https://jigsaw.w3.org/css-validator/.
// FileType describes the type of file being validated: the W3C validator seems to have the unfortunate
// property of reporting that the data validated successfully if the wrong type is supplied.
//
// Parsed issues and the raw HTML results page returned by the validation service are returned.
// If the returned error is non-nil, an issue occurred in the validation process.
func CSS(ctx context.Context, r io.Reader, ft FileType) ([]Issue, []byte, error) {
	// TODO: Maybe make these form values configurable.
	// Available values can be seen in the source of https://jigsaw.w3.org/css-validator.
	resp, err := post(ctx, "https://jigsaw.w3.org/css-validator/validator",
		map[string]string{
			"profile":     "css3svg", // "none", "css1", "css2", "css21", "css3", "svg", etc.
			"usermedium":  "all",     // "screen", "print", etc.
			"warning":     "1",       // "no", "0" ("most important"), 1 ("normal report"), 2 ("all")
			"vextwarning": "",        // "" ("default"), "true" ("warnings"), "false" ("errors")
			"lang":        "en",
		},
		[]fileInfo{fileInfo{field: "file", name: "data", ctype: string(ft), r: r}})
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	node, err := html.Parse(bytes.NewReader(out))
	if err != nil {
		return nil, out, fmt.Errorf("failed to parse response: %v", err)
	}
	issues := extractCSSIssues(node)
	err = checkResponse(strings.Contains(string(out), cssSuccess), issues)
	return issues, out, err
}

// extractCSSIssues recursively walks n and returns validation issues.
// n is all or part of a document returned by https://jigsaw.w3.org/css-validator/.
func extractCSSIssues(n *html.Node) []Issue {
	if n.Type == html.ElementNode && n.Data == "tr" {
		switch getAttr(n, "class") {
		case "error":
			return []Issue{makeCSSIssue(n, Error)}
		case "warning":
			return []Issue{makeCSSIssue(n, Warning)}
		}
	}

	var issues []Issue
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		issues = append(issues, extractCSSIssues(c)...)
	}
	return issues
}

// makeCSSIssue creates a new issue by examining the supplied <tr class="error">
// or <tr class="warning"> node.
//
// Errors look like this:
//
//   <tr class="error">
//     <td class="linenumber" title="Line 17">17</td>
//     <td class="codeContext"> body </td>
//     <td class="parse-error">
//   </tr>
//
// Warnings look like this:
//
//   <tr class="warning">
//     <td class="linenumber" title="Line 15">15</td>
//     <td class="codeContext"></td>
//     <td class="level0" title="warning level 0"><code>-webkit-transform</code> is an unknown vendor extension</td>
//   </tr>
func makeCSSIssue(tr *html.Node, sev Severity) Issue {
	is := Issue{Severity: sev}
	for n := tr.FirstChild; n != nil; n = n.NextSibling {
		if n.Type != html.ElementNode || n.Data != "td" {
			continue
		}

		// Squish together all of the text content inside the <td>.
		text := strings.TrimSpace(getText(n, nil))
		if len(text) == 0 {
			continue
		}

		switch getAttr(n, "class") {
		case "linenumber":
			is.Line, _ = strconv.Atoi(text)
		case "codeContext":
			is.Context = text
		default:
			is.Message = text
		}
	}
	return is
}
