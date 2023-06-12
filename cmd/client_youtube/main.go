package main

import (
	"context"
	"flag"
	"github.com/ChineseSubFinder/csf-supplier-base/pkg"
	"github.com/VideoTranslationTools/client/pkg/settings"
	"github.com/WQGroup/logger"
	"github.com/allanpk716/conf"
	"github.com/allanpk716/rod_helper"
	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
	"github.com/wader/goutubedl"
	"io"
	"os"
	"path/filepath"
	"time"
)

var configFile = flag.String("f", "etc/client_youtube.yaml", "the config file")

type progressWriter struct {
	writer     io.Writer
	bar        *progressbar.ProgressBar
	downloaded int64
	total      int64
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.downloaded += int64(n)
	pw.bar.Set64(pw.downloaded)
	_, err := pw.writer.Write(p)
	return n, err
}

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
		logger.Fatalln("rod_helper.NewHttpClient", err)
	}

	logger.Infoln("Try Download Video From", dlUrl)

	logger.Infoln("New Downloader ...")
	goutubedl.Path = c.YTdlpFilePath
	gOpt := goutubedl.Options{
		Type:              goutubedl.TypeSingle, // 暂时不支持视频列表下载
		HTTPClient:        client.GetClient(),
		DebugLog:          logger.GetLogger(),
		DownloadSubtitles: false,
	}
	if c.ProxyType != "no" {
		// 设置代理
		gOpt.ProxyUrl = c.ProxyUrl
	}

	logger.Infoln("Get Download Info ...")
	result, err := goutubedl.New(context.Background(), dlUrl, gOpt)
	if err != nil {
		logger.Fatalln("goutubedl.New", err)
	}

	nowCacheRootFolder := filepath.Join(c.CacheRootFolder, pkg.ReplaceWindowsSpecString(result.Info.Title, "-"))

	logger.Infoln("Title:", result.Info.Title)
	logger.Infoln("Subtitles Count:", len(result.Info.Subtitles))

	logger.Infoln("Get Download Info ...")
	downloadResult, err := result.Download(context.Background(), "best")
	if err != nil {
		logger.Fatalln("result.Download", err)
	}
	defer downloadResult.Close()

	fileSize := int64(result.Info.Filesize)
	if fileSize <= 0 {
		logger.Warningln("Get Video Filesize == 0, So maybe the file size is not correct, try FilesizeApprox ...")
		logger.Infoln("FilesizeApprox:", result.Info.FilesizeApprox)
		fileSize = int64(result.Info.FilesizeApprox)
		if fileSize <= 0 {
			logger.Fatal("Invalid file size")
		}
	}

	progress := progressbar.NewOptions(
		int(fileSize),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(30),
		progressbar.OptionSetDescription("Downloading"),
	)

	logger.Infoln("Save to cache folder", nowCacheRootFolder)
	if pkg.IsDir(nowCacheRootFolder) == false {
		err = os.MkdirAll(nowCacheRootFolder, os.ModePerm)
		if err != nil {
			logger.Fatalln("os.MkdirAll", err)
		}
	}
	f, err := os.Create(filepath.Join(nowCacheRootFolder, "downloaded_video.mp4"))
	if err != nil {
		logger.Fatalln("os.Create", err)
	}
	defer f.Close()

	pw := &progressWriter{
		writer:     f,
		bar:        progress,
		downloaded: 0,
		total:      fileSize,
	}

	logger.Infoln("Wait for Downloading ...")

	_, err = io.Copy(pw, downloadResult)
	if err != nil && err != io.EOF {
		logger.Fatalln("io.Copy", err)
	}

	err = progress.Finish()
	if err != nil {
		logger.Fatalln("progress.Finish", err)
	}

	err = progress.Finish()
	if err != nil {
		logger.Fatalln("progress.Finish", err)
	}

	logger.Infoln("Done")
}
