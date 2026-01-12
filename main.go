package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"usb_tool/core"
	"usb_tool/ui"
	"usb_tool/utils"
)

func main() {
	if !utils.CheckRoot() {
		fmt.Println("错误: 需要 ROOT 权限")
		return
	}

	if err := utils.AcquireLock(); err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	defer utils.ReleaseLock()

	// 初始化核心逻辑 (包含日志初始化)
	ui.DetectScreenSize()
	if err := core.InitCore(); err != nil {
		fmt.Printf("初始化失败: %v\n", err)
		return
	}
	defer utils.CloseLog()

	for {
		ui.PrintLogo(core.UdcName, core.GetUSBSpeed(), core.CurrentMan, core.CurrentVID, "已启用", core.CurrentLoopDevice != "")
		ui.ShowMainMenu(core.CurrentMan)

		choice := ui.ReadLine("\n  # 选择: ")

		switch choice {
		case "1":
			entries, _ := os.ReadDir(core.SEARCH_DIR)
			var images []string
			for _, e := range entries {
				if strings.HasSuffix(e.Name(), ".img") {
					images = append(images, e.Name())
				}
			}

			if len(images) == 0 {
				fmt.Printf("\n  %s暂无镜像文件%s\n", ui.WARN, ui.RESET)
				ui.Pause()
				continue
			}

			fmt.Println()
			for i, img := range images {
				fmt.Printf("  [%d] %s\n", i+1, img)
			}
			
			idxStr := ui.ReadLine("\n  # 选择镜像: ")
			var idx int
			fmt.Sscanf(idxStr, "%d", &idx)
			if idx > 0 && idx <= len(images) {
				core.MountLogic(core.SEARCH_DIR + "/" + images[idx-1])
			}

		case "2":
			core.CreateImage()

		case "3":
			// 简单的方案选择逻辑，可以下沉但目前放在这也不算太乱
			fmt.Printf("\n  [1] SanDisk [2] Kingston [3] Samsung [4] Sony [5] Custom\n")
			c := ui.ReadLine("  # 选择: ")
			switch c {
			case "1":
				core.CurrentVID, core.CurrentPID, core.CurrentMan, core.CurrentProd = "0x0781", "0x5581", "SanDisk", "Ultra Fast"
			case "2":
				core.CurrentVID, core.CurrentPID, core.CurrentMan, core.CurrentProd = "0x0951", "0x1666", "Kingston", "DT100"
			case "3":
				core.CurrentVID, core.CurrentPID, core.CurrentMan, core.CurrentProd = "0x04E8", "0x61F5", "Samsung", "Portable SSD"
			case "4":
				core.CurrentVID, core.CurrentPID, core.CurrentMan, core.CurrentProd = "0x054C", "0x05BA", "Sony", "Media Drive"
			default:
				core.CurrentVID, core.CurrentPID, core.CurrentMan, core.CurrentProd = "0x18d1", "0x4ee7", "Google", "Pixel Disk"
			}
			core.SaveSpoofConfig()

		case "4":
			core.ManageImages()

		case "5":
			core.ShowStatus()

		case "6":
			core.ShowHelp()

		case "7":
			core.ReenumerateUSB()

		case "8":
			core.DisconnectUSB()

		case "0":
			core.RestoreSelinuxState()
			if core.CurrentLoopDevice != "" {
				os.WriteFile(core.ConfigDir+"/UDC", []byte(""), 0644)
				exec.Command("losetup", "-d", core.CurrentLoopDevice).Run()
			}
			exec.Command("setprop", "sys.usb.config", "mtp,adb").Run()
			fmt.Println("\n  再见！")
			return

		default:
			fmt.Printf("\n  %s无效的选择%s\n", ui.WARN, ui.RESET)
			ui.Pause()
		}
	}
}