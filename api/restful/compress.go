package restful

import (
	"net/http"

	"github.com/quanxiang-cloud/cabin/logger"
	"github.com/quanxiang-cloud/cabin/tailormade/header"
	"github.com/quanxiang-cloud/cabin/tailormade/resp"
	"github.com/quanxiang-cloud/fileserver/internal/service"

	"github.com/gin-gonic/gin"
)

// Compress Compress.
func (f *FileServer) Compress(c *gin.Context) {
	ctx := header.MutateContext(c)

	file, err := c.FormFile("file")
	if err != nil {
		logger.Logger.WithName("compress").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)
		resp.Format(nil, err).Context(c, http.StatusInternalServerError)

		return
	}

	resp.Format(f.fileserver.CompressFile(ctx, &service.CompressReq{
		FileHeader: file,
		AppID:      c.PostForm("appID"),
	})).Context(c, http.StatusOK)
}

// Blob Blob.
func (f *FileServer) Blob(c *gin.Context) {
	ctx := header.MutateContext(c)

	req := &service.BoCompressFileReq{}
	if err := c.ShouldBindUri(req); err != nil {
		logger.Logger.WithName("blob").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)
		resp.Format(nil, err).Context(c, http.StatusBadRequest)

		return
	}

	res, err := f.fileserver.BoCompressFile(ctx, req)
	if err != nil {
		resp.Format(nil, err).Context(c, http.StatusNotFound)

		return
	}

	c.DataFromReader(http.StatusOK, res.ContentLength, res.ContentType, res.Buffer, map[string]string{
		"Cache-Control": "public, max-age=31536000",
	})
}
