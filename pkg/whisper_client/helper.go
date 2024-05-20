package whisper_client

import (
	"github.com/VideoTranslationTools/client/pkg"
	"github.com/WQGroup/logger"
	"github.com/pkg/errors"
	"math/rand"
	"time"
)

var execWrapper *pkg.ExecWrapper

func StartWhisperServer() {

	if execWrapper == nil {
		execWrapper = pkg.NewExecWrapper(false)
	}

	logger.Infoln("Start Whisper Server ...")

	err := execWrapper.Start(".\\whisper-server.exe", GetWhisperServerCommandArgs()...)
	if err != nil {
		logger.Fatalln("Start Whisper Server failed", err)
		return
	}
}

func StopWhisperServer() {
	logger.Infoln("Stop Whisper Server ...")
	if execWrapper == nil {
		execWrapper = pkg.NewExecWrapper(false)
	}
	err := execWrapper.Stop()
	if err != nil {
		logger.Errorln("Stop Whisper Server failed", err)
		return
	}
	logger.Infoln("Stop Whisper Server Successfully")
}

func ProcessAudio2Srt(wclient *WhisperClient, audioFPath string, language string) error {

	// 等待 Whisper Server 启动
	isUp, whisperServerVersion := wclient.IsAlive()
	if isUp == false {
		for {
			isUp, whisperServerVersion = wclient.IsAlive()
			if isUp == true {
				break
			}
			logger.Infoln("Whisper Server is not up, wait 5 seconds to check again")
			time.Sleep(5 * time.Second)
		}
	}
	logger.Infoln("Whisper Server Version:", whisperServerVersion)

	logger.Infoln("Start Whisper Server Successfully")

	rand.Seed(time.Now().UnixNano())
	// 生成 8 位长度的随机数
	nowTaskID := rand.Intn(90000000) + 10000000
	reply, err := wclient.SendTask(nowTaskID, audioFPath, language)
	if err != nil {
		logger.Errorln("Code:", reply.Code, "Msg:", reply.Msg,
			"Status:", TaskStatus(reply.Status).String())
		return err
	}

	for {
		statusReply, err := wclient.GetTaskStatus(nowTaskID)
		if err != nil {
			logger.Errorln("Code:", statusReply.Code, "Msg:", statusReply.Msg,
				"Status:", TaskStatus(statusReply.Status).String())
			return err
		}

		if statusReply.Status == int(Finished) {
			logger.Infoln("Code:", statusReply.Code, "Msg:", statusReply.Msg,
				"Status:", TaskStatus(statusReply.Status).String())
			break
		} else if statusReply.Status == int(Error) {
			logger.Errorln("Code:", statusReply.Code, "Msg:", statusReply.Msg,
				"Status:", TaskStatus(statusReply.Status).String())
			return errors.New("Task failed")
		} else if statusReply.Status == int(Pending) {
			logger.Infoln("Code:", statusReply.Code, "Msg:", statusReply.Msg,
				"Status:", TaskStatus(statusReply.Status).String())
		} else if statusReply.Status == int(Running) {
			logger.Infoln("Code:", statusReply.Code, "Msg:", statusReply.Msg,
				"Status:", TaskStatus(statusReply.Status).String())
		}

		time.Sleep(5 * time.Second)
	}

	return nil
}
