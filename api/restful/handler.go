package restful

import (
	"net/http"

	error2 "github.com/quanxiang-cloud/cabin/error"
	"github.com/quanxiang-cloud/cabin/logger"
	"github.com/quanxiang-cloud/cabin/tailormade/header"
	"github.com/quanxiang-cloud/cabin/tailormade/resp"
	"github.com/quanxiang-cloud/fileserver/pkg/misc/code"

	"github.com/gin-gonic/gin"
)

// checkSize check upload file stream size.
func checkSize(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := header.MutateContext(c)

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		if err := c.Request.ParseMultipartForm(maxSize); err != nil {
			logger.Logger.WithName("check compress size").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)
			c.Abort()
			resp.Format(nil, error2.New(code.ErrFileLimit)).Context(c, http.StatusRequestEntityTooLarge)

			return
		}

		c.Next()
	}
}
