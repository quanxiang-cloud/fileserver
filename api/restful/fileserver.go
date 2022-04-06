package restful

import (
	"net/http"

	"github.com/quanxiang-cloud/cabin/logger"
	"github.com/quanxiang-cloud/cabin/tailormade/header"
	"github.com/quanxiang-cloud/cabin/tailormade/resp"
	"github.com/quanxiang-cloud/fileserver/internal/service"
	"github.com/quanxiang-cloud/fileserver/pkg/misc/config"

	"github.com/gin-gonic/gin"
)

// FileServer file service.
type FileServer struct {
	fileserver service.FileServer
}

// NewFileServer new a fileserver.
func NewFileServer(conf *config.Config) (*FileServer, error) {
	fileserver, err := service.NewFileServer(conf)
	if err != nil {
		return nil, err
	}

	return &FileServer{
		fileserver: fileserver,
	}, nil
}

// DelFile delete file.
func (f *FileServer) DelFile(c *gin.Context) {
	ctx := header.MutateContext(c)

	req := &service.DelUploadFileReq{}
	if err := c.ShouldBind(req); err != nil {
		logger.Logger.WithName("delete").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)
		resp.Format(nil, err).Context(c, http.StatusBadRequest)

		return
	}

	resp.Format(f.fileserver.DelUploadFile(ctx, req)).Context(c)
}

// Thumbnail Thumbnail.
func (f *FileServer) Thumbnail(c *gin.Context) {
	ctx := header.MutateContext(c)

	req := &service.ThumbnailReq{}
	if err := c.ShouldBind(req); err != nil {
		logger.Logger.WithName("thumbnail").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)
		resp.Format(nil, err).Context(c, http.StatusBadRequest)

		return
	}

	resp.Format(f.fileserver.Thumbnail(ctx, req)).Context(c)
}

// Domain Domain.
func (f *FileServer) Domain(c *gin.Context) {
	ctx := header.MutateContext(c)

	req := &service.DomainReq{}
	if err := c.ShouldBind(req); err != nil {
		logger.Logger.WithName("domain").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)
		resp.Format(nil, err).Context(c, http.StatusBadRequest)

		return
	}

	resp.Format(f.fileserver.Domain(ctx, req)).Context(c)
}
