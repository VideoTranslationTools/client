package translator_llm

import (
	"fmt"
	"github.com/WQGroup/logger"
	"github.com/spf13/viper"
)

type OllamaTranslatorCommand struct {
	MaxTokenInput int     `mapstructure:"max_token_input"`
	NumPredict    int     `mapstructure:"num_predict"`
	Temperature   float32 `mapstructure:"temperature"`
	Seed          int     `mapstructure:"seed"`
	Model         string  `mapstructure:"model"`
	OllamaUrl     string  `mapstructure:"ollama_url"`
}

func NewOllamaTranslatorCommandArgs(cmd OllamaTranslatorCommand) []string {
	args := []string{
		fmt.Sprintf("--max_token_input=%d", cmd.MaxTokenInput),
		fmt.Sprintf("--num_predict=%d", cmd.NumPredict),
		fmt.Sprintf("--temperature=%f", cmd.Temperature),
		fmt.Sprintf("--seed=%d", cmd.Seed),
		fmt.Sprintf("--model=%s", cmd.Model),
		fmt.Sprintf("--ollama_url=%s", cmd.OllamaUrl),
	}

	return args
}

func ReadOllamaTranslatorConfig() OllamaTranslatorCommand {
	// 设置配置文件名和路径
	viper.SetConfigName("ollama_config") // 配置文件名为 config.yml
	viper.AddConfigPath(".")             // 在当前目录查找配置文件
	viper.SetConfigType("yaml")          // 指定配置文件类型为 YAML

	// 读取配置文件
	err := viper.ReadInConfig()
	if err != nil {
		logger.Fatalln("Error reading config file:", err)
	}

	// 将配置文件内容解析到 PythonCommand 结构体
	var cmd OllamaTranslatorCommand
	err = viper.Unmarshal(&cmd)
	if err != nil {
		logger.Fatalln("Error unmarshalling config:", err)
	}
	return cmd
}

func GetOllamaTranslatorCommandArgs(srtFilePath string, translatedTitle string, outPutDir string) []string {
	cmd := ReadOllamaTranslatorConfig()
	configArgsStrings := NewOllamaTranslatorCommandArgs(cmd)
	return append([]string{
		"--srt_file_path=\"" + srtFilePath + "\"",
		"--translated_title=\"" + translatedTitle + "\"",
		"--output_dir=\"" + outPutDir + "\"",
	}, configArgsStrings...)
}
