// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package validate

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
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

func TestAMPFiles_Valid(t *testing.T) {
	dir := makeTempDir(t)
	defer os.RemoveAll(dir)

	p1 := filepath.Join(dir, "1.html")
	p2 := filepath.Join(dir, "2.html")
	for _, p := range []string{p1, p2} {
		if err := ioutil.WriteFile(p, []byte(minimalAMP), 0644); err != nil {
			t.Fatalf("Failed writing %v: %v", p, err)
		}
	}
	fileIssues, err := AMPFiles(context.Background(), []string{p1, p2})
	if err != nil {
		t.Error("AMPFiles failed: ", errorString(err))
	}
	for p, issues := range fileIssues {
		if len(issues) != 0 {
			t.Errorf("AMPFiles returned issues for %v: %v", p, issues)
		}
	}
}

func TestAMPFiles_Invalid(t *testing.T) {
	dir := makeTempDir(t)
	defer os.RemoveAll(dir)

	good := filepath.Join(dir, "good.html")
	if err := ioutil.WriteFile(good, []byte(minimalAMP), 0644); err != nil {
		t.Fatalf("Failed writing %v: %v", good, err)
	}
	bad := filepath.Join(dir, "bad.html")
	badData := strings.Replace(minimalAMP, "<html amp", "<html ", 1)
	if err := ioutil.WriteFile(bad, []byte(badData), 0644); err != nil {
		t.Fatalf("Failed writing %v: %v", bad, err)
	}

	fileIssues, err := AMPFiles(context.Background(), []string{good, bad})
	if err != nil {
		t.Error("AMPFiles failed: ", errorString(err))
	}
	if len(fileIssues) != 2 {
		t.Errorf("AMPFiles reported results for %v file(s); want 2", len(fileIssues))
	}
	if got := fileIssues[good]; len(got) != 0 {
		t.Errorf("Wanted no errors for %v; got %+v", good, got)
	}
	if got := fileIssues[bad]; len(got) != 1 {
		t.Errorf("Wanted 1 error for %v; got %+v", bad, got)
	}
}

func makeTempDir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "validate_test.*")
	if err != nil {
		t.Fatal("Failed creating temp dir: ", err)
	}

	// Make sure that amphtml-validator can read the files even if it's running as a different user.
	// The directory created by T.Dir() doesn't seem to be world-readable.
	if err := os.Chmod(dir, 0755); err != nil {
		os.RemoveAll(dir)
		t.Fatalf("Failed making %v world-readable: %v", dir, err)
	}

	return dir
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
