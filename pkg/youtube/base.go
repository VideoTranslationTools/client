package youtube

import (
	"context"
	"github.com/ChineseSubFinder/csf-supplier-base/pkg"
	npkg "github.com/VideoTranslationTools/client/pkg"
	"github.com/VideoTranslationTools/client/pkg/settings"
	"github.com/WQGroup/logger"
	"github.com/go-resty/resty/v2"
	"github.com/schollz/progressbar/v3"
	"github.com/wader/goutubedl"
	"io"
	"os"
	"path/filepath"
)

const (
	VideoDownloaded = "downloaded"
)

// DownloadVideo 下载 Youtube 视频
func DownloadVideo(c settings.Configs, client *resty.Client, dlUrl string) string {

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

			if pkg.IsFile(filepath.Join(nowCacheRootFolder, VideoDownloaded)) == true {
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
	f, err = os.Create(filepath.Join(nowCacheRootFolder, VideoDownloaded))
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
