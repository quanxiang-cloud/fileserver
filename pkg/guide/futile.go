package guide

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/quanxiang-cloud/cabin/tailormade/client"
)

type bucket int

const (
	Private bucket = iota + 1
	Readable
)

// FutileUploadFile upload file.
func (g *Guide) FutileUploadFile(ctx context.Context, path string, r io.Reader, size int64, bkt bucket) error {
	switch bkt {
	case Private:
		path = filepath.Join(g.bucket, path)
	case Readable:
		path = filepath.Join(g.readBucket, path)
	default:
		return fmt.Errorf("unknown bucket type: %d", bkt)
	}

	if size > defaultLimit {
		err := g.multipartUpload(ctx, path, r, size)
		if err != nil {
			return err
		}

		return g.finish(ctx, path)
	}

	resp, err := g.getUploadURL(ctx, path)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(http.MethodPut, resp.URL, r)
	if err != nil {
		return err
	}
	request.ContentLength = size
	request.Header.Set(contentTypeKey, contentType)

	response, err := g.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	return g.finish(ctx, path)
}

// FutileDownloadFile FutileDownloadFile.
func (g *Guide) FutileDownloadFile(ctx context.Context, path string, w io.Writer, bkt bucket) error {
	switch bkt {
	case Private:
		path = filepath.Join(g.bucket, path)
	case Readable:
		path = filepath.Join(g.readBucket, path)
	default:
		return fmt.Errorf("unknown bucket type: %d", bkt)
	}

	resp := &Resp{}
	url := g.getRequestURL(downloadPath)
	err := client.POST(
		ctx, g.client, url,
		struct {
			Path string `json:"path"`
		}{
			Path: path,
		},
		resp,
	)
	if err != nil {
		return err
	}

	response, err := http.Get(resp.URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	buf := make([]byte, byteSize)
	_, err = io.CopyBuffer(w, response.Body, buf)
	return err
}
