package main

import (
	"context"
	"flag"
	"github.com/ChineseSubFinder/csf-supplier-base/pkg"
	"github.com/VideoTranslationTools/base/pkg/task_system"
	npkg "github.com/VideoTranslationTools/client/pkg"
	"github.com/VideoTranslationTools/client/pkg/machine_translation_helper"
	"github.com/VideoTranslationTools/client/pkg/settings"
	"github.com/WQGroup/logger"
	"github.com/allanpk716/conf"
	"github.com/allanpk716/rod_helper"
	"github.com/go-resty/resty/v2"
	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
	"github.com/wader/goutubedl"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var configFile = flag.String("f", "etc/client_youtube.yaml", "the config file")

func main() {

	dlUrl := flag.String("yt_url", "", "the youtube video url")
	videoLang := flag.String("yt_lang", "", "the youtube video language")
	cancelThisTaskPackage := flag.Bool("cancel", false, "cancel this task package")

	logger.SetLoggerLevel(logrus.InfoLevel)

	flag.Parse()

	if *dlUrl == "" {
		logger.Fatalln("yt_url is empty")
	}

	if *videoLang == "" {
		logger.Infoln("yt_lang is empty, will auto detect language")
	}

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
	// 先下载 youtube 的视频
	youtubeVideoFPath := downloadYoutubeVideo(c, client, *dlUrl)
	// 然后调用 FFMPEG 进行音频的导出
	ffmpegInfo := npkg.ExportAudioFile(c.CacheRootFolder, youtubeVideoFPath)
	// 正常来说只会有一个音频，然后还是需要用户再传入 URL 的时候指定这个视频的语言，这里就不做判断了（因为不准）
	if len(ffmpegInfo.AudioInfoList) <= 0 {
		logger.Fatalln("ffmpegInfo.AudioInfoList <= 0")
	}
	// 获取这个 youtubeVideoFPath 视频文件的文件名称，不包含后缀名
	videoTitle := strings.TrimSuffix(filepath.Base(youtubeVideoFPath), filepath.Ext(youtubeVideoFPath))
	// 获取这个 youtubeVideoFPath 视频文件的文件夹目录
	videoRootFolder := filepath.Dir(youtubeVideoFPath)
	// 实例化一个 TaskSystemClient
	tc := task_system.NewTaskSystemClient(c.ServerBaseUrl, c.ApiKey)
	mth := machine_translation_helper.NewMachineTranslationHelper(tc)

	mth.Process(machine_translation_helper.Opts{
		CancelThisTaskPackage:        *cancelThisTaskPackage,
		InputFPath:                   ffmpegInfo.AudioInfoList[0].FullPath,
		IsAudioOrSRT:                 true,
		AudioLang:                    *videoLang,
		TargetTranslationLang:        "CN",
		DownloadedFileSaveFolderPath: videoRootFolder,
		TranslatedSRTFileName:        videoTitle + ".srt",
	})

	logger.Infoln("Done")
}

const (
	videoDownloaded = "downloaded"
)

// downloadYoutubeVideo 下载 Youtube 视频
func downloadYoutubeVideo(c settings.Configs, client *resty.Client, dlUrl string) string {

	logger.Infoln("Try Download Video From", dlUrl)

	logger.Infoln("New Downloader ...")
	goutubedl.Path = c.YTdlpFilePath
	gOpt := goutubedl.Options{
		Type:              goutubedl.TypeSingle, // 暂时不支持视频列表下载
		Filter:            "best",               // 下载的视频质量
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

	videoTitle := pkg.ReplaceWindowsSpecString(result.Info.Title, "-")
	nowCacheRootFolder := filepath.Join(c.CacheRootFolder, videoTitle)

	logger.Infoln("Title:", result.Info.Title)
	logger.Infoln("Subtitles Count:", len(result.Info.Subtitles))

	fileSize := int64(result.Info.Filesize)
	if fileSize <= 0 {
		logger.Warningln("Get Video Filesize == 0, So maybe the file size is not correct, try FilesizeApprox ...")
		logger.Infoln("FilesizeApprox:", result.Info.FilesizeApprox)
		fileSize = int64(result.Info.FilesizeApprox)
		if fileSize <= 0 {
			logger.Fatal("Invalid file size")
		}
	}

	logger.Infoln("Save to cache folder:", nowCacheRootFolder)
	if pkg.IsDir(nowCacheRootFolder) == false {
		err = os.MkdirAll(nowCacheRootFolder, os.ModePerm)
		if err != nil {
			logger.Fatalln("os.MkdirAll", err)
		}
	}
	outVideoFPath := filepath.Join(nowCacheRootFolder, videoTitle+".mp4")
	// 判断下载的目标目录下是否已经有的文件的大小于准备下载的文件大小是一样大的，如果是，则跳过，否则继续下载
	if pkg.IsFile(outVideoFPath) == true {
		logger.Infoln("Target Video File Exist:", outVideoFPath)
		// 获取文件大小
		fileInfo, err := os.Stat(outVideoFPath)
		if err != nil {
			logger.Fatalln("os.Stat", err)
		}
		if fileInfo.Size() == fileSize {
			// 文件大小一样，跳过下载
			logger.Infoln("Target Video File Already Downloaded:", outVideoFPath)
			return outVideoFPath
		} else {

			if pkg.IsFile(filepath.Join(nowCacheRootFolder, videoDownloaded)) == true {
				// 如果视频文件存在，且下载完成的标志位文件也存在，则认为下载完成了，无需再次下载
				logger.Infoln("Target Video File Already Downloaded:", outVideoFPath)
				return outVideoFPath
			}
			logger.Infoln("Target Video File Size Not Match, Delete it:", outVideoFPath)
			err = os.Remove(outVideoFPath)
			if err != nil {
				logger.Fatalln("os.Remove", err)
			}
		}
	}

	logger.Infoln("Get Download Info ...")
	downloadResult, err := result.Download(context.Background(), "")
	if err != nil {
		logger.Fatalln("result.Download", err)
	}
	defer downloadResult.Close()

	progress := progressbar.NewOptions(
		int(fileSize),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(30),
		progressbar.OptionSetDescription("Downloading"),
	)

	f, err := os.Create(outVideoFPath)
	if err != nil {
		logger.Fatalln("os.Create", err)
	}
	defer f.Close()

	pw := npkg.NewProgressWriter(f, progress, fileSize)

	logger.Infoln("Wait for Downloading ...")

	_, err = io.Copy(pw, downloadResult)
	if err != nil && err != io.EOF {
		logger.Fatalln("io.Copy", err)
	}

	// 写一个标志位文件到当前的目录下，表示已经下载完成
	f, err = os.Create(filepath.Join(nowCacheRootFolder, videoDownloaded))
	if err != nil {
		logger.Fatalln("os.Create", err)
	}
	defer f.Close()

	err = progress.Finish()
	if err != nil {
		logger.Fatalln("progress.Finish", err)
	}

	logger.Infoln("Download Video at:", outVideoFPath)

	return outVideoFPath
}
