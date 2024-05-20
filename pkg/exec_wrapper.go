package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// ExecWrapper 结构体用于执行外部程序
type ExecWrapper struct {
	cmd         *exec.Cmd
	hideWindows bool
}

func NewExecWrapper(hideWindows bool) *ExecWrapper {
	return &ExecWrapper{
		hideWindows: hideWindows,
	}
}

// Start 方法用于启动外部程序
func (e *ExecWrapper) Start(command string, args ...string) error {
	e.cmd = exec.Command(command, args...)
	e.cmd.Stdout = os.Stdout
	e.cmd.Stderr = os.Stderr
	// 隐藏执行窗口（仅适用于Windows平台）
	e.cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: e.hideWindows}

	err := e.cmd.Start()
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

	err := e.cmd.Process.Kill()
	if err != nil {
		return err
	}

	return nil
}
