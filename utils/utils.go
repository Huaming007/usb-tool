package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	LOCK_FILE = "/tmp/usb_tool.lock"
)
var LogFile *os.File

func CheckRoot() bool {
	return os.Geteuid() == 0
}

func InitLogging(logDir string) error {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}
	date := time.Now().Format("20060102")
	logPath := filepath.Join(logDir, "usb_tool_"+date+".log")
	var err error
	LogFile, err = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	Log("INFO", "=== USB Tool Session Started ===")
	return nil
}
func Log(level, format string, args ...interface{}) {
	if LogFile != nil {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		message := fmt.Sprintf(format, args...)
		LogFile.WriteString(fmt.Sprintf("[%s] [%s] %s\n", timestamp, level, message))
	}
}

func AcquireLock() error {
	if _, err := os.Stat(LOCK_FILE); err == nil {
		data, _ := os.ReadFile(LOCK_FILE)
		pidStr := strings.TrimSpace(string(data))
		if pid, err := strconv.Atoi(pidStr); err == nil {
			if process, err := os.FindProcess(pid); err == nil {
				// 发送信号 0 检查进程是否存在
				if err := process.Signal(syscall.Signal(0)); err == nil {
					return fmt.Errorf("脚本已经在运行中 (PID: %d)", pid)
				}
			}
		}
		// 进程不存在，删除旧锁
		os.Remove(LOCK_FILE)
	}
	pid := os.Getpid()
	return os.WriteFile(LOCK_FILE, []byte(strconv.Itoa(pid)), 0644)
}

func ReleaseLock() {
	os.Remove(LOCK_FILE)
}

func CloseLog() {
	if LogFile != nil {
		LogFile.Close()
	}
}
