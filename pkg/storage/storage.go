package storage

import (
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	s3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/quanxiang-cloud/fileserver/pkg/misc/config"
)

// bucket type
const (
	Private  = "private"
	Readable = "readable"
)

// Storage Storage
type Storage struct {
	client *s3.S3
}

// New New
func New(c config.Storage) (*Storage, error) {
	provider, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(c.AccessKeyID, c.SecretAccessKey, ""),
		Endpoint:    aws.String(c.Endpoint),
		Region:      aws.String(c.Region),
	})
	if err != nil {
		return nil, err
	}
	return &Storage{
		s3.New(provider),
	}, nil
}

// PutObject adds an object to a bucket.
func (s *Storage) PutObject(bucket, key string, body io.Reader, contentType string) error {
	_, err := s.client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        aws.ReadSeekCloser(body),
		ContentType: aws.String(contentType),
	})

	return err
}

// PutObjectRequest PutObjectRequest
func (s *Storage) PutObjectRequest(bucket, key string, expire time.Duration) (string, error) {
	req, _ := s.client.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	url, err := req.Presign(expire)
	if err != nil {
		return "", err
	}

	return url, nil
}

// GetObject GetObject
func (s *Storage) GetObject(bucket, key string) (io.ReadCloser, error) {
	output, err := s.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	return output.Body, nil
}

// GetObjectRequest GetObjectRequest
func (s *Storage) GetObjectRequest(bucket, key, disposition string, expire time.Duration) (string, error) {
	req, _ := s.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),

		ResponseContentDisposition: aws.String(disposition),
	})

	url, err := req.Presign(expire)
	if err != nil {
		return "", err
	}

	return url, nil
}

// CreateMultipartUpload CreateMultipartUpload
func (s *Storage) CreateMultipartUpload(bucket, key, contentType string) (string, error) {
	output, err := s.client.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", err
	}

	return *output.UploadId, nil
}

// UploadPartRequest UploadPartRequest
func (s *Storage) UploadPartRequest(bucket, key, uploadID string, partNumber int64, expire time.Duration) (string, error) {
	req, _ := s.client.UploadPartRequest(&s3.UploadPartInput{
		Bucket:     aws.String(bucket),
		Key:        aws.String(key),
		UploadId:   aws.String(uploadID),
		PartNumber: aws.Int64(partNumber),
	})

	url, err := req.Presign(expire)
	if err != nil {
		return "", err
	}

	return url, nil
}

// ListParts ListParts
func (s *Storage) ListParts(bucket, key, uploadID string) ([]*s3.Part, error) {
	output, err := s.client.ListParts(&s3.ListPartsInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
	})
	if err != nil {
		return nil, err
	}

	return output.Parts, nil
}

// CompleteMultipartUpload CompleteMultipartUpload
func (s *Storage) CompleteMultipartUpload(bucket, key, uploadID string) error {
	parts, err := s.ListParts(bucket, key, uploadID)
	if err != nil {
		return err
	}

	completedParts := make([]*s3.CompletedPart, 0, len(parts))
	for _, part := range parts {
		completedParts = append(completedParts, &s3.CompletedPart{
			ETag:       part.ETag,
			PartNumber: part.PartNumber,
		})
	}

	_, err = s.client.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: completedParts,
		},
	})

	return err
}

// AbortMultipartUpload AbortMultipartUpload
func (s *Storage) AbortMultipartUpload(bucket, key, uploadID string) error {
	_, err := s.client.AbortMultipartUpload(&s3.AbortMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
	})

	return err
}

// DeleteObject DeleteObject
func (s *Storage) DeleteObject(bucket, key string) error {
	_, err := s.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	return err
}
