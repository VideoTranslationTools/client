package main

import (
	"context"
	"flag"
	"github.com/VideoTranslationTools/client/pkg/settings"
	"github.com/WQGroup/logger"
	"github.com/allanpk716/conf"
	"github.com/allanpk716/rod_helper"
	"github.com/sirupsen/logrus"
	"github.com/wader/goutubedl"
	"io"
	"os"
	"time"
)

var configFile = flag.String("f", "etc/client_youtube.yaml", "the config file")

func main() {

	dlUrl := "https://www.youtube.com/watch?v=MpYy6wwqxoo&ab_channel=THEFIRSTTAKE"

	logger.SetLoggerLevel(logrus.InfoLevel)
	flag.Parse()

	var c settings.Configs
	conf.MustLoad(*configFile, &c)
	// 初始化代理设置
	logger.Infoln("InitFakeUA ...")
	rod_helper.InitFakeUA(true, "", "")
	opt := rod_helper.NewHttpClientOptions(5 * time.Second)
	if c.ProxyType != "no" {
		// 设置代理
		opt.SetHttpProxy(c.ProxyUrl)
	}
	client, err := rod_helper.NewHttpClient(opt)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Infoln("Try Download Video From", dlUrl)

	logger.Infoln("New Downloader ...")
	goutubedl.Path = c.YTdlpFilePath
	gOpt := goutubedl.Options{
		HTTPClient: client.GetClient(),
		DebugLog:   logger.GetLogger(),
	}
	result, err := goutubedl.New(context.Background(), dlUrl, gOpt)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Infoln("Get Download Info ...")
	downloadResult, err := result.Download(context.Background(), "best")
	if err != nil {
		logger.Fatal(err)
	}
	defer downloadResult.Close()

	logger.Infoln("Save to cache file ...")
	f, err := os.Create("output")
	if err != nil {
		logger.Fatal(err)
	}
	defer f.Close()

	logger.Infoln("Downloading ...")
	_, err = io.Copy(f, downloadResult)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Infoln("Done")
}
