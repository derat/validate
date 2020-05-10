// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package validate

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
)

// LaunchBrowser launches a web browser with the supplied HTML page.
// It can be used to display results pages returned by the CSS and HTML functions.
func LaunchBrowser(page []byte) error {
	// If X isn't running, just pipe the results into w3m.
	if os.Getenv("DISPLAY") == "" {
		cmd := exec.Command("w3m", "-T", "text/html")
		cmd.Stdin = bytes.NewReader(page)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// Otherwise, write the results to a temporary file and open it in the user's preferred browser.
	p, err := writeResults(page)
	if err != nil {
		return err
	}
	cmd := exec.Command("xdg-open", p)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// writeResults writes page to a new temporary file and returns its path.
func writeResults(page []byte) (string, error) {
	f, err := ioutil.TempFile("", "validate.*.html")
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := f.Write(page); err != nil {
		return "", err
	}
	return f.Name(), nil
}
