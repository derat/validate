// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package validate

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// Text included in https://validator.w3.org/nu/ results pages on success.
const htmlSuccess = "The document validates according to the specified schema(s)."

// HTML reads an HTML document from r and validates it using https://validator.w3.org/nu/.
// Parsed issues and the raw HTML results page returned by the validation service are returned.
// If the returned error is non-nil, an issue occurred in the validation process.
func HTML(ctx context.Context, r io.Reader) ([]Issue, []byte, error) {
	resp, err := post(ctx, "https://validator.w3.org/nu/",
		map[string]string{"action": "check"},
		[]fileInfo{fileInfo{field: "uploaded_file", name: "data", ctype: string(HTMLDoc), r: r}})
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
	issues := extractHTMLIssues(node)
	err = checkResponse(strings.Contains(string(out), htmlSuccess), issues)
	return issues, out, nil
}

// extractHTMLIssues recursively walks n and returns validation issues.
// n is all or part of a document returned by https://validator.w3.org/nu/,
// where errors are denoted by <li class="error">.
func extractHTMLIssues(n *html.Node) []Issue {
	// TODO: Does the validator return warnings?
	if n.Type == html.ElementNode && n.Data == "li" && getAttr(n, "class") == "error" {
		return []Issue{makeHTMLIssue(n, Error)}
	}

	var issues []Issue
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		issues = append(issues, extractHTMLIssues(c)...)
	}
	return issues
}

// makeHTMLIssue creates a new issue by examining the supplied <li class="error"> node.
//
// Here's an example error, with line breaks and whitespace added for legibility:
//
//   <li class="error">
//     <p>    <strong>Error</strong>: <span>Saw <code>&lt;&gt;</code>. Probable causes:
//       Unescaped <code>&lt;</code> (escape as <code>&amp;lt;</code>) or mistyped
//       start tag.</span>
//     </p>
//     <p class="location">
//       <a href="#cl6c14">At line <span class="last-line">6</span>, column
//       <span class="last-col">14</span></a>
//     </p>
//     <p class="extract">    <code>&gt;<span class="lf" title="Line break">↩</span>ueaueohtn
//       u&gt;&lt;<b>&gt;</b>&lt;&gt; Y<span class="lf" title="Line
//       break">↩</span>&lt;body&gt;<span class="lf" title="Line
//       break">↩</span>&lt;p</code>
//     </p>
//   </li>
func makeHTMLIssue(li *html.Node, sev Severity) Issue {
	is := Issue{Severity: sev}

	for n := li.FirstChild; n != nil; n = n.NextSibling {
		if n.Type != html.ElementNode || n.Data != "p" {
			continue
		}
		switch getAttr(n, "class") {
		case "location":
			lstr := getText(n, func(n *html.Node) bool {
				return n.Type == html.ElementNode && n.Data == "span" && getAttr(n, "class") == "last-line"
			})
			is.Line, _ = strconv.Atoi(strings.TrimSpace(lstr))

			cstr := getText(n, func(n *html.Node) bool {
				return n.Type == html.ElementNode && n.Data == "span" && getAttr(n, "class") == "last-col"
			})
			is.Col, _ = strconv.Atoi(strings.TrimSpace(cstr))
		case "extract":
			is.Context = strings.TrimSpace(getText(n, nil))
		case "":
			msg := strings.TrimSpace(getText(n, func(n *html.Node) bool {
				return n.Type == html.ElementNode && n.Data == "span"
			}))
			is.Message = spacesAroundLines.ReplaceAllString(msg, "\n")
		}
	}

	return is
}

var spacesAroundLines = regexp.MustCompile(`\s*\n\s*`)
