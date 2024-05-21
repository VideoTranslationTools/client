package whisper_client

import (
	"fmt"
	"github.com/WQGroup/logger"
	"github.com/spf13/viper"
)

type WhisperServerCommand struct {
	Device       string `mapstructure:"device"`
	DeviceIndex  string `mapstructure:"device_index"`
	ComputeType  string `mapstructure:"compute_type"`
	GPUID        int    `mapstructure:"gpu_id"`
	ModelSize    string `mapstructure:"model_size"`
	DownloadRoot string `mapstructure:"download_root"`
	Port         int    `mapstructure:"port"`
	Token        string `mapstructure:"token"`
	VAD          bool   `mapstructure:"vad"`
}

// NewWhisperServerCommandArgs 函数接收 WhisperServerCommand 结构体作为参数
func NewWhisperServerCommandArgs(cmd WhisperServerCommand) []string {
	// 构建命令行参数字符串
	args := []string{
		fmt.Sprintf("--device=%s", cmd.Device),
		fmt.Sprintf("--device_index=%s", cmd.DeviceIndex),
		fmt.Sprintf("--compute_type=%s", cmd.ComputeType),
		fmt.Sprintf("--gpu_id=%d", cmd.GPUID),
		fmt.Sprintf("--model_size=%s", cmd.ModelSize),
		fmt.Sprintf("--download_root=%s", cmd.DownloadRoot),
		fmt.Sprintf("--port=%d", cmd.Port),
		fmt.Sprintf("--token=%s", cmd.Token),
		fmt.Sprintf("--vad=%t", cmd.VAD),
	}

	return args
}

func ReadWhisperServerConfig() WhisperServerCommand {
	// 设置配置文件名和路径
	viper.SetConfigName("whisper_server_config") // 配置文件名为 config.yml
	viper.AddConfigPath(".")                     // 在当前目录查找配置文件
	viper.SetConfigType("yaml")                  // 指定配置文件类型为 YAML

	// 读取配置文件
	err := viper.ReadInConfig()
	if err != nil {
		logger.Fatalln("Error reading whisper config file:", err)
	}

	// 将配置文件内容解析到 PythonCommand 结构体
	var cmd WhisperServerCommand
	err = viper.Unmarshal(&cmd)
	if err != nil {
		logger.Fatalln("Error unmarshalling whisper config:", err)
	}
	return cmd
}

func GetWhisperServerCommandArgs() []string {
	args := NewWhisperServerCommandArgs(ReadWhisperServerConfig())
	return args
}
