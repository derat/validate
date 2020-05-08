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
	} else if want := "Element bogus not allowed as child of element body"; !strings.Contains(issues[0], want) {
		t.Errorf("HTML returned issue %q that doesn't contain %q", issues[0], want)
	}
	if len(out) == 0 {
		t.Error("HTML returned empty output")
	}
}
