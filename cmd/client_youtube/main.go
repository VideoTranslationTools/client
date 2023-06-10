package main

import (
	"context"
	"flag"
	"github.com/VideoTranslationTools/client/pkg/settings"
	"github.com/WQGroup/logger"
	"github.com/allanpk716/conf"
	"github.com/allanpk716/rod_helper"
	"github.com/wader/goutubedl"
	"io"
	"os"
	"time"
)

var configFile = flag.String("f", "etc/client_youtube.yaml", "the config file")

func main() {

	dlUrl := "https://www.youtube.com/watch?v=MpYy6wwqxoo&ab_channel=THEFIRSTTAKE"

	flag.Parse()

	var c settings.Configs
	conf.MustLoad(*configFile, &c)
	// 初始化代理设置
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

	goutubedl.Path = c.YTdlpFilePath
	gOpt := goutubedl.Options{
		HTTPClient: client.GetClient(),
	}
	result, err := goutubedl.New(context.Background(), dlUrl, gOpt)
	if err != nil {
		logger.Fatal(err)
	}
	downloadResult, err := result.Download(context.Background(), "best")
	if err != nil {
		logger.Fatal(err)
	}
	defer downloadResult.Close()
	f, err := os.Create("output")
	if err != nil {
		logger.Fatal(err)
	}
	defer f.Close()
	_, err = io.Copy(f, downloadResult)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Infoln("ok")
}
