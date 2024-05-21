package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/ChineseSubFinder/csf-supplier-base/pkg"
	"github.com/ChineseSubFinder/csf-supplier-base/pkg/ffmpeg_helper"
	npkg "github.com/VideoTranslationTools/client/pkg"
	"github.com/VideoTranslationTools/client/pkg/translator_llm"
	"github.com/VideoTranslationTools/client/pkg/whisper_client"
	"github.com/WQGroup/logger"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func init() {

	logger.Infoln("Version:", AppVersion)

	logger.Infoln("Init Whisper Client ...")

	whisperConfig := whisper_client.ReadWhisperServerConfig()
	serverURL := fmt.Sprintf("http://127.0.0.1:%d", whisperConfig.Port)
	token := whisperConfig.Token

	whisperClient = whisper_client.NewWhisperClient(serverURL, token)

	logger.Infoln("Init Whisper Client Done")

	ollamaClientVersion := translator_llm.GetOllamaClientVersion()

	logger.Infoln("Ollama Client Version:", ollamaClientVersion)

	logger.Infoln("Read Ollama Config ...")

	translator_llm.ReadOllamaTranslatorConfig()

	logger.Infoln("Ollama config read Done")
}

func getNeedTranslateSRTFPath(ffmpegInfo *ffmpeg_helper.FFMPEGInfo) string {

	// 如果没有对应的内置字幕文件，那么就需要转换音频文件为字幕文件
	needConvertAudio2Srt := false
	if len(ffmpegInfo.SubtitleInfoList) <= 0 {
		needConvertAudio2Srt = true
	}
	if needConvertAudio2Srt == true {
		logger.Infoln("Need Convert Audio File To SRT File ...")
	} else {
		logger.Infoln("Use Exist Video Insider SRT File ...")
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

		// 需要启动 WhisperX 的服务器
		whisper_client.StartWhisperServer(whisperClient)
		// 需要转换音频文件为字幕文件
		err := whisper_client.ProcessAudio2Srt(whisperClient, audioFPath, "")
		if err != nil {
			logger.Fatalln("whisper_client.ProcessAudio2Srt error:", err)
			return ""
		}
		// 得到的字幕就是 audioFPath 替换后缀名为 srt 即可
		needTranslateSRTFPath = strings.ReplaceAll(audioFPath, path.Ext(audioFPath), SrtExt)
		if pkg.IsFile(needTranslateSRTFPath) == false {
			logger.Fatalln("needTranslateSRTFPath is not exist")
		}
		//
		whisper_client.StopWhisperServer()
	}

	return needTranslateSRTFPath
}

func main() {

	inputVideoFPath := flag.String("video", "", "the video file path")
	inputOutPutDir := flag.String("out_dir", "", "output dir path")
	// ------------------------------------------------------------------------------
	logger.SetLoggerLevel(logrus.InfoLevel)
	flag.Parse()
	// ------------------------------------------------------------------------------
	videoFPath := *inputVideoFPath
	outPutDir := *inputOutPutDir
	// 获取这个 videoTitle 视频文件的文件名称，不包含后缀名
	videoTitle := strings.TrimSuffix(filepath.Base(videoFPath), filepath.Ext(videoFPath))

	if outPutDir == "" {
		outPutDir = "."
		// 转换到绝对路径
		var err error
		outPutDir, err = filepath.Abs(outPutDir)
		if err != nil {
			logger.Fatalln("filepath.Abs", err)
		}
		logger.Infoln("Will output translated file into:", outPutDir)
	} else {
		// 如果这个目录不存在，则新建
		if pkg.IsDir(outPutDir) == false {
			err := os.MkdirAll(outPutDir, os.ModePerm)
			if err != nil {
				logger.Fatalln("MkdirAll", err)
			}
		}
		logger.Infoln("Will output translated file into:", outPutDir)
	}
	logger.Infoln("Video File Path:", videoFPath)
	if pkg.IsFile(videoFPath) == false {
		logger.Fatalln("video file is not exist")
	}

	logger.Infoln("Will export video subtitle and audio...")
	// 然后调用 FFMPEG 进行音频的导出
	ffmpegInfo := npkg.ExportSubtitleAndAudioFile(AppCacheRootDirPath, videoFPath)
	// 正常来说只会有一个音频，然后还是需要用户再传入 URL 的时候指定这个视频的语言，这里就不做判断了（因为不准）
	if len(ffmpegInfo.AudioInfoList) <= 0 {
		logger.Fatalln("ffmpegInfo.AudioInfoList <= 0")
	}
	logger.Infoln("Export Audio File Done ...")
	logger.Infoln("SubtitleInfoList Count:", len(ffmpegInfo.SubtitleInfoList))
	logger.Infoln("AudioInfoList Count:", len(ffmpegInfo.AudioInfoList))

	needTranslateSRTFPath := getNeedTranslateSRTFPath(ffmpegInfo)
	var err error
	if filepath.IsAbs(filepath.Join(AppCacheRootDirPath, needTranslateSRTFPath)) == false {
		needTranslateSRTFPath, err = filepath.Abs(filepath.Join(AppCacheRootDirPath, needTranslateSRTFPath))
		if err != nil {
			logger.Fatalln("filepath.Abs", err)
		}
	}

	logger.Infoln("Need Translate SRT File Path:", needTranslateSRTFPath)

	// 对于 needTranslateSRTFPath 进行 base64 加密
	needTranslateSRTFPath = base64.StdEncoding.EncodeToString([]byte(needTranslateSRTFPath))
	// 准备进行翻译
	translator_llm.StartOllamaClient(needTranslateSRTFPath, videoTitle, outPutDir)

	logger.Infoln("All Done")
}

var AppVersion = "unknow"
var whisperClient *whisper_client.WhisperClient

const (
	SrtExt              = ".srt"
	AppCacheRootDirPath = "."
)
