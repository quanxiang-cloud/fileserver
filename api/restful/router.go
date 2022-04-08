package restful

import (
	"github.com/quanxiang-cloud/cabin/logger"
	cabinGin "github.com/quanxiang-cloud/cabin/tailormade/gin"
	"github.com/quanxiang-cloud/fileserver/pkg/misc/config"
	"github.com/quanxiang-cloud/fileserver/pkg/probe"

	"github.com/gin-gonic/gin"
)

const (
	// DebugMode indicates mode is debug.
	DebugMode = "debug"
	// ReleaseMode indicates mode is release.
	ReleaseMode = "release"
)

const (
	signPath = "sign"
	basePath = "base"
)

// Router route.
type Router struct {
	*probe.Probe
	c *config.Config

	e *gin.Engine
}

type router func(c *config.Config, r map[string]*gin.RouterGroup) error

var routers = []router{
	fileserverRouter,
}

// NewRouter create a new router.
func NewRouter(c *config.Config) (*Router, error) {
	e, err := newRouter(c)
	if err != nil {
		return nil, err
	}
	routerGroup := map[string]*gin.RouterGroup{
		basePath: e.Group("/api/v1"),
		signPath: e.Group("/api/v1/fileserver"),
	}

	for _, f := range routers {
		err = f(c, routerGroup)
		if err != nil {
			return nil, err
		}
	}

	probe := probe.New(logger.Logger)
	router := &Router{
		c:     c,
		e:     e,
		Probe: probe,
	}
	router.probe()

	return router, nil
}

func newRouter(c *config.Config) (*gin.Engine, error) {
	if c.Model == "" || (c.Model != ReleaseMode && c.Model != DebugMode) {
		c.Model = ReleaseMode
	}

	gin.SetMode(c.Model)
	engine := gin.New()
	engine.Use(cabinGin.LoggerFunc(), cabinGin.RecoveryFunc())

	return engine, nil
}

func fileserverRouter(c *config.Config, r map[string]*gin.RouterGroup) error {
	fileserver, err := NewFileServer(c)
	if err != nil {
		return err
	}

	base := r[basePath].Group("/fileserver")
	{
		// custom page
		base.POST("/compress", checkSize(c.MaxSize), fileserver.Compress)
		base.POST("/blob/:appID/:md5/*fileName", fileserver.Blob)

		base.POST("/del", fileserver.DelFile)
		base.POST("/thumbnail", fileserver.Thumbnail)
		base.POST("/domain", fileserver.Domain)
	}

	sign := r[signPath].Group("/sign")
	{
		sign.POST("/upload", fileserver.PresignedUpload)
		sign.POST("/download", fileserver.PresignedDownload)
		sign.POST("/uploadMultipart", fileserver.PresignedMultipart)
		sign.POST("/initMultipart", fileserver.InitMultipartUpload)
		sign.POST("/listMultipart", fileserver.ListMultiParts)
		sign.POST("/completeMultipart", fileserver.CompleteMultiParts)
		sign.POST("/abortMultipart", fileserver.AbortMultipartUpload)
		sign.POST("/finish", fileserver.Finish)
	}

	return nil
}

func (r *Router) probe() {
	r.e.GET("liveness", func(c *gin.Context) {
		r.Probe.LivenessProbe(c.Writer, c.Request)
	})

	r.e.Any("readiness", func(c *gin.Context) {
		r.Probe.ReadinessProbe(c.Writer, c.Request)
	})
}

// Run start server.
func (r *Router) Run() {
	r.Probe.SetRunning()
	r.e.Run(r.c.Port)
}

// Close close server.
func (r *Router) Close() {
}
