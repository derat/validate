// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package validate

import (
	"context"
	"strings"
	"testing"
)

func TestCSS_ValidCSS(t *testing.T) {
	issues, out, err := CSS(context.Background(), strings.NewReader(`
body {
  background-color: white;
  margin: 0;
}
`), Stylesheet)
	if err != nil {
		t.Error("CSS reported error: ", err)
	}
	if len(issues) != 0 {
		t.Errorf("CSS returned issues: %v", issues)
	}
	if len(out) == 0 {
		t.Error("CSS returned empty output")
	}
}

func TestCSS_InvalidCSS(t *testing.T) {
	issues, out, err := CSS(context.Background(), strings.NewReader(`
body {
  invalid-property: #aaa;
}
`), Stylesheet)
	if err != nil {
		t.Error("CSS reported error for invalid stylesheet: ", err)
	}
	if len(issues) != 1 {
		t.Errorf("CSS returned %v issues (%q); want 1", len(issues), issues)
	} else {
		is := issues[0]
		if want := "Property invalid-property doesn't exist"; !strings.Contains(is.Message, want) {
			t.Errorf("CSS returned issue %q that doesn't contain %q", is, want)
		}
		if want := 3; is.Line != want {
			t.Errorf("CSS returned issue on line %d; want %d", is.Line, want)
		}
	}
	if len(out) == 0 {
		t.Error("CSS returned empty output")
	}
}

func TestCSS_ValidHTML(t *testing.T) {
	issues, out, err := CSS(context.Background(), strings.NewReader(`
<html>
  <head>
    <meta charset="utf-8">
    <title>The title</title>
	<style>body{margin:0}</style>
  </head>
  <body></body>
</html>
}
`), HTMLDoc)
	if err != nil {
		t.Error("CSS reported error: ", err)
	}
	if len(issues) != 0 {
		t.Errorf("CSS returned issues: %v", issues)
	}
	if len(out) == 0 {
		t.Error("CSS returned empty output")
	}
}

func TestCSS_InvalidHTML(t *testing.T) {
	issues, out, err := CSS(context.Background(), strings.NewReader(`
<html>
  <head>
    <meta charset="utf-8">
    <title>The title</title>
	<style>body{invalid-property:0}</style>
  </head>
  <body></body>
</html>
`), HTMLDoc)
	if err != nil {
		t.Error("CSS reported error for invalid stylesheet: ", err)
	}
	if len(issues) != 1 {
		t.Errorf("CSS returned %v issues (%q); want 1", len(issues), issues)
	} else {
		is := issues[0]
		if want := "Property invalid-property doesn't exist"; !strings.Contains(is.Message, want) {
			t.Errorf("CSS returned issue %q that doesn't contain %q", is, want)
		}
		if want := 6; is.Line != want {
			t.Errorf("CSS returned issue on line %d; want %d", is.Line, want)
		}
		// The CSS validator doesn't provide columns.
	}
	if len(out) == 0 {
		t.Error("CSS returned empty output")
	}
}
