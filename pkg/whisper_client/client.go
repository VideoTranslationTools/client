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

/*
Status:
   pending = 1
   running = 2
   finished = 3
   error = 4
*/
type SendTaskReply struct {
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
	Status int    `json:"status"`
}

func (s SendTaskReply) GetStatus() TaskStatus {
	return TaskStatus(s.Status)
}

type TaskStatus int

const (
	Pending TaskStatus = iota + 1
	Running
	Finished
	Error
)

func (t TaskStatus) String() string {
	return [...]string{"Pending", "Running", "Finished", "Error"}[(int)(t)-1]
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

func (wc *WhisperClient) IsAlive() (bool, string) {

	resp, err := wc.Client.R().
		Get(wc.ServerURL + "/")

	if err != nil || resp.StatusCode() != 200 {
		return false, ""
	}

	return true, resp.String()
}
