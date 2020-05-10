// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package validate

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os/exec"
)

// AMP reads an AMP HTML document from r and validates it by running the amphtml-validator program,
// which must be present in $PATH. Issues identified by the validator are parsed and returned.
// If the returned error is non-nil, an issue occurred in the validation process.
//
// There unfortunately don't appear to be any online AMP validators that can be called programatically.
// As a result, this function requires that the amphtml-validator Node.js program is installed locally. See
// https://amp.dev/documentation/guides-and-tutorials/learn/validation-workflow/validate_amp/#command-line-tool
// for more information about installing and running amphtml-validator.
//
// More info about the lack of online validators:
//
// The official AMP validator at https://validator.ampproject.org/ performs validation in the browser
// rather providing an HTTP server accepting posted documents.
//
// Google provides an online AMP (and structured data) validator at https://search.google.com/test/amp, but
// it's very slow and it's not clear to me whether there's an easy way to use it outside of a web browser.
//
// Cloudflare announced an online validator in 2017 at https://blog.cloudflare.com/amp-validator-api/,
// but it has since been shut down. That post points at https://blog.cloudflare.com/announcing-amp-real-url/
// as justification, although I don't see any explanation of the validator's disappearance there.
//
// The ampbench project seems to have provided an online validator, but it was apparently shut down in 2019
// due to the codebase getting out of date: https://github.com/ampproject/ampbench/issues/126
//
// There's more discussion at https://github.com/ampproject/amphtml/issues/1968.
func AMP(ctx context.Context, r io.Reader) ([]Issue, error) {
	const exe = "amphtml-validator"
	if _, err := exec.LookPath(exe); err != nil {
		return nil, err
	}
	var b bytes.Buffer
	cmd := exec.CommandContext(ctx, exe, "--format=json", "-")
	cmd.Stdin = r
	cmd.Stdout = &b
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	// For reasons unclear to me, the JSON object printed by amphtml-validator is wrapped
	// in an object with a single "-" property that holds the actual results object.
	var out = struct {
		Result struct {
			// This is a subset of ValidationResult in
			// https://github.com/ampproject/amphtml/blob/master/validator/validator.proto.
			Status string `json:"status"` // UNKNOWN, PASS, FAIL
			Errors []struct {
				Severity string `json:"severity"` // UNKNOWN_SEVERITY, ERROR, WARNING
				Line     int    `json:"line"`
				Col      int    `json:"col"`
				Message  string `json:"message"`
				Code     string `json:"code"`
				SpecURL  string `json:"specUrl"`
			} `json:"errors"`
		} `json:"-,"` // trailing comma since '-' usually means 'ignore'
	}{}
	if err := json.Unmarshal(b.Bytes(), &out); err != nil {
		return nil, err
	}

	var issues []Issue
	res := out.Result
	for _, e := range res.Errors {
		is := Issue{
			Line:    e.Line,
			Col:     e.Col + 1, // these appear to be 0-indexed
			Message: e.Message,
			Code:    e.Code,
			URL:     e.SpecURL,
		}
		if e.Severity == "WARNING" {
			is.Severity = Warning
		}
		issues = append(issues, is)
	}
	return issues, checkResponse(res.Status == "PASS", issues)
}
