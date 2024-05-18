package main

import (
	"flag"
	"github.com/ChineseSubFinder/csf-supplier-base/pkg"
	npkg "github.com/VideoTranslationTools/client/pkg"
	"github.com/WQGroup/logger"
	"github.com/sirupsen/logrus"
)

func init() {
	logger.Infoln("Version:", AppVersion)
}

func main() {

	videoFPath := flag.String("video", "", "the video file path")

	// ------------------------------------------------------------------------------
	logger.SetLoggerLevel(logrus.InfoLevel)

	flag.Parse()

	logger.Infoln("Will export video subtitle and audio...")

	if pkg.IsFile(*videoFPath) == false {
		logger.Fatalln("video file is not exist")
	}

	// 然后调用 FFMPEG 进行音频的导出
	ffmpegInfo := npkg.ExportSubtitleAndAudioFile(".", *videoFPath)
	// 正常来说只会有一个音频，然后还是需要用户再传入 URL 的时候指定这个视频的语言，这里就不做判断了（因为不准）
	if len(ffmpegInfo.AudioInfoList) <= 0 {
		logger.Fatalln("ffmpegInfo.AudioInfoList <= 0")
	}
	//outAudioFPath := ffmpegInfo.AudioInfoList[0].FullPath

	logger.Infoln("Export Audio File Done ...")
	logger.Infoln("SubtitleInfoList Count:", len(ffmpegInfo.SubtitleInfoList))
	logger.Infoln("AudioInfoList Count:", len(ffmpegInfo.AudioInfoList))

	//// 获取这个 videoTitle 视频文件的文件名称，不包含后缀名
	//videoTitle := strings.TrimSuffix(filepath.Base(*videoFPath), filepath.Ext(*videoFPath))
	//// 获取这个 videoTitle 视频文件的文件夹目录
	//videoRootFolder := filepath.Dir(*videoFPath)
}

var AppVersion = "unknow"

const (
	SrtExt = ".srt"
)
