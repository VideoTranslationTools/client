package translator_llm

import (
	"github.com/VideoTranslationTools/client/pkg"
	"github.com/WQGroup/logger"
)

var execWrapper *pkg.ExecWrapper

const (
	ollamaExePath = ".\\translation_by_ollama.exe"
)

func GetOllamaClientVersion() string {
	if execWrapper == nil {
		execWrapper = pkg.NewExecWrapper(false)
	}
	err := execWrapper.Start(ollamaExePath, []string{"--get_version=1"}...)
	if err != nil {
		logger.Fatalln("Start Ollama Client failed", err)
	}

	err = execWrapper.Wait()
	if err != nil {
		logger.Fatalln("Wait Ollama Client failed", err)
	}

	return execWrapper.StdoutBuf.String()
}

func StartOllamaClient(needTranslateSrtFilePath, translatedTitle, outPutDir string) {

	if execWrapper == nil {
		execWrapper = pkg.NewExecWrapper(false)
	}

	logger.Infoln("Start Ollama Client ...")

	err := execWrapper.Start(ollamaExePath,
		GetOllamaTranslatorCommandArgs(needTranslateSrtFilePath, translatedTitle, outPutDir)...)
	if err != nil {
		logger.Fatalln("Start Ollama Client failed", err)
		return
	}
	logger.Infoln("Wait Ollama Client Process ...")
	err = execWrapper.Wait()
	if err != nil {
		logger.Fatalln("Wait Ollama Client failed", err)
		return
	}
	logger.Infoln("Wait Ollama Client Successfully")
}

func StopOllamaClient() {
	logger.Infoln("Stop Ollama Client ...")
	if execWrapper == nil {
		execWrapper = pkg.NewExecWrapper(false)
	}
	err := execWrapper.Stop()
	if err != nil {
		logger.Errorln("Stop Ollama Client failed", err)
		return
	}
	logger.Infoln("Stop Ollama Client Successfully")
}
