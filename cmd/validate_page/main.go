// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/derat/validate"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTION] <FILE>\n"+
			"Validate an HTML document.\n"+
			"If <FILE> isn't supplied, reads from stdin.\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	browser := flag.Bool("browser", false,
		"Display validation issues in browser (printed to stdout otherwise)")
	fileType := flag.String("type", "", `File type ("html"; inferred if empty)`)
	flag.Parse()

	var r io.Reader
	switch len(flag.Args()) {
	case 0:
		r = os.Stdin
	case 1:
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to open input file:", err)
			os.Exit(1)
		}
		defer f.Close()
		r = f
	default:
		flag.Usage()
		os.Exit(2)
	}

	var issues []string
	var out []byte
	var err error

	switch *fileType {
	case "html", "":
		issues, out, err = validate.HTML(context.Background(), r)
	default:
		fmt.Fprintf(os.Stderr, "Bad -type value %q\n", *fileType)
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Validation request failed:", err)
		os.Exit(1)
	}

	if len(issues) > 0 {
		if *browser {
			if err := validate.LaunchBrowser(out); err != nil {
				fmt.Fprintln(os.Stderr, "Failed to display results in browser:", err)
				os.Exit(1)
			}
		} else {
			for _, is := range issues {
				fmt.Println(is)
				fmt.Println()
			}
		}
	}
}
