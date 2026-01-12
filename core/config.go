package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"usb_tool/ui"
	"usb_tool/utils"
)

var (
	// SEARCH_DIR 默认为 /sdcard/ISO，这是 Android 通用路径
	SEARCH_DIR = "/sdcard/ISO"

	ConfigDir         string
	UdcName           string
	CurrentVID        = "0x0951"
	CurrentPID        = "0x1666"
	CurrentMan        = "Kingston"
	CurrentProd       = "DataTraveler"
	CurrentSer        string
	CurrentLoopDevice string
	SelinuxOriginal   string
	SpoofConfPath     string
)

func InitCore() error {
	// 1. 确定工作目录
	if envDir := os.Getenv("USB_TOOL_DIR"); envDir != "" {
		SEARCH_DIR = envDir
	}
	SpoofConfPath = filepath.Join(SEARCH_DIR, ".usb_spoof_config")

	// 2. 自动初始化目录结构
	firstRun := false
	if _, err := os.Stat(SEARCH_DIR); os.IsNotExist(err) {
		firstRun = true
		fmt.Printf("  %s正在初始化工作目录: %s...%s\n", ui.G_TEXT, SEARCH_DIR, ui.RESET)
		if err := os.MkdirAll(SEARCH_DIR, 0755); err != nil {
			return fmt.Errorf("无法创建工作目录: %v", err)
		}
	}
	
	// 初始化日志 (现在依赖 SEARCH_DIR 已存在)
	logDir := filepath.Join(SEARCH_DIR, "logs")
	if err := utils.InitLogging(logDir); err != nil {
		fmt.Printf("警告: 无法初始化日志: %v\n", err)
	}

	// 3. 检查 UDC
	udcDir := "/sys/class/udc"
	entries, err := os.ReadDir(udcDir)
	if err == nil && len(entries) > 0 {
		UdcName = entries[0].Name()
	} else {
		return fmt.Errorf("未检测到 UDC 控制器")
	}

	// 4. 检测 ConfigFS
	// 移除了 android0 支持，因为其配置方式不兼容 ConfigFS 逻辑
	paths := []string{
		"/config/usb_gadget/g1",
		"/sys/kernel/config/usb_gadget/g1",
	}
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			ConfigDir = path
			utils.Log("INFO", "检测到 ConfigFS 路径: %s", ConfigDir)
			break
		}
	}
	if ConfigDir == "" {
		return fmt.Errorf("无法检测到 ConfigFS 路径 (本工具仅支持 ConfigFS)")
	}

	LoadSpoofConfig()
	SaveSelinuxState()

	// 5. 如果是首次运行，显示引导信息
	if firstRun {
		ui.ClearScreen()
		fmt.Printf("\n%s  ========================================%s\n", ui.P_LIGHT, ui.RESET)
		fmt.Printf("  %s欢迎使用 USB Tool v9.0%s\n", ui.B_GREEN, ui.RESET)
		fmt.Printf("%s  ========================================%s\n\n", ui.P_LIGHT, ui.RESET)
		fmt.Printf("  检测到您是首次运行。\n\n")
		fmt.Printf("  已为您自动创建工作目录:\n")
		fmt.Printf("  %s%s%s\n\n", ui.B_CYAN, SEARCH_DIR, ui.RESET)
		fmt.Printf("  请将您的 %s.img%s 镜像文件放入此目录，\n", ui.B_CYAN, ui.RESET)
		fmt.Printf("  或者使用菜单中的 %s[2] 新建自定义镜像%s 功能。\n\n", ui.P_LIGHT, ui.RESET)
		ui.Pause()
	}

	return nil
}

func LoadSpoofConfig() {
	CurrentSer = fmt.Sprintf("MT%d", time.Now().Unix()) // Default
	if data, err := os.ReadFile(SpoofConfPath); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.Trim(strings.TrimSpace(parts[1]), "\"")
				switch key {
				case "CURRENT_VID":
					CurrentVID = val
				case "CURRENT_PID":
					CurrentPID = val
				case "CURRENT_MAN":
					CurrentMan = val
				case "CURRENT_PROD":
					CurrentProd = val
				case "CURRENT_SER":
					CurrentSer = val
				}
			}
		}
	} else {
		// 如果配置文件不存在，自动创建一个默认的
		SaveSpoofConfig()
	}
}

func SaveSpoofConfig() {
	content := fmt.Sprintf(`CURRENT_VID="%s"
CURRENT_PID="%s"
CURRENT_MAN="%s"
CURRENT_PROD="%s"
CURRENT_SER="%s"
`, CurrentVID, CurrentPID, CurrentMan, CurrentProd, CurrentSer)
	os.WriteFile(SpoofConfPath, []byte(content), 0644)
}

func SaveSelinuxState() {
	if out, err := exec.Command("getenforce").Output(); err == nil {
		SelinuxOriginal = strings.TrimSpace(string(out))
		utils.Log("INFO", "SELinux 原始状态: %s", SelinuxOriginal)
	}
}

func RestoreSelinuxState() {
	if SelinuxOriginal == "Enforcing" {
		exec.Command("setenforce", "1").Run()
		utils.Log("SUCCESS", "SELinux 恢复为 Enforcing 模式")
	}
}

func GetUSBSpeed() string {
	speedFile := fmt.Sprintf("/sys/class/udc/%s/current_speed", UdcName)
	// 优化: 减少轮询次数或间隔，因为主循环会频繁调用
	if data, err := os.ReadFile(speedFile); err == nil {
		speed := strings.TrimSpace(string(data))
		if speed != "" && speed != "unknown" {
			return speed
		}
	}
	return "UNKNOWN"
}
