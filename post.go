// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package validate

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
)

type fileInfo struct {
	field string // field name
	name  string // filename
	r     io.Reader
}

func post(ctx context.Context, url string, fields map[string]string, files []fileInfo) (io.ReadCloser, error) {
	// See https://stackoverflow.com/a/20397167.
	var b bytes.Buffer
	mw := multipart.NewWriter(&b)

	// Add normal non-file fields.
	for k, v := range fields {
		fw, err := mw.CreateFormField(k)
		if err != nil {
			return nil, err
		}
		if _, err := io.WriteString(fw, v); err != nil {
			return nil, err
		}
	}

	// Add files.
	for _, fi := range files {
		fw, err := mw.CreateFormFile(fi.field, fi.name)
		if err != nil {
			return nil, err
		}
		if _, err := io.Copy(fw, fi.r); err != nil {
			return nil, err
		}
	}

	// Finish the message.
	if err := mw.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, &b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
