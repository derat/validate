// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package validate

import (
	"context"
	"strings"
	"testing"
)

func TestAMP_Valid(t *testing.T) {
	issues, err := AMP(context.Background(), strings.NewReader(minimalAMP))
	if err != nil {
		t.Error("AMP reported error: ", err)
	}
	if len(issues) != 0 {
		t.Errorf("AMP returned issues: %v", issues)
	}
}

func TestAMP_Invalid(t *testing.T) {
	doc := strings.Replace(minimalAMP, "<html amp", "<html ", 1)
	issues, err := AMP(context.Background(), strings.NewReader(doc))
	if err != nil {
		t.Error("AMP reported error for invalid document: ", err)
	}
	if len(issues) != 1 {
		t.Errorf("AMP returned %v issues (%q); want 1", len(issues), issues)
	} else {
		is := issues[0]
		is.URL = ""
		if want := (Issue{
			Severity: Error,
			Line:     2,
			Message:  "The mandatory attribute '⚡' is missing in tag 'html'.",
			Code:     "MANDATORY_ATTR_MISSING",
		}); is != want {
			t.Errorf("AMP reported issue %q; want %q", is, want)
		}
		// URLs seem likely to change, but make sure that one is at least present.
		if issues[0].URL == "" {
			t.Errorf("AMP didn't return URL for issue %q", issues[0])
		}
	}
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
