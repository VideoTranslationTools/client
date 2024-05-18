package main

import (
	"flag"
	"github.com/ChineseSubFinder/csf-supplier-base/pkg"
	"github.com/ChineseSubFinder/csf-supplier-base/pkg/ffmpeg_helper"
	npkg "github.com/VideoTranslationTools/client/pkg"
	"github.com/VideoTranslationTools/client/pkg/whisper_client"
	"github.com/WQGroup/logger"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
)

func init() {
	logger.Infoln("Version:", AppVersion)

	serverURL := "http://127.0.0.1:5000"
	token := "1234567890"

	whisperClient = whisper_client.NewWhisperClient(serverURL, token)
}

func getNeedTranslateSRTFPath(ffmpegInfo *ffmpeg_helper.FFMPEGInfo) string {

	// 如果没有对应的内置字幕文件，那么就需要转换音频文件为字幕文件
	needConvertAudio2Srt := false
	if len(ffmpegInfo.SubtitleInfoList) <= 0 {
		needConvertAudio2Srt = true
	}
	// 获取这个 audioTitle 音频文件的文件名称，不包含后缀名
	audioFPath := ffmpegInfo.AudioInfoList[0].FullPath
	audioTitle := strings.TrimSuffix(filepath.Base(audioFPath), filepath.Ext(audioFPath))
	// 得到这个音频文件的主要语言名称：英_1，中_1，那么这里就是英，中
	if strings.Contains(audioTitle, "_") == false {
		logger.Fatalln("audioTitle not contains _")
	}
	audioLang := strings.Split(audioTitle, "_")[0]
	// 需要进行翻译的 srt 文件
	needTranslateSRTFPath := ""
	if needConvertAudio2Srt == false {
		// 说明有导出内置字幕，那么大概率内置字幕是有对应的音频文件的
		// 首先需要做到主要的音频对应的内置字幕文件的转换
		// 这里有一个特点，导出的字幕，除了 ass 一定还伴随着响应的 srt 文件
		// 那么就从 SubtitleInfoList 中找到这个 ass 文件出来
		for _, subtitleInfo := range ffmpegInfo.SubtitleInfoList {
			tmpTitle := strings.TrimSuffix(filepath.Base(subtitleInfo.FullPath), filepath.Ext(subtitleInfo.FullPath))
			if strings.Contains(tmpTitle, audioLang) == true {
				needTranslateSRTFPath = subtitleInfo.FullPath
				break
			}
		}
		// 如果遍历一圈下来都没得，那么就使用第一个字幕文件
		if needTranslateSRTFPath == "" {
			needTranslateSRTFPath = ffmpegInfo.SubtitleInfoList[0].FullPath
		}
	} else {
		// 需要转换音频文件为字幕文件

	}

	return needTranslateSRTFPath
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

	needTranslateSRTFPath := getNeedTranslateSRTFPath(ffmpegInfo)
	logger.Infoln("Need Translate SRT File Path:", needTranslateSRTFPath)

	//// 获取这个 videoTitle 视频文件的文件名称，不包含后缀名
	//videoTitle := strings.TrimSuffix(filepath.Base(*videoFPath), filepath.Ext(*videoFPath))
	//// 获取这个 videoTitle 视频文件的文件夹目录
	//videoRootFolder := filepath.Dir(*videoFPath)
}

var AppVersion = "unknow"
var whisperClient *whisper_client.WhisperClient

const (
	SrtExt = ".srt"
)
