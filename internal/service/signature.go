package service

import (
	"context"
	"fmt"

	error2 "github.com/quanxiang-cloud/cabin/error"
	id2 "github.com/quanxiang-cloud/cabin/id"
	"github.com/quanxiang-cloud/cabin/logger"
	"github.com/quanxiang-cloud/cabin/tailormade/header"
	time2 "github.com/quanxiang-cloud/cabin/time"
	"github.com/quanxiang-cloud/fileserver/internal/models"
	"github.com/quanxiang-cloud/fileserver/pkg/misc/code"
	"github.com/quanxiang-cloud/fileserver/pkg/utils"
)

// PresignedUploadReq PresignedUploadReq.
type PresignedUploadReq struct {
	Path string `json:"path" binding:"required"`
}

// PresignedUploadResp PresignedUploadResp.
type PresignedUploadResp struct {
	URL string `json:"url"`
}

func (f *fileserver) PresignedUpload(ctx context.Context, req *PresignedUploadReq) (*PresignedUploadResp, error) {
	bucket, path := utils.Split(req.Path, "/")
	if !utils.ExistBucket(f.conf.Buckets, bucket) {
		logger.Logger.WithName("presigned upload").Infow("invalid storage", header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, error2.New(code.InvalidStorage)
	}

	expire := f.conf.Storage.URLExpire
	url, err := f.storages.PutObjectRequest(bucket, path, expire)
	if err != nil {
		logger.Logger.WithName("presigned upload").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, error2.New(code.ErrSinger)
	}

	return &PresignedUploadResp{
		URL: url,
	}, nil
}

// PresignedDownloadReq PresignedDownloadReq.
type PresignedDownloadReq struct {
	Path     string `json:"path" binding:"required"`
	FileName string `json:"fileName"`
}

// PresignedDownloadResp PresignedDownloadResp.
type PresignedDownloadResp struct {
	URL string `json:"url"`
}

func (f *fileserver) PresignedDownload(ctx context.Context, req *PresignedDownloadReq) (*PresignedDownloadResp, error) {
	bucket, path := utils.Split(req.Path, "/")
	if !utils.ExistBucket(f.conf.Buckets, bucket) {
		logger.Logger.WithName("presigned upload").Infow("invalid storage", header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, error2.New(code.InvalidStorage)
	}

	info, err := f.fileServerRepo.GetByPath(f.db, path)
	if err != nil {
		logger.Logger.WithName("presigned upload").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, err
	}

	if info == nil {
		logger.Logger.WithName("presigned upload").Infow("file not found", header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, error2.New(code.InvalidExist)
	}

	var disposition string
	if req.FileName != "" {
		filename := utils.URLQueryEscape(req.FileName)
		disposition = fmt.Sprintf("attachment; filename=\"%q\"; filename*=utf-8''%s", filename, filename)
	}

	expire := f.conf.Storage.URLExpire
	url, err := f.storages.GetObjectRequest(bucket, path, disposition, expire)
	if err != nil {
		logger.Logger.WithName("presigned upload").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, error2.New(code.ErrSinger)
	}

	return &PresignedDownloadResp{
		URL: url,
	}, nil
}

// InitMultipartUploadReq InitMultipartUploadReq.
type InitMultipartUploadReq struct {
	Path        string `json:"path" binding:"required"`
	ContentType string `json:"contentType" binding:"required"`
}

// InitMultipartUploadResp InitMultipartUploadResp.
type InitMultipartUploadResp struct {
	UploadID string `json:"uploadID"`
}

func (f *fileserver) InitMultipartUpload(ctx context.Context, req *InitMultipartUploadReq) (*InitMultipartUploadResp, error) {
	bucket, path := utils.Split(req.Path, "/")
	if !utils.ExistBucket(f.conf.Buckets, bucket) {
		logger.Logger.WithName("init multipart upload").Infow("invalid storage", header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, error2.New(code.InvalidStorage)
	}

	uploadID, err := f.multipartRepo.Get(ctx, path)
	if err != nil {
		logger.Logger.WithName("init multipart upload").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, err
	}

	if uploadID != "" {
		return &InitMultipartUploadResp{
			UploadID: uploadID,
		}, nil
	}

	uploadID, err = f.storages.CreateMultipartUpload(bucket, path, req.ContentType)
	if err != nil {
		logger.Logger.WithName("init multipart upload").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, error2.New(code.ErrSinger)
	}

	err = f.multipartRepo.Create(ctx, path, uploadID, f.conf.Storage.PartExpire)
	if err != nil {
		logger.Logger.WithName("init multipart upload").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, err
	}

	return &InitMultipartUploadResp{
		UploadID: uploadID,
	}, nil
}

// PresignedMultipartReq  PresignedMultipartReq.
type PresignedMultipartReq struct {
	UploadID   string `json:"uploadID" binding:"required"`
	PartNumber int64  `json:"partNumber" binding:"required"`
	Path       string `json:"path" binding:"required"`
}

// PresignedMultipartResp PresignedMultipartResp.
type PresignedMultipartResp struct {
	URL string `json:"url"`
}

func (f *fileserver) PresignedMultipart(ctx context.Context, req *PresignedMultipartReq) (*PresignedMultipartResp, error) {
	bucket, path := utils.Split(req.Path, "/")
	if !utils.ExistBucket(f.conf.Buckets, bucket) {
		logger.Logger.WithName("presigned multipart").Infow("invalid storage", header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, error2.New(code.InvalidStorage)
	}

	expire := f.conf.Storage.URLExpire
	url, err := f.storages.UploadPartRequest(bucket, path, req.UploadID, req.PartNumber, expire)
	if err != nil {
		logger.Logger.WithName("presigned multipart").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, error2.New(code.ErrSinger)
	}

	return &PresignedMultipartResp{
		URL: url,
	}, nil
}

// ListMultiPartsReq ListMultiPartsReq.
type ListMultiPartsReq struct {
	Path     string `json:"path" binding:"required"`
	UploadID string `json:"uploadID" binding:"required"`
}

// ListMultiPartsResp ListMultiPartsResp.
type ListMultiPartsResp struct {
	Parts []int64 `json:"parts"`
}

func (f *fileserver) ListMultiParts(ctx context.Context, req *ListMultiPartsReq) (*ListMultiPartsResp, error) {
	bucket, path := utils.Split(req.Path, "/")
	if !utils.ExistBucket(f.conf.Buckets, bucket) {
		logger.Logger.WithName("list multipart").Infow("invalid storage", header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, error2.New(code.InvalidStorage)
	}

	s3Parts, err := f.storages.ListParts(bucket, path, req.UploadID)
	if err != nil {
		logger.Logger.WithName("list multipart").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, error2.New(code.ErrListMultiPart)
	}

	parts := make([]int64, 0, len(s3Parts))
	for _, p := range s3Parts {
		parts = append(parts, *p.PartNumber)
	}

	return &ListMultiPartsResp{Parts: parts}, nil
}

// CompleteMultiPartsReq CompleteMultiPartsReq.
type CompleteMultiPartsReq struct {
	Path     string `json:"path" binding:"required"`
	UploadID string `json:"uploadID" binding:"required"`
}

// CompleteMultiPartsResp CompleteMultiPartsResp.
type CompleteMultiPartsResp struct{}

func (f *fileserver) CompleteMultiParts(ctx context.Context, req *CompleteMultiPartsReq) (*CompleteMultiPartsResp, error) {
	bucket, path := utils.Split(req.Path, "/")
	if !utils.ExistBucket(f.conf.Buckets, bucket) {
		logger.Logger.WithName("complete multipart").Infow("invalid storage", header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, error2.New(code.InvalidStorage)
	}

	err := f.storages.CompleteMultipartUpload(bucket, path, req.UploadID)
	if err != nil {
		logger.Logger.WithName("complete multipart").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, error2.New(code.ErrCompleteMultiPart)
	}

	err = f.multipartRepo.Delete(ctx, path)
	if err != nil {
		logger.Logger.WithName("complete multipart").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, err
	}

	return &CompleteMultiPartsResp{}, nil
}

// AbortMultipartUploadReq AbortMultipartUploadReq.
type AbortMultipartUploadReq struct {
	Path     string `json:"path" binding:"required"`
	UploadID string `json:"uploadID" binding:"required"`
}

// AbortMultipartUploadResp AbortMultipartUploadResp.
type AbortMultipartUploadResp struct{}

func (f *fileserver) AbortMultipartUpload(ctx context.Context, req *AbortMultipartUploadReq) (*AbortMultipartUploadResp, error) {
	bucket, path := utils.Split(req.Path, "/")
	if !utils.ExistBucket(f.conf.Buckets, bucket) {
		logger.Logger.WithName("abort multipart").Infow("invalid storage", header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, error2.New(code.InvalidStorage)
	}

	err := f.multipartRepo.Delete(ctx, path)
	if err != nil {
		logger.Logger.WithName("abort multipart").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, err
	}

	err = f.storages.AbortMultipartUpload(bucket, path, req.UploadID)
	if err != nil {
		logger.Logger.WithName("abort multipart").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, error2.New(code.InvalidDelFile)
	}

	return &AbortMultipartUploadResp{}, nil
}

// FinishReq FinishReq.
type FinishReq struct {
	Path string `json:"path" binding:"required"`
}

// FinishResp FinishResp.
type FinishResp struct{}

func (f *fileserver) Finish(ctx context.Context, req *FinishReq) (*FinishResp, error) {
	bucket, path := utils.Split(req.Path, "/")
	if !utils.ExistBucket(f.conf.Buckets, bucket) {
		logger.Logger.WithName("finish").Infow("invalid storage", header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, error2.New(code.InvalidStorage)
	}
	info, err := f.fileServerRepo.GetByPath(f.db, path)
	if err != nil {
		logger.Logger.WithName("finish").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, err
	}
	if info != nil {
		return &FinishResp{}, nil
	}

	newInfo := &models.FileServer{
		ID:       id2.StringUUID(),
		Path:     path,
		CreateAt: time2.NowUnix(),
		UpdateAt: time2.NowUnix(),
	}

	tx := f.db.Begin()
	err = f.fileServerRepo.Create(tx, newInfo)
	if err != nil {
		tx.Rollback()
		logger.Logger.WithName("finish").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, err
	}
	tx.Commit()

	return &FinishResp{}, nil
}
