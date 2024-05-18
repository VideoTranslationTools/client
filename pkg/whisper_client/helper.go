package whisper_client

import "math/rand"

func ProcessAudio2Srt(wclient *WhisperClient, audioFPath string, language string) error {

	nowTaskID := rand.Int()
	reply, err := wclient.SendTask(nowTaskID, audioFPath, language)
	if err != nil {
		return err
	}
}
