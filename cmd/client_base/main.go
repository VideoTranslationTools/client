package main

import (
	"flag"
	"github.com/VideoTranslationTools/base/pkg/task_system"
	npkg "github.com/VideoTranslationTools/client/pkg"
	"github.com/VideoTranslationTools/client/pkg/machine_translation_helper"
	"github.com/VideoTranslationTools/client/pkg/settings"
	"github.com/WQGroup/logger"
	"github.com/allanpk716/conf"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
)

var configFile = flag.String("f", "etc/client_base.yaml", "the config file")

func main() {

	videoFPath := flag.String("video", "", "the video file path")
	videoLang := flag.String("v_lang", "", "the video language")
	cancelThisTaskPackage := flag.Bool("cancel", false, "cancel this task package")

	logger.SetLoggerLevel(logrus.InfoLevel)

	flag.Parse()

	if *videoLang == "" {
		logger.Infoln("v_lang is empty, will auto detect language")
	}

	var c settings.Configs
	conf.MustLoad(*configFile, &c)

	// 然后调用 FFMPEG 进行音频的导出
	ffmpegInfo := npkg.ExportAudioFile(c.CacheRootFolder, *videoFPath)
	// 正常来说只会有一个音频，然后还是需要用户再传入 URL 的时候指定这个视频的语言，这里就不做判断了（因为不准）
	if len(ffmpegInfo.AudioInfoList) <= 0 {
		logger.Fatalln("ffmpegInfo.AudioInfoList <= 0")
	}
	// 获取这个 youtubeVideoFPath 视频文件的文件名称，不包含后缀名
	videoTitle := strings.TrimSuffix(filepath.Base(*videoFPath), filepath.Ext(*videoFPath))
	// 获取这个 youtubeVideoFPath 视频文件的文件夹目录
	videoRootFolder := filepath.Dir(*videoFPath)
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
