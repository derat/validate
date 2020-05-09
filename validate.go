// Package validate validates HTML and related documents.

package validate

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// FileType describes the type of file being validated.
type FileType string

const (
	// Stylesheet is a standalone CSS stylesheet.
	Stylesheet FileType = "text/css"
	// HTMLDoc is an HTML document.
	HTMLDoc = "text/html"
)

// Severity describes the severity of an issue.
type Severity int

const (
	// Error indicates an actual problem, e.g. an unclosed HTML tag or invalid CSS property.
	Error Severity = iota
	// Warning indicates a minor issue, e.g. a vendor-prefixed CSS property.
	Warning
)

func (s Severity) String() string {
	switch s {
	case Error:
		return "Error"
	case Warning:
		return "Warning"
	default:
		return ""
	}
}

// Issue describes a problem reported by a validator.
type Issue struct {
	// Severity describes the seriousness of the issue.
	Severity Severity
	// Line contains the 1-indexed line number where the issue occurred.
	// It is 0 if the line is unknown.
	Line int
	// Col contains the 0-indexed column number where the issue occurred.
	// It is 0 if the column is unknown.
	Col int
	// Message describes the issue.
	Message string
	// Context optionally provides more detail about where the issue occurred.
	Context string
}

func (is Issue) String() string {
	s := fmt.Sprintf("%d:%d %s: %s", is.Line, is.Col, is.Severity, is.Message)
	if is.Context != "" {
		s += " (" + is.Context + ")"
	}
	return s
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

// spaces matches one or more whitespace characters.
var spaces = regexp.MustCompile(`\s+`)

// getText recursively walks n and concatenates the contents of all nodes of type html.TextNode.
// Repeated spaces are compressed and <p> elements are converted to newlines.
// If f is non-nil, nodes will only be printed if f returns true for them or one of their ancestors.
func getText(n *html.Node, f func(*html.Node) bool) string {
	// If f isn't supplied or says that we should be included, ensure that all children are included.
	include := f == nil || f(n)
	if include {
		f = func(*html.Node) bool { return true }
	}

	if n.Type == html.TextNode {
		if !include {
			return ""
		}
		return spaces.ReplaceAllString(n.Data, " ")
	}

	var s string
	if include && n.Type == html.ElementNode && n.Data == "p" {
		s += "\n"
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		s += getText(c, f)
	}
	return s
}

// checkResponse returns an error if the validator's report of success
// and the issues parsed from its response disagree, i.e. success is true
// but there are errors in issues but success is false but no errors were found.
// This should help prevent reporting success falsely if/when the results format changes.
func checkResponse(success bool, issues []Issue) error {
	gotError := false
	for _, is := range issues {
		if is.Severity == Error {
			gotError = true
			break
		}
	}

	if !success && !gotError {
		return errors.New("got neither errors nor success message")
	} else if success && gotError {
		return errors.New("got both errors and success message")
	}
	return nil
}

// fileInfo describes a file to be uploaded by the post function.
type fileInfo struct {
	field string    // field name
	name  string    // filename
	ctype string    // content-type
	r     io.Reader // file data
}

// post executes a POST request to URL with the supplied fields
// and files sent as a multipart/form-data body.
func post(ctx context.Context, url string, fields map[string]string, files []fileInfo) (*http.Response, error) {
	// See https://stackoverflow.com/a/20397167.
	var b bytes.Buffer
	mw := multipart.NewWriter(&b)

	// Add non-file fields.
	for k, v := range fields {
		fw, err := mw.CreateFormField(k)
		if err != nil {
			return nil, err
		}
		if _, err := io.WriteString(fw, v); err != nil {
			return nil, err
		}
	}

	// Add files.
	for _, fi := range files {
		// This is directly lifted from CreateFormFile in Go's src/mime/multipart/writer.go.
		// We can't use that fuction because it hardcodes "application/octet-stream", while
		// the W3C's CSS validator appears to require the correct MIME type for the file.
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition",
			fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
				escapeQuotes(fi.field), escapeQuotes(fi.name)))
		h.Set("Content-Type", fi.ctype)
		fw, err := mw.CreatePart(h)
		if err != nil {
			return nil, err
		}
		if _, err := io.Copy(fw, fi.r); err != nil {
			return nil, err
		}
	}

	// Finish the message.
	if err := mw.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, &b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	return http.DefaultClient.Do(req)
}

// From Go's src/mime/multipart/writer.go.
var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}
