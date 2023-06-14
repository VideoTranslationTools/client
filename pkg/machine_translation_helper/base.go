package machine_translation_helper

import (
	"github.com/ChineseSubFinder/csf-supplier-base/pkg"
	t "github.com/VideoTranslationTools/base/db/task_system"
	"github.com/VideoTranslationTools/base/pkg/task_system"
	"github.com/WQGroup/logger"
	"path/filepath"
	"time"
)

type MachineTranslationHelper struct {
	client *task_system.TaskSystemClient
}

func NewMachineTranslationHelper(client *task_system.TaskSystemClient) *MachineTranslationHelper {
	return &MachineTranslationHelper{client: client}
}

func (m MachineTranslationHelper) Process(opt Opts) {

	if pkg.IsFile(opt.InputFPath) == false {
		logger.Fatal("InputFPath is not a file: %s", opt.InputFPath)
	}
	logger.Infoln("Try AddMachineTranslationTask...")
	// 申请添加任务
	taskPackageResp, err := m.client.AddMachineTranslationTask(
		"tt0000000", true, -1, -1,
		opt.IsAudioOrSRT, opt.InputFPath, opt.AudioLang, opt.TargetTranslationLang)
	if err != nil {
		logger.Fatal("AddMachineTranslationTask error: %v", err)
	}
	if taskPackageResp.Status != 1 {
		logger.Fatal("AddMachineTranslationTask Status != 1, error: %v", taskPackageResp.Message)
	}
	logger.Infoln("AddMachineTranslationTask:", taskPackageResp.Status, taskPackageResp.Message)

	logger.Infoln("Try Upload File...")
	// 上传文件
	err = pkg.UploadFile2R2(taskPackageResp.UploadURL, opt.InputFPath)
	if err != nil {
		logger.Fatal("UploadFile2R2 error: %v", err)
	}
	logger.Infoln("Upload File Done.")

	token := taskPackageResp.Token
	taskPackageID := taskPackageResp.TaskPackageId

	// 设置上传完毕
	firstPackageTaskDone, err := m.client.SetFirstPackageTaskDone(taskPackageID, token)
	if err != nil {
		logger.Fatal("SetFirstPackageTaskDone error: %v", err)
	}
	logger.Infoln("SetFirstPackageTaskDone:", firstPackageTaskDone.Status, firstPackageTaskDone.Message)
	logger.Infoln("--------------------------------------------")
	for true {

		time.Sleep(5 * time.Second)
		// 获取任务包状态
		taskPackageStatus, err := m.client.GetTaskPackageStatus(taskPackageID)
		if err != nil {
			logger.Errorln("GetTaskPackageStatus error: %v", err)
			continue
		}
		if taskPackageStatus.Status != 1 {
			logger.Errorln("GetTaskPackageStatus error: %v", taskPackageStatus.Message)
			continue
		}

		logger.Infoln("TaskPackageStatus:", taskPackageStatus.TaskPackageStatus.ToString())
		logger.Infoln("AudioToSubtitleStatus:", taskPackageStatus.AudioToSubtitleStatus.ToString())
		logger.Infoln("SplitTaskStatus:", taskPackageStatus.SplitTaskStatus.ToString())
		logger.Infoln("TranslationTask:", taskPackageStatus.TranslationTaskDoneCount, "/", taskPackageStatus.TranslationTaskCount)
		logger.Infoln("--------------------------------------------")

		if taskPackageStatus.TaskPackageStatus == t.Finished ||
			taskPackageStatus.TaskPackageStatus == t.Failed ||
			taskPackageStatus.TaskPackageStatus == t.Canceled {
			// 任务包结束
			break
		}
	}

	// 下载翻译结果
	translatedResult, err := m.client.GetTranslatedResult(taskPackageID)
	if err != nil {
		logger.Fatal("GetTranslatedResult error: %v", err)
	}
	if translatedResult.Status != 1 {
		logger.Fatal("GetTranslatedResult Status != 1, error: %v", translatedResult.Message)
	}
	logger.Infoln("GetTranslatedResult:", translatedResult.Status, translatedResult.Message)

	err = pkg.DownloadFile(opt.TranslatedSRTFileName, opt.DownloadedFileSaveFolderPath, translatedResult.ResultDownloadUrl)
	if err != nil {
		logger.Fatal("DownloadFile error: %v", err)
	}
	logger.Infoln("Translated SRT:", filepath.Join(opt.DownloadedFileSaveFolderPath, opt.TranslatedSRTFileName))
	logger.Infoln("DownloadFile Done.")
}

type Opts struct {
	InputFPath            string // 输入文件路径
	IsAudioOrSRT          bool   // 是否是音频或者字幕
	AudioLang             string // 音频语言
	TargetTranslationLang string // 目标翻译语言

	DownloadedFileSaveFolderPath string // 下载文件保存路径
	TranslatedSRTFileName        string // 翻译后的文件名
}
