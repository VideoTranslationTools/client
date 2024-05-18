package whisper_client

import (
	"fmt"
	"testing"
	"time"
)

func TestNewWhisperClient(t *testing.T) {

	serverURL := "http://127.0.0.1:5000"
	token := "1234567890"

	wc := NewWhisperClient(serverURL, token)

	// Example usage
	taskID := 123
	audioFPath := "D:\\tmp\\test_audio\\All Creatures Great and Small (2020) - S01E04 - A Tricki Case WEBDL-1080p\\英_1.mp3"
	language := "en"

	reply, err := wc.SendTask(taskID, audioFPath, language)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("SendTask Reply:", reply)

	for i := 0; i < 100; i++ {
		// Example of getting task status
		statusReply, err := wc.GetTaskStatus(taskID)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		fmt.Println("GetTaskStatus Reply:", statusReply)
		time.Sleep(1 * time.Second)
	}
}
