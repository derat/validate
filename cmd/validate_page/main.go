// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/derat/validate"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTION] <FILE>\n"+
			"Validate an HTML or CSS document.\n"+
			"If <FILE> isn't supplied, reads from stdin.\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	browser := flag.Bool("browser", false,
		"Display validation issues in browser (printed to stdout otherwise)")
	fileType := flag.String("type", "",
		`File type: "amp", "css", "html", "htmlcss" (validate CSS in HTML); inferred if empty`)
	flag.Parse()

	var r io.Reader
	var p string // file path; empty for stdin
	switch len(flag.Args()) {
	case 0:
		r = os.Stdin
	case 1:
		p = flag.Arg(0)
		if *fileType == "" {
			if strings.HasSuffix(p, ".amp") || strings.HasSuffix(p, ".amp.html") {
				*fileType = "amp"
			} else if strings.HasSuffix(p, ".css") {
				*fileType = "css"
			} else if strings.HasSuffix(p, ".html") || strings.HasSuffix(p, ".htm") {
				*fileType = "html"
			}
		}
		f, err := os.Open(p)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to open input file:", err)
			os.Exit(1)
		}
		defer f.Close()
		r = f
	default:
		flag.Usage()
		os.Exit(2)
	}

	if *fileType == "" {
		br := bufio.NewReader(r)
		r = br
		b, err := br.Peek(512)
		if err != nil && err != io.EOF {
			fmt.Fprintln(os.Stderr, "Failed to read file to infer type:", err)
			os.Exit(1)
		}
		ctype := http.DetectContentType(b)
		switch {
		case strings.HasPrefix(ctype, "text/html"):
			lower := strings.ToLower(string(b))
			if strings.Contains(lower, "<html amp>") || strings.Contains(lower, "<html ⚡>") {
				*fileType = "amp"
			} else {
				*fileType = "html"
			}
		case strings.HasPrefix(ctype, "text/plain"): // all we get for stylesheets :-/
			*fileType = "css"
		default:
			fmt.Fprintf(os.Stderr, "Inferred unsupported file type %q; pass -type\n", ctype)
			os.Exit(1)
		}
	}

	var issues []validate.Issue
	var out []byte
	var err error

	switch *fileType {
	case "amp":
		// amphtml-validator doesn't generate a results page, so make our own.
		issues, err = validate.AMP(context.Background(), r)
		if err == nil && *browser {
			out, err = makeAMPResultsPage(issues)
		}
	case "css":
		issues, out, err = validate.CSS(context.Background(), r, validate.Stylesheet)
	case "html":
		issues, out, err = validate.HTML(context.Background(), r)
	case "htmlcss":
		issues, out, err = validate.CSS(context.Background(), r, validate.HTMLDoc)
	default:
		fmt.Fprintf(os.Stderr, "Bad -type value %q\n", *fileType)
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Validation request failed:", err)
		os.Exit(1)
	}

	if *browser {
		if err := validate.LaunchBrowser(out); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to display results in browser:", err)
			os.Exit(1)
		}
	} else {
		for _, is := range issues {
			fmt.Println(is)
		}
	}
}

// guessType attempts to infer the MIME type of the data in r,
// which must be positioned at the beginning of the file.
func guessType(r bufio.Reader) (string, error) {
	b, err := r.Peek(512)
	if err != nil {
		return "", err
	}
	return http.DetectContentType(b), nil
}

// makeAMPResultsPage generates a minimal HTML page listing the supplied issues.
func makeAMPResultsPage(issues []validate.Issue) ([]byte, error) {
	tmpl := template.Must(template.New("").Parse(`<DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <title>AMP validation results</title>
  </head>
  <body>
{{- if not .}}
    No issues found.
{{- else -}}
{{range .}}
    {{.Line}}:{{.Col}} {{.Severity}} {{.Message}} {{if .URL}}<a href="{{.URL}}">{{end}}{{.Code}}{{if .URL}}</a>{{end}}<br>
{{- end}}
{{- end}}
  </body>
</html>
`))
	var b bytes.Buffer
	if err := tmpl.Execute(&b, issues); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
