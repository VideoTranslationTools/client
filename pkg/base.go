package pkg

import (
	"github.com/ChineseSubFinder/csf-supplier-base/pkg/ffmpeg_helper"
	"github.com/WQGroup/logger"
	"github.com/schollz/progressbar/v3"
	"io"
	"path/filepath"
	"strings"
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

func ExportSubtitleAndAudioFile(CacheRootFolder, videoFPath string) *ffmpeg_helper.FFMPEGInfo {

	logger.Infoln("Check ffmpeg ...")
	ff := ffmpeg_helper.NewFFMPEGHelper(logger.GetLogger(), filepath.Join(CacheRootFolder, "ffmpeg_cache"))
	_, err := ff.Version()
	if err != nil {
		logger.Fatalln("ff.Version", err)
	}
	logger.Infoln("Export Subtitle And Audio File ...")
	bok, ffmpegInfo, err := ff.ExportFFMPEGInfo(videoFPath, ffmpeg_helper.SubtitleAndAudio, ffmpeg_helper.MP3)
	if err != nil {
		logger.Fatalln("ff.ExportFFMPEGInfo", err)
	}
	if bok == false {
		logger.Fatalln("ff.ExportFFMPEGInfo", "bok== false")
	}

	logger.Infoln("Export Subtitle And Audio Done")

	// 导出了那些音频文件，列举出来
	for _, a := range ffmpegInfo.AudioInfoList {
		logger.Infof("Audio Index: %d, CodecType: %s, CodecName: %s, Duration: %f, GetOrgLanguage(): %s",
			a.Index, a.CodecType, a.CodecName, a.Duration, a.GetOrgLanguage())
	}
	// 导出了哪些字幕文件，列举出来
	for _, s := range ffmpegInfo.SubtitleInfoList {
		logger.Infof("Subtitle Index: %d, CodecType: %s, CodecName: %s, Name: %s",
			s.Index, s.CodecType, s.CodecName, s.GetName())
	}

	return ffmpegInfo
}

func IsExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".mp4", ".mkv", ".avi", ".flv", ".mov", ".wmv", ".webm", ".rm", ".asf", ".mpg", ".mpeg", ".m4v", ".3gp", ".3g2", ".ts", ".mts", ".m2ts", ".vob", ".f4v", ".m2v", ".dat", ".amv", ".divx", ".mpv", ".ogv", ".qt", ".f4p", ".f4a", ".f4b", ".m2p", ".m4b", ".m4p", ".m4r", ".nsv", ".ogm", ".ogx", ".rec", ".rmvb", ".tod", ".tts", ".vro", ".wtv", ".xesc":
		return true
	}
	return false
}
