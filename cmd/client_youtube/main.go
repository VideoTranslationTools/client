package main

import (
	"context"
	"flag"
	"github.com/ChineseSubFinder/csf-supplier-base/pkg"
	"github.com/ChineseSubFinder/csf-supplier-base/pkg/ffmpeg_helper"
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

func exportAudioFile(c settings.Configs, youtubeVideoFPath string) *ffmpeg_helper.FFMPEGInfo {

	logger.Infoln("Export Audio File ...")
	ff := ffmpeg_helper.NewFFMPEGHelper(logger.GetLogger(), filepath.Join(c.CacheRootFolder, "ffmpeg_cache"))
	bok, ffmpegInfo, err := ff.ExportFFMPEGInfo(youtubeVideoFPath, ffmpeg_helper.Audio, ffmpeg_helper.MP3)
	if err != nil {
		logger.Fatalln("ff.ExportFFMPEGInfo", err)
	}
	if bok == false {
		logger.Fatalln("ff.ExportFFMPEGInfo", "bok== false")
	}

	logger.Infoln("Export Audio Done")

	// 导出了那些音频文件，列举出来
	for _, a := range ffmpegInfo.AudioInfoList {
		logger.Infof("Audio Index: %d, CodecType: %s, CodecName: %s, Duration: %f, GetOrgLanguage(): %s, GetName(): %s, GetLanguage(): %s\n",
			a.Index, a.CodecType, a.CodecName, a.Duration, a.GetOrgLanguage(), a.GetName(), a.GetLanguage().String())
	}

	return ffmpegInfo
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
	// 先下载 youtube 的视频
	youtubeVideoFPath := downloadYoutubeVideo(c, client, dlUrl)
	// 然后调用 FFMPEG 进行音频的导出
	ffmpegInfo := exportAudioFile(c, youtubeVideoFPath)

	println("ffmpegInfo.AudioInfoList[0].Index", ffmpegInfo.AudioInfoList[0].Index)

	logger.Infoln("Done")
}

const (
	videoDownloaded = "downloaded"
)
