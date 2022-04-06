package service

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"

	error2 "github.com/quanxiang-cloud/cabin/error"
	id2 "github.com/quanxiang-cloud/cabin/id"
	"github.com/quanxiang-cloud/cabin/logger"
	mysql2 "github.com/quanxiang-cloud/cabin/tailormade/db/mysql"
	redis2 "github.com/quanxiang-cloud/cabin/tailormade/db/redis"
	"github.com/quanxiang-cloud/cabin/tailormade/header"
	time2 "github.com/quanxiang-cloud/cabin/time"
	"github.com/quanxiang-cloud/fileserver/internal/models"
	repo "github.com/quanxiang-cloud/fileserver/internal/models/mysql"
	"github.com/quanxiang-cloud/fileserver/internal/models/redis"
	"github.com/quanxiang-cloud/fileserver/pkg/decompress"
	"github.com/quanxiang-cloud/fileserver/pkg/mime"
	"github.com/quanxiang-cloud/fileserver/pkg/misc/code"
	"github.com/quanxiang-cloud/fileserver/pkg/misc/config"
	"github.com/quanxiang-cloud/fileserver/pkg/storage"
	"github.com/quanxiang-cloud/fileserver/pkg/utils"

	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

// FileServer file service interface.
type FileServer interface {
	DelUploadFile(ctx context.Context, req *DelUploadFileReq) (*DelUploadFileResp, error)
	Thumbnail(ctx context.Context, req *ThumbnailReq) (*ThumbnailResp, error)
	Domain(ctx context.Context, req *DomainReq) (*DomainResp, error)
	CompressFile(ctx context.Context, req *CompressReq) (*CompressResp, error)
	BoCompressFile(ctx context.Context, req *BoCompressFileReq) (*BoCompressFileResp, error)
	PresignedUpload(ctx context.Context, req *PresignedUploadReq) (*PresignedUploadResp, error)
	PresignedDownload(ctx context.Context, req *PresignedDownloadReq) (*PresignedDownloadResp, error)
	InitMultipartUpload(ctx context.Context, req *InitMultipartUploadReq) (*InitMultipartUploadResp, error)
	PresignedMultipart(ctx context.Context, req *PresignedMultipartReq) (*PresignedMultipartResp, error)
	ListMultiParts(ctx context.Context, req *ListMultiPartsReq) (*ListMultiPartsResp, error)
	CompleteMultiParts(ctx context.Context, req *CompleteMultiPartsReq) (*CompleteMultiPartsResp, error)
	AbortMultipartUpload(ctx context.Context, req *AbortMultipartUploadReq) (*AbortMultipartUploadResp, error)
	Finish(ctx context.Context, req *FinishReq) (*FinishResp, error)
}

type fileserver struct {
	db   *gorm.DB
	conf *config.Config

	storages       *storage.Storage
	extract        *decompress.Decompressor
	fileServerRepo models.FileServerRepo
	multipartRepo  models.MultipartRepo
	eg             *errgroup.Group
}

// NewFileServer new fileserver.
func NewFileServer(conf *config.Config) (FileServer, error) {
	c := conf.Mysql
	c.SetDSN(mysql2.DSN_UTF8MB4)
	db, err := mysql2.New(c, logger.Logger)
	if err != nil {
		return nil, err
	}
	redisClient, err := redis2.NewClient(conf.Redis)
	if err != nil {
		return nil, err
	}

	storages, err := storage.New(conf.Storage)
	if err != nil {
		return nil, err
	}

	f := &fileserver{
		db:             db,
		conf:           conf,
		extract:        decompress.NewDecompressor(),
		storages:       storages,
		fileServerRepo: repo.NewFileServerRepo(),
		multipartRepo:  redis.NewMultipartRepo(redisClient),
		eg:             &errgroup.Group{},
	}

	return f, nil
}

// DelUploadFileReq DelUploadFileReq.
type DelUploadFileReq struct {
	Path string `json:"path" binding:"required"`
}

// DelUploadFileResp DelUploadFileResp.
type DelUploadFileResp struct{}

func (f *fileserver) DelUploadFile(ctx context.Context, req *DelUploadFileReq) (*DelUploadFileResp, error) {
	bucket, path := utils.Split(req.Path, "/")

	if !utils.ExistBucket(f.conf.Buckets, bucket) {
		logger.Logger.WithName("delete file").Infow("invalid storage", header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, error2.New(code.InvalidStorage)
	}

	info, err := f.fileServerRepo.GetByPath(f.db, path)
	if err != nil {
		logger.Logger.WithName("delete file").Errorw("get file info failed", header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, err
	}

	if info == nil {
		return &DelUploadFileResp{}, nil
	}

	tx := f.db.Begin()
	err = f.fileServerRepo.Delete(tx, info.ID)
	if err != nil {
		tx.Rollback()
		logger.Logger.WithName("delete file").Errorw("delete file info failed", header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, err
	}

	err = f.storages.DeleteObject(bucket, path)
	if err != nil {
		tx.Rollback()
		logger.Logger.WithName("delete file").Errorw("delete file object failed", header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, error2.New(code.InvalidDelFile)
	}

	tx.Commit()

	return &DelUploadFileResp{}, nil
}

// ThumbnailReq ThumbnailReq.
type ThumbnailReq struct {
	Path  string `json:"path" binding:"required"`
	Width int    `json:"width"`
	Hight int    `json:"hight"`
}

// ThumbnailResp ThumbnailResp.
type ThumbnailResp struct{}

func (f *fileserver) Thumbnail(ctx context.Context, req *ThumbnailReq) (*ThumbnailResp, error) {
	bucket, path := utils.Split(req.Path, "/")

	if !utils.ExistBucket(f.conf.Buckets, bucket) {
		logger.Logger.WithName("thumbnail").Infow("invalid storage", header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, error2.New(code.InvalidStorage)
	}

	info, err := f.fileServerRepo.GetByPath(f.db, path)
	if err != nil {
		logger.Logger.WithName("thumbnail").Errorw("get file info failed", header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, err
	}

	if info == nil {
		logger.Logger.WithName("thumbnail").Infow("file not found", header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, error2.New(code.InvalidExist)
	}

	// thumbnail suffix
	dir, file := filepath.Split(path)
	middle := fmt.Sprintf("%dx%d", req.Width, req.Hight)
	thumbnailPath := filepath.Join(dir, middle, file)

	thumbnailInfo, err := f.fileServerRepo.GetByPath(f.db, thumbnailPath)
	if err != nil {
		logger.Logger.WithName("thumbnail").Errorw("get thumbnail info failed", header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, err
	}

	if thumbnailInfo != nil {
		return &ThumbnailResp{}, nil
	}

	reader, err := f.storages.GetObject(bucket, path)
	if err != nil {
		logger.Logger.WithName("thumbnail").Errorw("get file object failed", header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, error2.New(code.InvalidExist)
	}

	out := &bytes.Buffer{}
	err = utils.Scale(reader, out, req.Width, req.Hight, 100)
	if err != nil {
		logger.Logger.WithName("thumbnail").Errorw("scale image failed", header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, error2.New(code.ErrThumbnail)
	}

	tx := f.db.Begin()
	contentType := mime.DetectFilePath(thumbnailPath)

	err = f.storages.PutObject(bucket, thumbnailPath, bytes.NewReader(out.Bytes()), contentType)
	if err != nil {
		tx.Rollback()
		logger.Logger.WithName("thumbnail").Errorw("upload thumbnail object failed", header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, error2.New(code.ErrThumbnail)
	}

	err = f.fileServerRepo.Create(tx, &models.FileServer{
		ID:       id2.StringUUID(),
		Path:     thumbnailPath,
		CreateAt: time2.NowUnix(),
		UpdateAt: time2.NowUnix(),
	})
	if err != nil {
		tx.Rollback()
		logger.Logger.WithName("thumbnail").Errorw("create thumbnail info failed", header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, err
	}

	tx.Commit()

	return &ThumbnailResp{}, nil
}

// DomainReq DomainReq.
type DomainReq struct{}

// DomainResp DomainResp.
type DomainResp struct {
	Domain   string `json:"domain"`
	Private  string `json:"private"`
	Readable string `json:"readable"`
}

func (f *fileserver) Domain(ctx context.Context, req *DomainReq) (*DomainResp, error) {
	_, domain := utils.Split(f.conf.Storage.Endpoint, "://")
	readable := f.conf.Buckets[storage.Readable]
	private := f.conf.Buckets[storage.Private]

	return &DomainResp{
		Domain:   domain,
		Private:  private,
		Readable: readable,
	}, nil
}
