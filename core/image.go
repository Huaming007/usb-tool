package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"usb_tool/ui"
	"usb_tool/utils"
)

func CreateImage() {
	ui.PrintLogo(UdcName, GetUSBSpeed(), CurrentMan, CurrentVID, "未使用", false)
	
	name := ui.ReadLine(fmt.Sprintf("  %s输入镜像名: %s", ui.P_LIGHT, ui.RESET))
	if name == "" {
		name = "STORE_DISK"
	}

	if !validateImageName(name) {
		fmt.Printf("\n  %s错误: 镜像名称只能包含字母、数字、下划线和连字符%s\n", ui.WARN, ui.RESET)
		ui.Pause()
		return
	}

	sizeStr := ui.ReadLine(fmt.Sprintf("  %s输入镜像大小 (MB) [512-8192]: %s", ui.P_LIGHT, ui.RESET))
	size := 1024
	if s, err := strconv.Atoi(sizeStr); err == nil {
		size = s
	}

	if size < 512 || size > 8192 {
		fmt.Printf("\n  %s错误: 镜像大小必须在 512-8192 MB 之间%s\n", ui.WARN, ui.RESET)
		ui.Pause()
		return
	}

	path := filepath.Join(SEARCH_DIR, name+".img")

	if _, err := os.Stat(path); err == nil {
		fmt.Printf("\n  %s镜像文件已存在: %s.img%s\n", ui.WARN, name, ui.RESET)
		confirm := ui.ReadLine(fmt.Sprintf("  %s是否覆盖? (y/N): %s", ui.WARN, ui.RESET))
		if strings.ToLower(confirm) != "y" {
			fmt.Printf("  %s取消创建%s\n", ui.G_TEXT, ui.RESET)
			ui.Pause()
			return
		}
		os.Remove(path)
	}

	// 检查空间逻辑略过，直接尝试创建

	fmt.Printf("\n  %s选择文件系统格式:%s\n", ui.P_LIGHT, ui.RESET)
	fmt.Printf("  %s[1] exFAT (推荐，支持大文件)%s\n", ui.P_LIGHT, ui.RESET)
	fmt.Printf("  %s[2] FAT32 (兼容性最好)%s\n", ui.P_LIGHT, ui.RESET)
	fmt.Printf("  %s[3] NTFS (Windows 原生)%s\n", ui.P_LIGHT, ui.RESET)
	
	fsChoice := ui.ReadLine("\n  # 选择 [1-3]: ")
	
	fsType := "exfat"
	fsName := "exFAT"
	
	switch fsChoice {
	case "2":
		fsType = "vfat"
		fsName = "FAT32"
	case "3":
		fsType = "ntfs"
		fsName = "NTFS"
	}

	clusterSize := "128K"
	if fsType == "vfat" {
		if size < 512 {
			clusterSize = "16K"
		} else if size < 1024 {
			clusterSize = "32K"
		} else {
			clusterSize = "64K"
		}
	} else if fsType == "exfat" {
		if size < 1024 {
			clusterSize = "64K"
		} else if size > 4096 {
			clusterSize = "256K"
		}
	} else if fsType == "ntfs" {
		clusterSize = "4K"
	}

	fmt.Printf("\n  %s▣ 启动高性能磁盘初始化流程...%s\n", ui.P_LIGHT, ui.RESET)
	utils.Log("INFO", "开始创建镜像: %s.img (%dMB, %s, 簇大小: %s)", name, size, fsName, clusterSize)

	fmt.Printf("  %s├─ 预分配 %dMB 物理空间... %s", ui.G_TEXT, size, ui.RESET)
	
	// Try fallocate first
	if err := exec.Command("fallocate", "-l", fmt.Sprintf("%dM", size), path).Run(); err != nil {
		// Fallback to dd
		if err := exec.Command("dd", "if=/dev/zero", "of="+path, "bs=1M", fmt.Sprintf("count=%d", size), "conv=notrunc").Run(); err != nil {
			fmt.Printf("%s[FAILED]%s\n", ui.WARN, ui.RESET)
			fmt.Printf("\n  %s错误: 预分配空间失败%s\n", ui.WARN, ui.RESET)
			os.Remove(path)
			ui.Pause()
			return
		}
	}
	fmt.Printf("%s[DONE]%s\n", ui.B_GREEN, ui.RESET)

	fmt.Printf("  %s└─ 格式化 %s (%s 簇对齐)... %s", ui.G_TEXT, fsName, clusterSize, ui.RESET)

	var cmd *exec.Cmd
	switch fsType {
	case "exfat":
		cmd = exec.Command("mkfs.exfat", "-s", clusterSize, path)
	case "vfat":
		cmd = exec.Command("mkfs.vfat", "-F", "32", "-s", clusterSize, path)
	case "ntfs":
		cmd = exec.Command("mkfs.ntfs", "-f", path)
	}

	if err := cmd.Run(); err != nil {
		fmt.Printf("%s[FAILED]%s\n", ui.WARN, ui.RESET)
		utils.Log("ERROR", "格式化失败: %v", err)
		os.Remove(path)
		ui.Pause()
		return
	}
	fmt.Printf("%s[DONE]%s\n", ui.B_GREEN, ui.RESET)
	utils.Log("SUCCESS", "镜像格式化完成: %s.img (%s)", name, fsName)

	time.Sleep(time.Second)
	MountLogic(path)
	ui.Pause()
}

func validateImageName(name string) bool {
	if name == "" {
		return false
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

func ListImages() {
	fmt.Printf("\n  %s▣ 镜像列表%s\n\n", ui.P_LIGHT, ui.RESET)
	entries, err := os.ReadDir(SEARCH_DIR)
	if err != nil {
		fmt.Printf("  %s无法读取目录%s\n", ui.WARN, ui.RESET)
		ui.Pause()
		return
	}

	found := false
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".img") {
			found = true
			info, _ := entry.Info()
			sizeMB := info.Size() / 1024 / 1024
			fmt.Printf("  %-25s %8d MB\n", entry.Name(), sizeMB)
		}
	}

	if !found {
		fmt.Printf("  %s暂无镜像文件%s\n", ui.WARN, ui.RESET)
	}
	ui.Pause()
}

func ManageImages() {
	for {
		ui.PrintLogo(UdcName, GetUSBSpeed(), CurrentMan, CurrentVID, "未使用", false)
		fmt.Printf("\n  %s▣ 镜像管理%s\n\n", ui.P_LIGHT, ui.RESET)
		fmt.Printf("  %s[1] 列出所有镜像%s\n", ui.P_LIGHT, ui.RESET)
		fmt.Printf("  %s[2] 删除镜像%s\n", ui.P_LIGHT, ui.RESET)
		fmt.Printf("  %s[3] 重命名镜像%s\n", ui.P_LIGHT, ui.RESET)
		fmt.Printf("  %s[0] 返回主菜单%s\n", ui.P_LIGHT, ui.RESET)
		
		choice := ui.ReadLine("\n  # 选择: ")

		switch choice {
		case "1":
			ListImages()
		case "2":
			deleteImage()
		case "3":
			renameImage()
		case "0":
			return
		}
	}
}

func deleteImage() {
	// 简化版，不列出列表
	name := ui.ReadLine(fmt.Sprintf("\n  %s输入要删除的镜像名: %s", ui.P_LIGHT, ui.RESET))
	path := filepath.Join(SEARCH_DIR, name+".img")
	if _, err := os.Stat(path); err == nil {
		confirm := ui.ReadLine(fmt.Sprintf("  %s确认删除 %s.img? (y/N): %s", ui.WARN, name, ui.RESET))
		if strings.ToLower(confirm) == "y" {
			os.Remove(path)
			fmt.Printf("  %s镜像已删除%s\n", ui.B_GREEN, ui.RESET)
			ui.Pause()
		}
	} else {
		fmt.Printf("  %s镜像不存在%s\n", ui.WARN, ui.RESET)
		ui.Pause()
	}
}

func renameImage() {
	oldName := ui.ReadLine(fmt.Sprintf("\n  %s输入旧镜像名: %s", ui.P_LIGHT, ui.RESET))
	oldPath := filepath.Join(SEARCH_DIR, oldName+".img")
	if _, err := os.Stat(oldPath); err != nil {
		fmt.Printf("  %s镜像不存在%s\n", ui.WARN, ui.RESET)
		ui.Pause()
		return
	}
	
	newName := ui.ReadLine(fmt.Sprintf("  %s输入新镜像名: %s", ui.P_LIGHT, ui.RESET))
	newPath := filepath.Join(SEARCH_DIR, newName+".img")
	if _, err := os.Stat(newPath); err == nil {
		fmt.Printf("  %s目标已存在%s\n", ui.WARN, ui.RESET)
		ui.Pause()
		return
	}
	
os.Rename(oldPath, newPath)
	fmt.Printf("  %s已重命名%s\n", ui.B_GREEN, ui.RESET)
	ui.Pause()
}
