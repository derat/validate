// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package validate

import (
	"context"
	"strings"
	"testing"
)

func TestHTML_Valid(t *testing.T) {
	issues, out, err := HTML(context.Background(), strings.NewReader(`<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8">
    <title>The title</title>
  </head>
  <body>Here's some text.</body>
</html>
`))
	if err != nil {
		t.Error("HTML reported error: ", err)
	}
	if len(issues) != 0 {
		t.Errorf("HTML returned issues: %v", issues)
	}
	if len(out) == 0 {
		t.Error("HTML returned empty output")
	}
}

func TestHTML_Invalid(t *testing.T) {
	issues, out, err := HTML(context.Background(), strings.NewReader(`<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8">
    <title>The title</title>
  </head>
  <body>
    <bogus></bogus>
  </body>
</html>
`))
	if err != nil {
		t.Error("HTML reported error for invalid document: ", err)
	}
	if len(issues) != 1 {
		t.Errorf("HTML returned %v issues (%q); want 1", len(issues), issues)
	} else {
		is := issues[0]
		if want := "Element bogus not allowed as child of element body"; !strings.Contains(is.Message, want) {
			t.Errorf("HTML returned issue %q that doesn't contain %q", is, want)
		}
		// The error appears to be reported at the start of the closing tag.
		if is.Line != 8 || is.Col != 11 {
			t.Errorf("HTML reported issue at %d:%d; want 8:11", is.Line, is.Col)
		}
	}
	if len(out) == 0 {
		t.Error("HTML returned empty output")
	}
}
