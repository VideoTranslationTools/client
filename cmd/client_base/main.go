package main

import (
	"flag"
	"github.com/ChineseSubFinder/csf-supplier-base/pkg"
	"github.com/VideoTranslationTools/base/pkg/task_system"
	npkg "github.com/VideoTranslationTools/client/pkg"
	"github.com/VideoTranslationTools/client/pkg/machine_translation_helper"
	"github.com/VideoTranslationTools/client/pkg/settings"
	"github.com/VideoTranslationTools/client/pkg/youtube"
	"github.com/WQGroup/logger"
	"github.com/allanpk716/conf"
	"github.com/allanpk716/rod_helper"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
	"time"
)

var configFile = flag.String("f", "etc/client_base.yaml", "the config file")

func init() {
	logger.Infoln("Version:", AppVersion)
}

func main() {

	// 要么是视频文件，要么是 youtube 的 URL
	videoFPath := flag.String("video", "", "the video file path")
	subtitleFPath := flag.String("subtitle", "", "the subtitle file path")
	dlUrl := flag.String("url", "", "the youtube video url")
	// ------------------------------------------------------------------------------
	videoLang := flag.String("lang", "", "the video language")
	cancelThisTaskPackage := flag.Bool("cancel", false, "cancel this task package")
	// ------------------------------------------------------------------------------
	logger.SetLoggerLevel(logrus.InfoLevel)

	flag.Parse()

	if *videoFPath == "" && *dlUrl == "" && *subtitleFPath == "" {
		logger.Fatalln("video and url and subtitle can't all empty")
	}
	if *videoFPath != "" && *dlUrl != "" {
		logger.Fatalln("video and url can't both exist")
	} else if *videoFPath != "" && *subtitleFPath != "" {
		logger.Fatalln("video and subtitle can't both exist")
	} else if *dlUrl != "" && *subtitleFPath != "" {
		logger.Fatalln("url and subtitle can't both exist")
	}
	// 如果没有指定视频的语言，那么就会自动检测语言
	if *subtitleFPath == "" && *videoLang == "" {
		logger.Infoln("lang is empty, will auto detect language")
	}

	var c settings.Configs
	conf.MustLoad(*configFile, &c)

	if *videoFPath != "" {
		// 本地视频的处理逻辑
		processLocalVideo(c, *videoFPath, *videoLang, *cancelThisTaskPackage)
	} else if *subtitleFPath != "" {
		// 本地字幕的处理逻辑
		processLocalSubtitle(c, *subtitleFPath, *cancelThisTaskPackage)
	} else if *dlUrl != "" {
		// youtube 视频的处理逻辑
		processYoutubeVideo(c, *dlUrl, *videoLang, *cancelThisTaskPackage)
	}

	logger.Infoln("Done")
}

func processLocalSubtitle(c settings.Configs, subtitleFPath string, cancelThisTaskPackage bool) {

	logger.Infoln("Will process local srt file ...")

	if pkg.IsFile(subtitleFPath) == false {
		logger.Fatalln("subtitle file is not exist")
	}

	// 获取这个 subtitleFPath 文件的文件名称，不包含后缀名
	subtitleTitle := strings.TrimSuffix(filepath.Base(subtitleFPath), filepath.Ext(subtitleFPath))
	// 获取这个 subtitleFPath 视频文件的文件夹目录
	subtitleRootFolder := filepath.Dir(subtitleFPath)
	// 实例化一个 TaskSystemClient
	tc := task_system.NewTaskSystemClient(c.ServerBaseUrl, c.ApiKey)
	mth := machine_translation_helper.NewMachineTranslationHelper(tc)

	mth.Process(machine_translation_helper.Opts{
		CancelThisTaskPackage:        cancelThisTaskPackage,
		InputFPath:                   subtitleFPath,
		IsAudioOrSRT:                 false,
		AudioLang:                    "",
		TargetTranslationLang:        TargetTranslationLang,
		DownloadedFileSaveFolderPath: subtitleRootFolder,
		TranslatedSRTFileName:        subtitleTitle + Translated + SrtExt,
	})
}

func processLocalVideo(c settings.Configs, videoFPath, videoLang string, cancelThisTaskPackage bool) {

	logger.Infoln("Will process local video file ...")

	if pkg.IsFile(videoFPath) == false {
		logger.Fatalln("video file is not exist")
	}

	// 然后调用 FFMPEG 进行音频的导出
	ffmpegInfo := npkg.ExportAudioFile(c.CacheRootFolder, videoFPath)
	// 正常来说只会有一个音频，然后还是需要用户再传入 URL 的时候指定这个视频的语言，这里就不做判断了（因为不准）
	if len(ffmpegInfo.AudioInfoList) <= 0 {
		logger.Fatalln("ffmpegInfo.AudioInfoList <= 0")
	}
	// 获取这个 videoTitle 视频文件的文件名称，不包含后缀名
	videoTitle := strings.TrimSuffix(filepath.Base(videoFPath), filepath.Ext(videoFPath))
	// 获取这个 videoTitle 视频文件的文件夹目录
	videoRootFolder := filepath.Dir(videoFPath)
	// 实例化一个 TaskSystemClient
	tc := task_system.NewTaskSystemClient(c.ServerBaseUrl, c.ApiKey)
	mth := machine_translation_helper.NewMachineTranslationHelper(tc)

	mth.Process(machine_translation_helper.Opts{
		CancelThisTaskPackage:        cancelThisTaskPackage,
		InputFPath:                   ffmpegInfo.AudioInfoList[0].FullPath,
		IsAudioOrSRT:                 true,
		AudioLang:                    videoLang,
		TargetTranslationLang:        TargetTranslationLang,
		DownloadedFileSaveFolderPath: videoRootFolder,
		TranslatedSRTFileName:        videoTitle + Translated + SrtExt,
	})
}

func processYoutubeVideo(c settings.Configs, dlUrl, videoLang string, cancelThisTaskPackage bool) {

	logger.Infoln("Will process online youtube video ...")

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
	youtubeVideoFPath := youtube.DownloadVideo(c, client, dlUrl)
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
		CancelThisTaskPackage:        cancelThisTaskPackage,
		InputFPath:                   ffmpegInfo.AudioInfoList[0].FullPath,
		IsAudioOrSRT:                 true,
		AudioLang:                    videoLang,
		TargetTranslationLang:        TargetTranslationLang,
		DownloadedFileSaveFolderPath: videoRootFolder,
		TranslatedSRTFileName:        videoTitle + Translated + SrtExt,
	})
}

var AppVersion = "unknow"

const (
	TargetTranslationLang = "CN"
	SrtExt                = ".srt"
	Translated            = "_translated"
)
