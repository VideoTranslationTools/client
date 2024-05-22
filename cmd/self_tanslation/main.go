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
	"strconv"
	"strings"
)

func init() {
	logger.Infoln("Version:", AppVersion)
}

// 准备工作
func prepare(inputOnlyDoFFMPEG bool) {

	if inputOnlyDoFFMPEG == true {
		logger.Infoln("Only Do FFMPEG, Not Do Recognize And Translate")
		return
	}
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

// 将一个视频文件处理得到翻译后的 srt 文件
func processVideo2TranslatedSrt(videoFPath, outPutDir string, onlyDoFFMPEG bool) {

	// 获取这个 videoTitle 视频文件的文件名称，不包含后缀名
	videoTitle := strings.TrimSuffix(filepath.Base(videoFPath), filepath.Ext(videoFPath))

	logger.Infoln("Video Title:", videoTitle)
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

	if onlyDoFFMPEG == true {
		logger.Infoln("Only Do FFMPEG, Not Do Recognize And Translate")
		return
	}

	needTranslateSRTFPath := getNeedTranslateSRTFPath(ffmpegInfo)
	translateProcessor(needTranslateSRTFPath, videoTitle, outPutDir)
}

// 获取需要翻译的 srt 文件路径
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
		// 需要进行语音转 SRT
		needTranslateSRTFPath = audio2SrtProcessor(whisperClient, audioFPath)
	}

	return needTranslateSRTFPath
}

// 翻译字幕处理器
func translateProcessor(needTranslateSRTFPath string, videoTitle string, outPutDir string) {

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
	videoTitle = base64.StdEncoding.EncodeToString([]byte(videoTitle))
	outPutDir = base64.StdEncoding.EncodeToString([]byte(outPutDir))
	// 准备进行翻译
	translator_llm.StartOllamaClient(needTranslateSRTFPath, videoTitle, outPutDir)
}

// 音频转字幕处理器
func audio2SrtProcessor(whisperClient *whisper_client.WhisperClient, audioFPath string) string {
	// 需要启动 WhisperX 的服务器
	whisper_client.StartWhisperServer(whisperClient)
	// 需要转换音频文件为字幕文件
	err := whisper_client.ProcessAudio2Srt(whisperClient, audioFPath, "")
	if err != nil {
		logger.Fatalln("whisper_client.ProcessAudio2Srt error:", err)
	}
	// 得到的字幕就是 audioFPath 替换后缀名为 srt 即可
	needTranslateSRTFPath := strings.ReplaceAll(audioFPath, path.Ext(audioFPath), SrtExt)
	if pkg.IsFile(needTranslateSRTFPath) == false {
		logger.Fatalln("needTranslateSRTFPath is not exist")
	}
	//
	whisper_client.StopWhisperServer()
	return needTranslateSRTFPath
}

// 获取索引对应的音频和 srt 文件路径
func getIndexAudioAndSrtFileFPath(mp3FPathList []string, index int) (bool, string, string) {

	for _, mp3 := range mp3FPathList {
		// 获取这个 mp3 文件的文件名称，不包含后缀名
		mp3Title := strings.TrimSuffix(filepath.Base(mp3), filepath.Ext(mp3))
		// 获取这个 mp3 文件的主要语言名称：英_1，中_1，那么这里就是英，中
		if strings.Contains(mp3Title, "_") == false {
			logger.Errorln("mp3Title not contains _,", mp3Title)
			continue
		}

		nowIndexStr := strings.Split(mp3Title, "_")[1]
		nowIndex, err := strconv.Atoi(nowIndexStr)
		if err != nil {
			logger.Errorln("strconv.Atoi", err)
			continue
		}
		if nowIndex != index {
			continue
		}
		// 查找可能存在的 srt 文件
		needTranslateSRTFPath := strings.ReplaceAll(mp3, filepath.Ext(mp3), SrtExt)
		if pkg.IsFile(needTranslateSRTFPath) == false {
			needTranslateSRTFPath = ""
		}
		return true, mp3, needTranslateSRTFPath
	}

	return false, "", ""
}

// 搜索指定的 FFMPEG 文件目录，从这里开始处理
func searchFFMPEGFilesDir(inputFFMPEGFilesDir string) []FFMPEGCacheInfo {

	// 遍历根目录下的文件夹名称
	logger.Infoln("Search FFMPEG Files Dir ...")
	var videoFolderFPath = make([]string, 0)
	var videoNames = make([]string, 0)
	err := filepath.Walk(inputFFMPEGFilesDir, func(fPath string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() == true && inputFFMPEGFilesDir != fPath {
			videoFolderFPath = append(videoFolderFPath, fPath)
			videoNames = append(videoNames, f.Name())
		}
		return nil
	})
	if err != nil {
		logger.Fatalln("filepath.Walk", err)
	}

	var ffmpegCacheInfoList = make([]FFMPEGCacheInfo, 0)

	// 遍历每一个文件夹下面的 mp3 和 srt 文件，这些文件可能有多个
	for index, videoFolder := range videoFolderFPath {

		logger.Infof("Search .mp3 and .srt Files Dir Index: %d / %d, Video Folder: %s\n", index+1, len(videoFolderFPath), videoFolder)

		// 遍历这个文件夹下的所有文件
		var mp3FPathList = make([]string, 0)
		var srtFPathList = make([]string, 0)
		// 是否有导出的标记
		var hasExported = false
		err := filepath.Walk(videoFolder, func(fPath string, f os.FileInfo, err error) error {
			if f == nil {
				return err
			}
			if f.IsDir() {
				return nil
			}
			if filepath.Ext(fPath) == ".mp3" {
				mp3FPathList = append(mp3FPathList, fPath)
			} else if filepath.Ext(fPath) == SrtExt {
				srtFPathList = append(srtFPathList, fPath)
			} else if f.Name() == "Exported" {
				hasExported = true
			}
			return nil
		})
		if err != nil {
			logger.Fatalln("filepath.Walk", err)
		}
		// 必须是导出成功的才继续处理，否则跳过
		if hasExported == false {
			logger.Infoln("Not Found Exported Folder")
			continue
		}
		/*
			获取到的 mp3 文件的命名规则举例：未知语言_0.mp3、英_1.mp3、中_2.mp3、日_3.mp3
			_0 代表是首选的音频语言，_1、_2、_3 代表是备选的音频语言
			对应的可能会出现的 srt 文件的命名规则举例：未知语言_0.srt、英_1.srt、中_2.srt、日_3.srt，这些文件未必会存在
			下面需要做的就是找到对应的 mp3 文件和 srt 文件，便于后续的查询和使用
		*/
		mainMP3FPath := ""
		mainSrtFPath := ""
		for index := 0; index < 5; index++ {
			var has bool
			has, mainMP3FPath, mainSrtFPath = getIndexAudioAndSrtFileFPath(mp3FPathList, index)
			if has == true {
				logger.Infof("Index: %d, MP3: %s, SRT: %s\n", index, mainMP3FPath, mainSrtFPath)
				break
			}
		}
		if mainMP3FPath == "" {
			logger.Errorln("mainMP3FPath is empty")
			continue
		}

		ffmpegCacheInfoList = append(ffmpegCacheInfoList, FFMPEGCacheInfo{
			VideoTitle: videoNames[index],
			AudioFPath: mainMP3FPath,
			SrtFPath:   mainSrtFPath,
		})
	}

	return ffmpegCacheInfoList
}

func main() {

	inputVideoFPath := flag.String("video", "", "需要制作机翻字幕的视频文件路径")
	inputVideosDir := flag.String("videos_dir", "", "需要制作机翻字幕的视频文件目录")
	inputOutPutDir := flag.String("out_dir", "", "翻译后字幕输出的根目录")
	inputOnlyDoFFMPEG := flag.Bool("only_do_ffmpeg", false, "只做音频、SRT导出，不做语音识别、翻译")
	inputFFMPEGFilesDir := flag.String("ffmpeg_files_dir", "", "已经使用 FFMPEG 提取好的音频和SRT文件目录")
	// ------------------------------------------------------------------------------
	logger.SetLoggerLevel(logrus.InfoLevel)
	flag.Parse()

	videoFPath := *inputVideoFPath
	videosDir := *inputVideosDir
	outPutDir := *inputOutPutDir
	onlyDoFFMPEG := *inputOnlyDoFFMPEG
	ffmpegFilesDir := *inputFFMPEGFilesDir
	// ------------------------------------------------------------------------------
	prepare(onlyDoFFMPEG)
	// ------------------------------------------------------------------------------
	if videoFPath != "" {
		logger.Infoln("--------------------------------------------------")
		logger.Infoln("Process Video File:", videoFPath)
		// 如果设置了视频文件路径，那么就直接处理这个视频文件
		processVideo2TranslatedSrt(videoFPath, outPutDir, onlyDoFFMPEG)

	} else if videosDir != "" {

		logger.Infoln("Process Videos Dir:", videosDir)
		var videosFPath = make([]string, 0)
		// 如果设置了视频文件目录，那么就遍历这个目录下的所有视频文件
		err := filepath.Walk(videosDir, func(fPath string, f os.FileInfo, err error) error {
			if f == nil {
				return err
			}
			if f.IsDir() {
				return nil
			}
			if npkg.IsExt(filepath.Ext(fPath)) == false {
				return nil
			}

			videosFPath = append(videosFPath, fPath)
			return nil
		})
		if err != nil {
			logger.Fatalln("filepath.Walk", err)
		}

		for index, videoFPath := range videosFPath {

			logger.Infoln("--------------------------------------------------")
			logger.Infof("Process Video Index: %d / %d, Video File Path: %s\n", index+1, len(videosFPath), videoFPath)

			processVideo2TranslatedSrt(videoFPath, outPutDir, onlyDoFFMPEG)
		}
	} else if ffmpegFilesDir != "" {
		FFMPEGCacheInfos := searchFFMPEGFilesDir(ffmpegFilesDir)
		for index, ffmpegCacheInfo := range FFMPEGCacheInfos {

			logger.Infoln("--------------------------------------------------")
			logger.Infof("Process FFMPEG Files Dir Index: %d / %d, Video Title: %s\n", index+1, len(FFMPEGCacheInfos), ffmpegCacheInfo.VideoTitle)

			// 如果有原始字幕，优先直接翻译字幕
			if ffmpegCacheInfo.SrtFPath != "" {
				// 如果有值，那么就是直接翻译字幕
				translateProcessor(ffmpegCacheInfo.SrtFPath, ffmpegCacheInfo.VideoTitle, outPutDir)
			} else {
				// 否则，转换音频为字幕
				var needTranslateSRTFPath = audio2SrtProcessor(whisperClient, ffmpegCacheInfo.AudioFPath)
				translateProcessor(needTranslateSRTFPath, ffmpegCacheInfo.VideoTitle, outPutDir)
			}
		}
	}

	logger.Infoln("All Done")
}

var AppVersion = "v0.1.1"
var whisperClient *whisper_client.WhisperClient

type FFMPEGCacheInfo struct {
	VideoTitle string
	AudioFPath string
	SrtFPath   string
}

const (
	SrtExt              = ".srt"
	AppCacheRootDirPath = "."
)
