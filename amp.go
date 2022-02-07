// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package validate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strconv"
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
	issues, err := runAMP(ctx, []string{"-"}, r)
	return issues["-"], err
}

// AMPFiles runs amphtml-validator to validate multiple AMP HTML files at the supplied paths.
// The returned map is keyed by the filenames from the paths argument.
//
// AMPFiles may be much faster than AMP when validating multiple files, since the
// WebAssembly-based amphtml-validator can take a substantial amount of time to start:
// https://github.com/ampproject/amphtml/issues/37585.
func AMPFiles(ctx context.Context, paths []string) (map[string][]Issue, error) {
	return runAMP(ctx, paths, nil)
}

// runAMP runs the amphtml-validator command with the provided filename arguments and stdin
// (possibly nil) and parses the results. The returned map is keyed by filename (or "-" if
// it was passed to tell the validator to read input from stdin).
func runAMP(ctx context.Context, fileArgs []string, stdin io.Reader) (map[string][]Issue, error) {
	const exe = "amphtml-validator"
	if _, err := exec.LookPath(exe); err != nil {
		return nil, err
	}
	var stdout bytes.Buffer
	cmd := exec.CommandContext(ctx, exe, append([]string{"--format=json"}, fileArgs...)...)
	cmd.Stdin = stdin
	cmd.Stdout = &stdout

	// amphtml-validator appears to exit with 1 if it identifies errors (but not just warnings).
	// Only report other errors here.
	runErr := cmd.Run()
	if runErr != nil {
		if _, ok := runErr.(*exec.ExitError); !ok {
			return nil, runErr
		}
	}

	// amphtml-validator prints a JSON object that maps from the filenames that were passed
	// to it (or "-" for stdin) to each file's results object.
	type result struct {
		// This is a subset of ValidationResult in
		// https://github.com/ampproject/amphtml/blob/master/validator/validator.proto.
		Status string `json:"status"` // UNKNOWN, PASS, FAIL
		Errors []struct {
			Severity string          `json:"severity"` // UNKNOWN_SEVERITY, ERROR, WARNING
			Line     int             `json:"line"`
			Col      int             `json:"col"`
			Message  string          `json:"message"`
			Code     json.RawMessage `json:"code"` // either number or string? see below
			SpecURL  string          `json:"specUrl"`
		} `json:"errors"`
	}
	var out map[string]result
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		return nil, err
	}

	allPassed := true
	var allIssues []Issue
	fileIssues := make(map[string][]Issue)
	for fn, res := range out {
		var issues []Issue
		for _, e := range res.Errors {
			is := Issue{
				Line:    e.Line,
				Col:     e.Col + 1, // these appear to be 0-indexed
				Message: e.Message,
				URL:     e.SpecURL,
			}
			if e.Severity == "WARNING" {
				is.Severity = Warning
			}

			// It looks like amphtml-validator got changed at some point (May 2021?) such that the
			// 'code' field is a number (e.g. 5) rather than a string (e.g. "MANDATORY_ATTR_MISSING").
			// Handle either case.
			var scode string
			var icode int
			if err := json.Unmarshal(e.Code, &scode); err == nil {
				is.Code = scode
			} else if err := json.Unmarshal(e.Code, &icode); err == nil {
				is.Code = strconv.Itoa(icode)
			}

			issues = append(issues, is)
		}
		fileIssues[fn] = issues
		allIssues = append(allIssues, issues...)

		if res.Status != "PASS" {
			allPassed = false
		}
	}

	if allPassed && runErr != nil {
		return fileIssues, fmt.Errorf("%v reported pass but exited with error: %v", exe, runErr)
	}
	return fileIssues, checkResponse(allPassed, allIssues)
}
