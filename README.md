# validate

[![GoDoc](https://godoc.org/github.com/derat/validate?status.svg)](https://godoc.org/github.com/derat/validate)

The `github.com/derat/validate` Go package validates HTML documents using the
W3C's "[Nu Html Checker]" and CSS using the W3C's [CSS Validation Service].
Supplied documents are uploaded and any issues identified by the services are
parsed and returned.

[Nu Html Checker]: https://validator.w3.org/nu/
[CSS Validation Service]: https://jigsaw.w3.org/css-validator/

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

	// See also validate.CSS().
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
% validate_page <style.css
% validate_page -type=htmlcss -browser page.html  # check CSS in HTML doc
```
