package machine_translation_helper

import (
	"github.com/ChineseSubFinder/csf-supplier-base/pkg"
	"github.com/ChineseSubFinder/csf-supplier-base/pkg/struct_json"
	t "github.com/VideoTranslationTools/base/db/task_system"
	"github.com/VideoTranslationTools/base/pkg/task_system"
	"github.com/WQGroup/logger"
	"path/filepath"
	"time"
)

type MachineTranslationHelper struct {
	client *task_system.TaskSystemClient
	cache  *NowTaskPackageCacheInfo
}

func NewMachineTranslationHelper(client *task_system.TaskSystemClient) *MachineTranslationHelper {
	return &MachineTranslationHelper{client: client, cache: NewNowTaskPackageCacheInfo()}
}

func (m *MachineTranslationHelper) Process(opt Opts) {

	if pkg.IsFile(opt.InputFPath) == false {
		logger.Fatal("InputFPath is not a file: %s", opt.InputFPath)
	}
	logger.Infoln("Try AddMachineTranslationTask...")
	// 申请添加任务
	taskPackageResp, err := m.client.AddMachineTranslationTask(
		"tt0000000", true, -1, -1,
		opt.IsAudioOrSRT, opt.InputFPath, opt.AudioLang, opt.TargetTranslationLang)
	if err != nil {
		logger.Fatal("AddMachineTranslationTask error:", err)
	}
	//cancelTaskPackage, err := m.client.CancelTaskPackage(taskPackageResp.TaskPackageId)
	//if err != nil {
	//	logger.Fatal("CancelTaskPackage error:", err)
	//}
	//if cancelTaskPackage.Status != 1 {
	//	logger.Fatal("CancelTaskPackage error:", cancelTaskPackage.Message)
	//}
	// 查询本地的任务缓存信息
	err = m.cache.Load()
	if err != nil {
		logger.Fatal("Load NowTaskPackageCacheInfo error:", err)
	}
	if m.cache.GetTaskPackageID() != taskPackageResp.TaskPackageId && taskPackageResp.TaskPackageId != "" {
		// 任务包信息不一致，需要更新
		err = m.cache.Update(taskPackageResp.TaskPackageId, "")
		if err != nil {
			logger.Fatal("Save NowTaskPackageCacheInfo TaskPackageId error:", err)
		}
	}
	if m.cache.GetToken() != taskPackageResp.Token && taskPackageResp.Token != "" {
		// 任务包信息不一致，需要更新
		err = m.cache.Update("", taskPackageResp.Token)
		if err != nil {
			logger.Fatal("Save NowTaskPackageCacheInfo Token error:", err)
		}
	}
	if taskPackageResp.Status != 1 {
		logger.Warningln("AddMachineTranslationTask Status != 1, Msg:", taskPackageResp.Message)
	}
	logger.Infoln("AddMachineTranslationTask:", taskPackageResp.Status, taskPackageResp.Message)

	logger.Infoln("Try Upload File...")
	// 上传文件
	err = pkg.UploadFile2R2(taskPackageResp.UploadURL, opt.InputFPath)
	if err != nil {
		logger.Fatal("UploadFile2R2 error:", err)
	}
	logger.Infoln("Upload File Done.")

	// 设置上传完毕
	firstPackageTaskDone, err := m.client.SetFirstPackageTaskDone(m.cache.GetTaskPackageID(), m.cache.GetToken())
	if err != nil {
		logger.Fatal("SetFirstPackageTaskDone error:", err)
	}
	logger.Infoln("SetFirstPackageTaskDone:", firstPackageTaskDone.Status, firstPackageTaskDone.Message)
	logger.Infoln("--------------------------------------------")
	for true {

		time.Sleep(5 * time.Second)
		// 获取任务包状态
		taskPackageStatus, err := m.client.GetTaskPackageStatus(m.cache.GetTaskPackageID())
		if err != nil {
			logger.Errorln("GetTaskPackageStatus error:", err)
			continue
		}
		if taskPackageStatus.Status != 1 {
			logger.Errorln("GetTaskPackageStatus error:", taskPackageStatus.Message)
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
	translatedResult, err := m.client.GetTranslatedResult(m.cache.GetTaskPackageID())
	if err != nil {
		logger.Fatal("GetTranslatedResult error:", err)
	}
	if translatedResult.Status != 1 {
		logger.Fatal("GetTranslatedResult Status != 1, error:", translatedResult.Message)
	}
	logger.Infoln("GetTranslatedResult:", translatedResult.Status, translatedResult.Message)

	err = pkg.DownloadFile(opt.TranslatedSRTFileName, opt.DownloadedFileSaveFolderPath, translatedResult.ResultDownloadUrl)
	if err != nil {
		logger.Fatal("DownloadFile error:", err)
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

type NowTaskPackageCacheInfo struct {
	taskPackageID string // 任务包 ID
	token         string // 任务包 Token
}

func NewNowTaskPackageCacheInfo() *NowTaskPackageCacheInfo {
	return &NowTaskPackageCacheInfo{}
}

func (n *NowTaskPackageCacheInfo) GetTaskPackageID() string {
	return n.taskPackageID
}

func (n *NowTaskPackageCacheInfo) GetToken() string {
	return n.token
}

func (n *NowTaskPackageCacheInfo) Update(taskPackageID, token string) error {
	if taskPackageID != "" {
		n.taskPackageID = taskPackageID
	}
	if token != "" {
		n.token = token
	}
	return struct_json.ToFile(saveJsonCacheFileName, n)
}

func (n *NowTaskPackageCacheInfo) Load() error {

	if pkg.IsFile(saveJsonCacheFileName) == false {
		return nil
	}
	// 从硬盘加载
	var nn NowTaskPackageCacheInfo
	err := struct_json.ToStruct(saveJsonCacheFileName, &nn)
	if err != nil {
		return err
	}
	n.taskPackageID = nn.taskPackageID
	n.token = nn.token
	return nil
}

const (
	saveJsonCacheFileName = "now_task_package_cache.json"
)
