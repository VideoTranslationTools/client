package pkg

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
)

// ExecWrapper 结构体用于执行外部程序
type ExecWrapper struct {
	cmd         *exec.Cmd
	hideWindows bool
	StdoutBuf   *bytes.Buffer // 用于保存标准输出信息
	StderrBuf   *bytes.Buffer // 用于保存标准错误输出信息
}

func NewExecWrapper(hideWindows bool) *ExecWrapper {
	return &ExecWrapper{
		hideWindows: hideWindows,
		StdoutBuf:   new(bytes.Buffer),
		StderrBuf:   new(bytes.Buffer),
	}
}

// Start 方法用于启动外部程序
func (e *ExecWrapper) Start(command string, args ...string) error {
	e.cmd = exec.CommandContext(context.Background(), command, args...)
	e.cmd.Stdout = io.MultiWriter(os.Stdout, e.StdoutBuf)
	e.cmd.Stderr = io.MultiWriter(os.Stderr, e.StderrBuf)
	// 隐藏执行窗口（仅适用于Windows平台）
	e.cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: e.hideWindows}

	err := e.cmd.Start()
	if err != nil {
		return err
	}

	return nil
}

// Wait 方法用于等待外部程序执行完毕
func (e *ExecWrapper) Wait() error {
	if e.cmd == nil {
		return fmt.Errorf("程序未启动")
	}

	err := e.cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}

// Stop 方法用于停止外部程序
func (e *ExecWrapper) Stop() error {
	if e.cmd == nil || e.cmd.Process == nil {
		return fmt.Errorf("程序未启动")
	}

	cmd := exec.CommandContext(context.Background(), "taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", e.cmd.Process.Pid))
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
