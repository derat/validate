// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package validate

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func TestAMP_Valid(t *testing.T) {
	issues, err := AMP(context.Background(), strings.NewReader(minimalAMP))
	if err != nil {
		t.Error("AMP reported error:", errorString(err))
	}
	if len(issues) != 0 {
		t.Errorf("AMP returned issues: %v", issues)
	}
}

func TestAMP_Invalid(t *testing.T) {
	doc := strings.Replace(minimalAMP, "<html amp", "<html ", 1) + "  <bogus></bogus>\n"
	issues, err := AMP(context.Background(), strings.NewReader(doc))
	if err != nil {
		t.Error("AMP reported error for invalid document:", errorString(err))
	}
	if len(issues) != 2 {
		t.Errorf("AMP returned %v issues (%q); want 2", len(issues), issues)
	} else {
		check := func(got, want Issue, altCode string, needURL bool) {
			// URLs seem likely to change, so just check that one was set.
			if needURL && got.URL == "" {
				t.Errorf("AMP reported issue %q with unexpectedly empty URL", got)
			}
			got.URL = ""

			altWant := want
			altWant.Code = altCode
			if got != want && got != altWant {
				t.Errorf("AMP reported issue %q; want %q", got, want)
			}
		}
		check(issues[0], Issue{
			Severity: Error,
			Line:     2,
			Col:      1,
			Message:  "The mandatory attribute '⚡' is missing in tag 'html'.",
			Code:     "5",
		}, "MANDATORY_ATTR_MISSING", true /* needURL */)
		check(issues[1], Issue{
			Severity: Error,
			Line:     26,
			Col:      3,
			Message:  "The tag 'bogus' is disallowed.",
			Code:     "2",
		}, "DISALLOWED_TAG", false /* needURL */)
	}
}

func errorString(err error) string {
	s := err.Error()
	if exitErr, ok := err.(*exec.ExitError); ok {
		s += fmt.Sprintf(" (%v)", string(exitErr.Stderr))
	}
	return s
}

// This comes from https://amp.dev/documentation/guides-and-tutorials/start/create/basic_markup/.
const minimalAMP = `<!doctype html>
<html amp lang="en">
  <head>
    <meta charset="utf-8">
    <script async src="https://cdn.ampproject.org/v0.js"></script>
    <title>Hello, AMPs</title>
    <link rel="canonical" href="https://amp.dev/documentation/guides-and-tutorials/start/create/basic_markup/">
    <meta name="viewport" content="width=device-width,minimum-scale=1,initial-scale=1">
    <script type="application/ld+json">
      {
        "@context": "http://schema.org",
        "@type": "NewsArticle",
        "headline": "Open-source framework for publishing content",
        "datePublished": "2015-10-07T12:02:41Z",
        "image": [
          "logo.jpg"
        ]
      }
    </script>
    <style amp-boilerplate>body{-webkit-animation:-amp-start 8s steps(1,end) 0s 1 normal both;-moz-animation:-amp-start 8s steps(1,end) 0s 1 normal both;-ms-animation:-amp-start 8s steps(1,end) 0s 1 normal both;animation:-amp-start 8s steps(1,end) 0s 1 normal both}@-webkit-keyframes -amp-start{from{visibility:hidden}to{visibility:visible}}@-moz-keyframes -amp-start{from{visibility:hidden}to{visibility:visible}}@-ms-keyframes -amp-start{from{visibility:hidden}to{visibility:visible}}@-o-keyframes -amp-start{from{visibility:hidden}to{visibility:visible}}@keyframes -amp-start{from{visibility:hidden}to{visibility:visible}}</style><noscript><style amp-boilerplate>body{-webkit-animation:none;-moz-animation:none;-ms-animation:none;animation:none}</style></noscript>
  </head>
  <body>
    <h1>Welcome to the mobile web</h1>
  </body>
</html>
`
