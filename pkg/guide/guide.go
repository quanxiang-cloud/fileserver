package guide

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/quanxiang-cloud/cabin/tailormade/client"
)

const (
	contentTypeKey = "Content-Type"
	contentType    = "application/octet-stream"
)

const (
	domainPath        = "%s/api/v1/fileserver/domain"
	uploadPath        = "%s/api/v1/fileserver/sign/upload"
	downloadPath      = "%s/api/v1/fileserver/sign/download"
	initMultipartPath = "%s/api/v1/fileserver/sign/initMultipart"
	uploadPartPath    = "%s/api/v1/fileserver/sign/uploadMultipart"
	completePath      = "%s/api/v1/fileserver/sign/completeMultipart"
	deletePath        = "%s/api/v1/fileserver/del"
	finishPath        = "%s/api/v1/fileserver/sign/finish"
)

const (
	// the maximum upload limit of a single file. If the limit is exceeded, it will be uploaded by fragment.
	defaultLimit = 5 * 1024 * 1024 // 30MB
	byteSize     = 5 * 1024 * 1024 // 5MB
)

const (
	defaultTimeout = 20 * time.Second
	maxIdleConns   = 10
)

// Guide Guide.
type Guide struct {
	endpoint   string
	bucket     string
	readBucket string
	client     *http.Client
}

// Option option.
type Option func(*Guide)

// WithHTTPClient WithHTTPClient.
func WithHTTPClient(timeout time.Duration, maxIdleConns int) Option {
	return func(g *Guide) {
		cli := client.New(client.Config{
			Timeout:      timeout,
			MaxIdleConns: maxIdleConns,
		})

		g.client = &cli
	}
}

// NewGuide NewGuide.
func NewGuide(opts ...Option) (*Guide, error) {
	endpoint := os.Getenv("FILESERVER_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://fileserver"
	}

	cli := client.New(client.Config{
		Timeout:      defaultTimeout,
		MaxIdleConns: maxIdleConns,
	})

	g := &Guide{
		endpoint: endpoint,
		client:   &cli,
	}

	for _, opt := range opts {
		opt(g)
	}

	resp, err := g.getBucket()
	if err != nil {
		return nil, err
	}

	g.bucket = resp.Private
	g.readBucket = resp.Readable

	return g, nil
}

// Resp get return parameters.
type Resp struct {
	Private  string `json:"private"`
	Readable string `json:"readable"`
	URL      string `json:"url"`
	UploadID string `json:"uploadID"`
}

// UploadFile upload file.
func (g *Guide) UploadFile(ctx context.Context, path string, r io.Reader, size int64) error {
	path = filepath.Join(g.bucket, path)
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

func (g *Guide) multipartUpload(ctx context.Context, path string, r io.Reader, size int64) error {
	partNums := getPartNums(size)

	// get uploadid
	resp, err := g.getUploadID(ctx, path)
	if err != nil {
		return err
	}

	byteArr := make([]byte, byteSize)
	for i := 1; i <= partNums; i++ {
		partSize, err := r.Read(byteArr)
		if err != nil {
			return err
		}
		// get block upload link
		partResp, err := g.getPartUploadURL(ctx, i, int64(partSize), path, resp.UploadID)
		if err != nil {
			return err
		}
		// upload block
		request, err := http.NewRequest(http.MethodPut, partResp.URL, bytes.NewReader(byteArr[:partSize]))
		if err != nil {
			return err
		}
		request.ContentLength = int64(partSize)
		request.Header.Set(contentTypeKey, contentType)

		_, err = g.client.Do(request)
		if err != nil {
			return err
		}
	}

	// merge block
	return g.completeMultipart(ctx, path, resp.UploadID)
}

// DownloadFile download file.
func (g *Guide) DownloadFile(ctx context.Context, path string, w io.Writer) error {
	resp := &Resp{}
	url := g.getRequestURL(downloadPath)
	err := client.POST(
		ctx, g.client, url,
		struct {
			Path string `json:"path"`
		}{
			Path: filepath.Join(g.bucket, path),
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

// DeleteFile delete file.
func (g *Guide) DeleteFile(ctx context.Context, path string) error {
	resp := &Resp{}
	url := g.getRequestURL(deletePath)
	err := client.POST(
		ctx, g.client, url,
		struct {
			Path string `json:"path"`
		}{
			Path: filepath.Join(g.bucket, path),
		},
		resp,
	)

	return err
}

func (g *Guide) getBucket() (*Resp, error) {
	resp := &Resp{}
	url := g.getRequestURL(domainPath)
	err := client.POST(
		context.Background(), g.client, url,
		nil,
		resp,
	)

	return resp, err
}

func (g *Guide) getUploadURL(ctx context.Context, path string) (*Resp, error) {
	resp := &Resp{}

	url := g.getRequestURL(uploadPath)
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
		return nil, err
	}

	return resp, nil
}

func (g *Guide) getUploadID(ctx context.Context, path string) (*Resp, error) {
	resp := &Resp{}

	url := g.getRequestURL(initMultipartPath)
	err := client.POST(
		ctx, g.client, url,
		struct {
			Path        string `json:"path"`
			ContentType string `json:"contentType"`
		}{
			Path:        path,
			ContentType: contentType,
		},
		resp,
	)

	return resp, err
}

func getPartNums(size int64) int {
	partNum := int(size / byteSize)

	if (size % byteSize) == 0 {
		return partNum
	}

	return partNum + 1
}

func (g *Guide) getPartUploadURL(ctx context.Context, partNumber int, partSize int64, path, uploadID string) (*Resp, error) {
	resp := &Resp{}

	url := g.getRequestURL(uploadPartPath)
	err := client.POST(
		ctx, g.client, url,
		struct {
			UploadID   string `json:"uploadID"`
			PartNumber int    `json:"partNumber"`
			Path       string `json:"path"`
		}{
			UploadID:   uploadID,
			PartNumber: partNumber,
			Path:       path,
		},
		resp,
	)

	return resp, err
}

func (g *Guide) completeMultipart(ctx context.Context, path, uploadID string) error {
	resp := &Resp{}

	url := g.getRequestURL(completePath)
	err := client.POST(
		ctx, g.client, url,
		struct {
			UploadID string `json:"uploadID"`
			Path     string `json:"path"`
		}{
			UploadID: uploadID,
			Path:     path,
		},
		resp,
	)

	return err
}

func (g *Guide) finish(ctx context.Context, path string) error {
	resp := &Resp{}

	url := g.getRequestURL(finishPath)
	err := client.POST(
		ctx, g.client, url,
		struct {
			Path string `json:"path"`
		}{
			Path: path,
		},
		resp,
	)
	return err
}

func (g *Guide) getRequestURL(format string) string {
	return fmt.Sprintf(format, g.endpoint)
}
