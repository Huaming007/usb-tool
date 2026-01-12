package ui

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"usb_tool/utils"
)

const (
	P_DARK  = "\033[38;5;93m"
	P_LIGHT = "\033[38;5;129m"
	G_TEXT  = "\033[38;5;242m"
	B_CYAN  = "\033[1;36m"
	B_GREEN = "\033[1;32m"
	WARN    = "\033[1;31m"
	RESET   = "\033[0m"
)

var (
	ScreenWidth  = 80
	ScreenHeight = 24
	CompactMode  = false
)

func DetectScreenSize() {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	if output, err := cmd.Output(); err == nil {
		parts := strings.Fields(string(output))
		if len(parts) == 2 {
			h, _ := strconv.Atoi(parts[0])
			w, _ := strconv.Atoi(parts[1])
			if w > 0 && h > 0 {
				ScreenWidth = w
				ScreenHeight = h
			}
		}
	} else {
		if cols := os.Getenv("COLUMNS"); cols != "" {
			if w, err := strconv.Atoi(cols); err == nil && w > 0 {
				ScreenWidth = w
			}
		}
	}

	if ScreenWidth < 60 {
		CompactMode = true
	}
	utils.Log("INFO", "屏幕分辨率检测: %dx%d, 紧凑模式: %v", ScreenWidth, ScreenHeight, CompactMode)
}

func ClearScreen() {
	fmt.Print("\033[H\033[2J")
}

// padString 用于处理含中文的字符串对齐
func padString(s string, width int) string {
	// 简单估算：假设每个 rune 都是 1 宽，实际上中文是 2。
	// 更精确的做法需要 unicode/eastasianwidth，但这里我们用简单补偿
	// 计算实际显示宽度 (大致)
	displayWidth := 0
	for _, r := range s {
		if r > 127 {
			displayWidth += 2
		} else {
			displayWidth++
		}
	}
	
	padding := width - displayWidth
	if padding < 0 {
		padding = 0
	}
	return s + strings.Repeat(" ", padding)
}

func PrintLogo(udcName, speed, currentMan, currentVid, loopStatus string, isLoopActive bool) {
	ClearScreen()

	separator := strings.Repeat("─", ScreenWidth-2)
	fmt.Printf("%s %s%s\n", P_DARK, separator, RESET)
	fmt.Printf("%s    USB MASS STORAGE ENFORCER v9.0 - Go Edition%s\n", P_LIGHT, RESET)
	fmt.Printf("%s %s%s\n", P_DARK, separator, RESET)

	loopColor := G_TEXT
	if isLoopActive {
		loopColor = B_GREEN
	}

	// 使用 padString 替代 printf 的 %-Ns，手动控制对齐
	if CompactMode {
		fmt.Printf("  %sUDC:%s%s%s SPD:%s%s%s\n", G_TEXT, B_CYAN, udcName, G_TEXT, B_CYAN, speed, RESET)
		fmt.Printf("  %sMAN:%s%s%s SPOOF:%s%s%s\n", G_TEXT, B_CYAN, currentMan, G_TEXT, B_CYAN, currentVid, RESET)
		fmt.Printf("  %sLOOP:%s%s%s\n", G_TEXT, loopColor, loopStatus, RESET)
	} else {
		// 手动对齐布局
		// Label (4) + Value (18) + Label (6) + Value (Rest)
		
		l1 := fmt.Sprintf("%sUDC: %s", G_TEXT, B_CYAN)
		v1 := padString(udcName, 15)
		l2 := fmt.Sprintf("%s%sSPEED: %s", RESET, G_TEXT, B_CYAN)
		v2 := speed

		fmt.Printf("  %s%s%s%s%s\n", l1, v1, l2, v2, RESET)

		l3 := fmt.Sprintf("%sMAN: %s", G_TEXT, B_CYAN)
		v3 := padString(currentMan, 15)
		l4 := fmt.Sprintf("%s%sSPOOF: %s", RESET, G_TEXT, B_CYAN)
		v4 := currentVid

		fmt.Printf("  %s%s%s%s%s\n", l3, v3, l4, v4, RESET)

		l5 := fmt.Sprintf("%sLOOP: %s", G_TEXT, loopColor)
		v5 := padString(loopStatus, 15)
		l6 := fmt.Sprintf("%s%sUSB:   %s", RESET, G_TEXT, B_GREEN) // 稍微调整 label 让对齐好看点
		v6 := "USB 3.1" // 简化

		fmt.Printf("  %s%s%s%s%s\n", l5, v5, l6, v6, RESET)
	}
	fmt.Printf("%s %s%s\n", P_DARK, separator, RESET)
}

func ShowMainMenu(currentMan string) {
	fmt.Printf("  %s[1] 挂载镜像文件%s\n", P_LIGHT, RESET)
	fmt.Printf("  %s[2] 新建自定义镜像%s\n", P_LIGHT, RESET)
	fmt.Printf("  %s[3] 切换伪装方案: %s%s%s\n", P_LIGHT, B_CYAN, currentMan, RESET)
	fmt.Printf("  %s[4] 镜像文件管理%s\n", P_LIGHT, RESET)
	fmt.Printf("  %s[5] 设备状态监控%s\n", P_LIGHT, RESET)
	fmt.Printf("  %s[6] 显示使用帮助%s\n", P_LIGHT, RESET)
	fmt.Printf("  %s[7] 重新枚举设备%s\n", P_LIGHT, RESET)
	fmt.Printf("  %s[8] 安全卸载恢复%s\n", P_LIGHT, RESET)
	fmt.Printf("  %s[0] 退出程序%s\n", P_LIGHT, RESET)

	separator := strings.Repeat("─", ScreenWidth-2)
	fmt.Printf("%s %s%s\n", P_DARK, separator, RESET)
}

func ReadLine(prompt string) string {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text())
	}
	return ""
}

func Pause() {
	fmt.Printf("\n  %s按回车继续...%s", G_TEXT, RESET)
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}