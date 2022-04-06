package service

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	error2 "github.com/quanxiang-cloud/cabin/error"
	id2 "github.com/quanxiang-cloud/cabin/id"
	"github.com/quanxiang-cloud/cabin/logger"
	"github.com/quanxiang-cloud/cabin/tailormade/header"
	time2 "github.com/quanxiang-cloud/cabin/time"
	"github.com/quanxiang-cloud/fileserver/internal/models"
	"github.com/quanxiang-cloud/fileserver/pkg/decompress"
	"github.com/quanxiang-cloud/fileserver/pkg/mime"
	"github.com/quanxiang-cloud/fileserver/pkg/misc/code"
	"github.com/quanxiang-cloud/fileserver/pkg/storage"
	"github.com/quanxiang-cloud/fileserver/pkg/utils"
)

// CompressReq CompressReq.
type CompressReq struct {
	AppID      string `form:"appID"`
	FileHeader *multipart.FileHeader
}

// CompressResp CompressResp.
type CompressResp struct {
	URL string `json:"url"`
}

func (f *fileserver) CompressFile(ctx context.Context, req *CompressReq) (*CompressResp, error) {
	md5, err := utils.GetMD5ByMultipart(req.FileHeader)
	if err != nil {
		logger.Logger.WithName("compress file").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, error2.New(code.ErrUnarchive)
	}

	ext := utils.GetExt(req.FileHeader.Filename)
	extract, err := f.extract.GetDecompress(ext)
	if err != nil {
		logger.Logger.WithName("compress file").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, error2.New(code.InvalidCompress)
	}

	if extract == nil {
		logger.Logger.WithName("compress file").Infow("extract is nil", header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, error2.New(code.InvalidCompress)
	}

	// Check whether there is index. In the root directory html
	indexPath, err := f.walkIndex(ctx, req.FileHeader, extract)
	if err != nil {
		return nil, err
	}

	// unzip package
	dst := filepath.Join(f.conf.Blob.TempPath, md5)
	err = f.unarchive(ctx, req.FileHeader, extract, dst)
	if err != nil {
		return nil, error2.New(code.ErrUnarchive)
	}
	defer os.RemoveAll(dst)

	path := utils.ExecuteURL(utils.Blob{
		AppID:    req.AppID,
		MD5:      md5,
		FileName: indexPath,
	}, f.conf.Blob.Template)

	f.eg.Go(func() error {
		return f.uploadCompressFile(ctx, dst, req.AppID, md5)
	})

	f.eg.Go(func() error {
		return f.uploadArchive(ctx, req, path, md5)
	})

	err = f.eg.Wait()
	if err != nil {
		logger.Logger.WithName("compress upload").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, error2.New(code.ErrUnarchive)
	}

	return &CompressResp{
		URL: path,
	}, nil
}

// check whether there is an index page under the change directory.
func (f *fileserver) walkIndex(ctx context.Context, fileHeader *multipart.FileHeader, extract decompress.Decompress) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		logger.Logger.WithName("walk index").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)

		return "", error2.New(code.InvalidCompress)
	}
	defer file.Close()

	indexPath, err := extract.Walk(file)
	if err != nil {
		logger.Logger.WithName("walk index").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)

		return "", error2.New(code.InvalidCompress)
	}

	if indexPath == "" {
		logger.Logger.WithName("walk index").Infow("indexPath is empty", header.GetRequestIDKV(ctx).Fuzzy()...)

		return "", error2.New(code.InvalidIndex)
	}

	return indexPath, nil
}

// unzip package.
func (f *fileserver) unarchive(ctx context.Context, fileHeader *multipart.FileHeader, extract decompress.Decompress, dst string) error {
	file, err := fileHeader.Open()
	if err != nil {
		logger.Logger.WithName("unarchive").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)

		return error2.New(code.InvalidCompress)
	}
	defer file.Close()

	err = extract.Unarchive(file, dst)
	if err != nil {
		logger.Logger.WithName("unarchive").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)

		return error2.New(code.ErrUnarchive)
	}

	return nil
}

func (f *fileserver) uploadCompressFile(ctx context.Context, dst, appID, md5 string) error {
	blobTemplate := f.conf.Blob.Template
	blobTemplatePath := f.conf.Blob.TempPath

	bucket := f.conf.Buckets[storage.Private]
	if bucket == "" {
		logger.Logger.WithName("upload compress file").Infow("bucket is empty", header.GetRequestIDKV(ctx).Fuzzy()...)

		return error2.New(code.InvalidStorage)
	}

	return filepath.Walk(dst, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			logger.Logger.WithName("upload compress file").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)

			return err
		}
		if info.IsDir() {
			return nil
		}

		contentType := mime.DetectFilePath(info.Name())
		switch {
		case strings.HasSuffix(info.Name(), ".htm"):
			fallthrough
		case strings.HasSuffix(info.Name(), ".html"):
			obj := utils.Blob{AppID: appID, MD5: md5}

			buf, err := utils.ReplaceAttr(obj, blobTemplate, path, dst)
			if err != nil {
				logger.Logger.WithName("upload compress file").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)

				return err
			}
			path = strings.Replace(path, blobTemplatePath, "", 1)

			return f.storages.PutObject(bucket, filepath.Join(appID, path), bytes.NewReader(buf.Bytes()), contentType)
		default:
			file, err := os.Open(path)
			if err != nil {
				logger.Logger.WithName("upload compress file").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)

				return err
			}
			defer file.Close()

			path = strings.Replace(path, blobTemplatePath, "", 1)

			return f.storages.PutObject(bucket, filepath.Join(appID, path), file, contentType)
		}
	})
}

func (f *fileserver) uploadArchive(ctx context.Context, req *CompressReq, indexPath, md5 string) error {
	file, err := req.FileHeader.Open()
	if err != nil {
		logger.Logger.WithName("upload archive").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)

		return err
	}
	defer file.Close()

	bucket := f.conf.Buckets[storage.Private]
	if bucket == "" {
		logger.Logger.WithName("upload archive").Infow("bucket is empty", header.GetRequestIDKV(ctx).Fuzzy()...)

		return error2.New(code.InvalidStorage)
	}

	path := genArchiveName(indexPath, req.FileHeader.Filename)
	info, err := f.fileServerRepo.GetByPath(f.db, path)
	if err != nil {
		return err
	}
	if info != nil {
		return nil
	}

	contentType := mime.DetectFilePath(req.FileHeader.Filename)
	tx := f.db.Begin()
	newInfo := &models.FileServer{
		ID:       id2.StringUUID(),
		Path:     path,
		CreateAt: time2.NowUnix(),
		UpdateAt: time2.NowUnix(),
	}

	err = f.fileServerRepo.Create(tx, newInfo)
	if err != nil {
		tx.Rollback()

		return err
	}

	err = f.storages.PutObject(bucket, path, file, contentType)
	if err != nil {
		tx.Rollback()
		logger.Logger.WithName("upload archive").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)

		return error2.New(code.ErrUploadFile)
	}

	tx.Commit()

	return nil
}

const (
	blob = "/blob/"
)

func genArchiveName(indexPath, oldArchiveName string) string {
	indexExt := filepath.Ext(indexPath)
	archiveExt := filepath.Ext(oldArchiveName)
	tempName := strings.Replace(indexPath, blob, "", 1)

	return strings.Replace(tempName, indexExt, archiveExt, 1)
}

// BoCompressFileReq BoCompressFileReq.
type BoCompressFileReq struct {
	AppID    string `uri:"appID"`
	MD5      string `uri:"md5"`
	FileName string `uri:"fileName"`
}

// BoCompressFileResp BoCompressFileResp.
type BoCompressFileResp struct {
	ContentType   string
	ContentLength int64
	Buffer        io.Reader
}

func (f *fileserver) BoCompressFile(ctx context.Context, req *BoCompressFileReq) (*BoCompressFileResp, error) {
	bucket := f.conf.Buckets[storage.Private]
	if bucket == "" {
		logger.Logger.WithName("BoCompressFile").Infow("bucket is empty", header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, error2.New(code.InvalidStorage)
	}

	buffer := &bytes.Buffer{}

	path := filepath.Join(req.AppID, req.MD5, req.FileName)
	reader, err := f.storages.GetObject(bucket, path)
	if err != nil {
		logger.Logger.WithName("BoCompressFile").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, error2.New(code.InvalidExist)
	}

	size, err := io.Copy(buffer, reader)
	if err != nil {
		logger.Logger.WithName("BoCompressFile").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)

		return nil, error2.New(code.ErrDownload)
	}

	contentType := mime.DetectFilePath(req.FileName)

	return &BoCompressFileResp{
		ContentType:   contentType,
		ContentLength: size,
		Buffer:        buffer,
	}, nil
}
