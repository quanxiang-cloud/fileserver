package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/quanxiang-cloud/cabin/logger"
	"github.com/quanxiang-cloud/fileserver/api/restful"
	"github.com/quanxiang-cloud/fileserver/pkg/misc/config"
)

var (
	configPath      string
	accessKeyID     string
	secretAccessKey string
	endpoint        string
	region          string
	urlExpire       time.Duration
	partExpire      time.Duration
)

func main() {
	flag.StringVar(&configPath, "config", "../configs/config.yml", "config file path")
	flag.StringVar(&accessKeyID, "accesskey", "", "access key id")
	flag.StringVar(&secretAccessKey, "secretkey", "", "secret access key")
	flag.StringVar(&endpoint, "endpoint", "", "endpoint")
	flag.StringVar(&region, "region", "", "region")
	flag.DurationVar(&urlExpire, "urlExpire", 10*time.Minute, "url expire")
	flag.DurationVar(&partExpire, "partExpire", 24*time.Hour, "part expire")
	flag.Parse()
	conf, err := config.NewConfig(configPath)
	if err != nil {
		panic(err)
	}

	conf.Storage = config.Storage{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		Endpoint:        endpoint,
		Region:          region,
		URLExpire:       urlExpire,
		PartExpire:      partExpire,
	}

	logger.Logger = logger.New(&conf.Log)
	if err != nil {
		panic(err)
	}

	// start restful
	router, err := restful.NewRouter(conf)
	if err != nil {
		panic(err)
	}
	go router.Run()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			router.Close()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
