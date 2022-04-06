package restful

import (
	"net/http"

	"github.com/quanxiang-cloud/cabin/logger"
	"github.com/quanxiang-cloud/cabin/tailormade/header"
	"github.com/quanxiang-cloud/cabin/tailormade/resp"
	"github.com/quanxiang-cloud/fileserver/internal/service"

	"github.com/gin-gonic/gin"
)

// PresignedUpload PresignedUpload.
func (f *FileServer) PresignedUpload(c *gin.Context) {
	ctx := header.MutateContext(c)

	req := &service.PresignedUploadReq{}
	if err := c.ShouldBind(req); err != nil {
		logger.Logger.WithName("presigned upload").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)
		resp.Format(nil, err).Context(c, http.StatusBadRequest)

		return
	}

	resp.Format(f.fileserver.PresignedUpload(ctx, req)).Context(c)
}

// PresignedDownload PresignedDownload.
func (f *FileServer) PresignedDownload(c *gin.Context) {
	ctx := header.MutateContext(c)

	req := &service.PresignedDownloadReq{}
	if err := c.ShouldBind(req); err != nil {
		logger.Logger.WithName("presigned download").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)
		resp.Format(nil, err).Context(c, http.StatusBadRequest)

		return
	}

	resp.Format(f.fileserver.PresignedDownload(ctx, req)).Context(c)
}

// InitMultipartUpload InitMultipartUpload.
func (f *FileServer) InitMultipartUpload(c *gin.Context) {
	ctx := header.MutateContext(c)

	req := &service.InitMultipartUploadReq{}
	if err := c.ShouldBind(req); err != nil {
		logger.Logger.WithName("init multipart upload").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)
		resp.Format(nil, err).Context(c, http.StatusBadRequest)

		return
	}

	resp.Format(f.fileserver.InitMultipartUpload(ctx, req)).Context(c)
}

// PresignedMultipart PresignedMultipart.
func (f *FileServer) PresignedMultipart(c *gin.Context) {
	ctx := header.MutateContext(c)

	req := &service.PresignedMultipartReq{}
	if err := c.ShouldBind(req); err != nil {
		logger.Logger.WithName("presigned multipart").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)
		resp.Format(nil, err).Context(c, http.StatusBadRequest)

		return
	}

	resp.Format(f.fileserver.PresignedMultipart(ctx, req)).Context(c)
}

// ListMultiParts ListMultiParts.
func (f *FileServer) ListMultiParts(c *gin.Context) {
	ctx := header.MutateContext(c)

	req := &service.ListMultiPartsReq{}
	if err := c.ShouldBind(req); err != nil {
		logger.Logger.WithName("list multipart").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)
		resp.Format(nil, err).Context(c, http.StatusBadRequest)

		return
	}

	resp.Format(f.fileserver.ListMultiParts(ctx, req)).Context(c)
}

// CompleteMultiParts CompleteMultiParts.
func (f *FileServer) CompleteMultiParts(c *gin.Context) {
	ctx := header.MutateContext(c)

	req := &service.CompleteMultiPartsReq{}
	if err := c.ShouldBind(req); err != nil {
		logger.Logger.WithName("complete multipart").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)
		resp.Format(nil, err).Context(c, http.StatusBadRequest)

		return
	}

	resp.Format(f.fileserver.CompleteMultiParts(ctx, req)).Context(c)
}

// AbortMultipartUpload AbortMultipartUpload.
func (f *FileServer) AbortMultipartUpload(c *gin.Context) {
	ctx := header.MutateContext(c)

	req := &service.AbortMultipartUploadReq{}
	if err := c.ShouldBind(req); err != nil {
		logger.Logger.WithName("abort multipart upload").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)
		resp.Format(nil, err).Context(c, http.StatusBadRequest)

		return
	}

	resp.Format(f.fileserver.AbortMultipartUpload(ctx, req)).Context(c)
}

// Finish Finish.
func (f *FileServer) Finish(c *gin.Context) {
	ctx := header.MutateContext(c)

	req := &service.FinishReq{}
	if err := c.ShouldBind(req); err != nil {
		logger.Logger.WithName("finish").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)
		resp.Format(nil, err).Context(c, http.StatusBadRequest)

		return
	}

	resp.Format(f.fileserver.Finish(ctx, req)).Context(c)
}
