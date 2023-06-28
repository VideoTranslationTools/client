package pkg

import (
	"github.com/ChineseSubFinder/csf-supplier-base/pkg/ffmpeg_helper"
	"github.com/WQGroup/logger"
	"github.com/schollz/progressbar/v3"
	"io"
	"path/filepath"
)

type ProgressWriter struct {
	writer     io.Writer
	bar        *progressbar.ProgressBar
	downloaded int64
	total      int64
}

func NewProgressWriter(writer io.Writer, bar *progressbar.ProgressBar, total int64) *ProgressWriter {
	return &ProgressWriter{writer: writer, bar: bar, downloaded: 0, total: total}
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.downloaded += int64(n)
	pw.bar.Set64(pw.downloaded)
	_, err := pw.writer.Write(p)
	return n, err
}

func ExportAudioFile(CacheRootFolder, videoFPath string) *ffmpeg_helper.FFMPEGInfo {

	logger.Infoln("Check ffmpeg ...")
	ff := ffmpeg_helper.NewFFMPEGHelper(logger.GetLogger(), filepath.Join(CacheRootFolder, "ffmpeg_cache"))
	_, err := ff.Version()
	if err != nil {
		logger.Fatalln("ff.Version", err)
	}
	logger.Infoln("Export Audio File ...")
	bok, ffmpegInfo, err := ff.ExportFFMPEGInfo(videoFPath, ffmpeg_helper.Audio, ffmpeg_helper.MP3)
	if err != nil {
		logger.Fatalln("ff.ExportFFMPEGInfo", err)
	}
	if bok == false {
		logger.Fatalln("ff.ExportFFMPEGInfo", "bok== false")
	}

	logger.Infoln("Export Audio Done")

	// 导出了那些音频文件，列举出来
	for _, a := range ffmpegInfo.AudioInfoList {
		logger.Infof("Audio Index: %d, CodecType: %s, CodecName: %s, Duration: %f, GetOrgLanguage(): %s",
			a.Index, a.CodecType, a.CodecName, a.Duration, a.GetOrgLanguage())
	}

	return ffmpegInfo
}
