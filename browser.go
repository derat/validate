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
// It can be used to display the HTML function's out return value.
func LaunchBrowser(out []byte) error {
	// If X isn't running, just pipe the results into w3m.
	if os.Getenv("DISPLAY") == "" {
		cmd := exec.Command("w3m", "-T", "text/html")
		cmd.Stdin = bytes.NewReader(out)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// Otherwise, write the results to a temporary file and open it in the user's preferred browser.
	p, err := writeResults(out)
	if err != nil {
		return err
	}
	cmd := exec.Command("xdg-open", p)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// writeResults writes out to a new temporary file and returns its path.
func writeResults(out []byte) (string, error) {
	f, err := ioutil.TempFile("", "validate.*.html")
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := f.Write(out); err != nil {
		return "", err
	}
	return f.Name(), nil
}
