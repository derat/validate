# validate

[![GoDoc](https://godoc.org/github.com/derat/validate?status.svg)](https://godoc.org/github.com/derat/validate)
![Build Status](https://storage.googleapis.com/derat-build-badges/8e6f9183-1faf-4d21-bf76-51440cb8c265.svg)

The `github.com/derat/validate` Go package validates:

*   HTML documents using the W3C's online "[Nu Html Checker]"
*   CSS using the W3C's online [CSS Validation Service]
*   AMP HTML documents using a local installation of the [amphtml-validator]
    Node.js program

Supplied documents are uploaded (in the case of online validators) and any
issues identified by the validator are parsed and returned.

[Nu Html Checker]: https://validator.w3.org/nu/
[CSS Validation Service]: https://jigsaw.w3.org/css-validator/
[amphtml-validator]: https://www.npmjs.com/package/amphtml-validator

## Usage

```go
import (
	"github.com/derat/validate"
)
// ...
	f, err := os.Open("page.html")
	if err != nil {
		// ...
	}
	defer f.Close()

	// See also validate.AMP() and validate.CSS().
	issues, out, err := validate.HTML(context.Background(), f)
	if err != nil {
		// ...
	}

	// Iterate over the parsed issues.
	for _, is := range issues {
		fmt.Println(is)
	}
	// Display the full results page in a browser.
	if err := validate.LaunchBrowser(out); err != nil {
		// ...
	}
```

A command-line program named [validate_page](./cmd/validate_page/main.go) that
validates a file or a document provided over stdin is also provided:

```sh
% go install github.com/derat/validate/cmd/validate_page
% validate_page -browser page.html
% validate_page -type=amp index.amp.html
% validate_page <style.css
% validate_page -type=htmlcss -browser page.html  # check CSS in HTML doc
```
