package whisper_client

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
)

type WhisperClient struct {
	ServerURL string
	Token     string
	Client    *resty.Client
}

type Task struct {
	TaskID     int    `json:"task_id"`
	InputAudio string `json:"input_audio"`
	Language   string `json:"language"`
}

type SendTaskReply struct {
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
	Status int    `json:"status"`
}

func NewWhisperClient(serverURL, token string) *WhisperClient {
	client := resty.New()
	client.SetHeader("Authorization", "Bearer "+token)

	return &WhisperClient{
		ServerURL: serverURL,
		Token:     token,
		Client:    client,
	}
}

func (wc *WhisperClient) SendTask(taskID int, audioFPath, language string) (*SendTaskReply, error) {
	task := Task{
		TaskID:     taskID,
		InputAudio: audioFPath,
		Language:   language,
	}

	resp, err := wc.Client.R().
		SetBody(task).
		SetResult(&SendTaskReply{}).
		Post(wc.ServerURL + "/transcribe")

	if err != nil || resp.StatusCode() != 200 {
		return nil, errors.New(fmt.Sprintf("HTTP request failed with status %d", resp.StatusCode()))
	}

	rtaskResp := resp.Result().(*SendTaskReply)
	return rtaskResp, nil
}

func (wc *WhisperClient) GetTaskStatus(taskID int) (*SendTaskReply, error) {
	resp, err := wc.Client.R().
		SetResult(&SendTaskReply{}).
		Get(wc.ServerURL + fmt.Sprintf("/transcribe?task_id=%d", taskID))

	if err != nil || resp.StatusCode() != 200 {
		return nil, errors.New(fmt.Sprintf("HTTP request failed with status %d", resp.StatusCode()))
	}

	rtaskResp := resp.Result().(*SendTaskReply)
	return rtaskResp, nil
}
