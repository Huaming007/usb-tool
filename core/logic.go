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

func MountLogic(target string) {
	if _, err := os.Stat(target); err != nil {
		fmt.Printf("\n  %s错误: 镜像文件不存在: %s%s\n", ui.WARN, target, ui.RESET)
		ui.Pause()
		return
	}

	utils.Log("INFO", "开始挂载镜像: %s", target)
	ui.PrintLogo(UdcName, GetUSBSpeed(), CurrentMan, CurrentVID, "处理中...", false)
	fmt.Printf("\n  %s▣ 启动深度注入与调试序列...%s\n", ui.P_LIGHT, ui.RESET)

	// 1. 停用系统 USB 服务
	fmt.Printf("  %s├─ 阻塞系统调度 (SELinux)... %s", ui.G_TEXT, ui.RESET)
	exec.Command("setenforce", "0").Run()
	exec.Command("setprop", "sys.usb.config", "none").Run()
	exec.Command("setprop", "sys.usb.configfs", "0").Run()
	exec.Command("setprop", "sys.usb.ffs.aio_compat", "0").Run()
	os.WriteFile(filepath.Join(ConfigDir, "UDC"), []byte("\n"), 0644)
	time.Sleep(2 * time.Second)
	fmt.Printf("%s[DOWN]%s\n", ui.B_GREEN, ui.RESET)

	// 2. 清理 ConfigFS
	fmt.Printf("  %s├─ 销毁并重建功能描述符... %s", ui.G_TEXT, ui.RESET)
	
			// Fix: 必须先断开 ConfigFS 的链接，才能删除 function 目录
			configPath := filepath.Join(ConfigDir, "configs/b.1")		// 模拟 Shell: find "$CONFIG_DIR/configs/b.1/" -type l -exec rm -f {} +
		if entries, err := os.ReadDir(configPath); err == nil {
			for _, entry := range entries {
				os.Remove(filepath.Join(configPath, entry.Name()))
			}
		}
		
	os.RemoveAll(filepath.Join(ConfigDir, "functions/mass_storage.0"))
	os.RemoveAll(filepath.Join(ConfigDir, "strings/0x409"))
	os.MkdirAll(filepath.Join(ConfigDir, "strings/0x409"), 0755)
	fmt.Printf("%s[CLEANED]%s\n", ui.B_GREEN, ui.RESET)

	// 3. 注入 ID
	fmt.Printf("  %s├─ 注入 ID [%s%s:%s%s]... %s", ui.G_TEXT, ui.B_CYAN, CurrentVID, CurrentPID, ui.G_TEXT, ui.RESET)
	os.WriteFile(filepath.Join(ConfigDir, "idVendor"), []byte(CurrentVID+"\n"), 0644)
	os.WriteFile(filepath.Join(ConfigDir, "idProduct"), []byte(CurrentPID+"\n"), 0644)
	os.WriteFile(filepath.Join(ConfigDir, "bcdUSB"), []byte("0x0320\n"), 0644) // USB 3.1
	os.WriteFile(filepath.Join(ConfigDir, "bDeviceClass"), []byte("0x00\n"), 0644)
	os.WriteFile(filepath.Join(ConfigDir, "bDeviceSubClass"), []byte("0x00\n"), 0644)
	os.WriteFile(filepath.Join(ConfigDir, "bDeviceProtocol"), []byte("0x00\n"), 0644)
	os.WriteFile(filepath.Join(ConfigDir, "bcdDevice"), []byte("0x0100\n"), 0644)
	os.WriteFile(filepath.Join(ConfigDir, "bMaxPacketSize0"), []byte("0x40\n"), 0644)
	fmt.Printf("%s[OK]%s\n", ui.B_GREEN, ui.RESET)

	fmt.Printf("  %s├─ 覆写描述符 [%s%s%s]... %s", ui.G_TEXT, ui.B_CYAN, CurrentMan, ui.G_TEXT, ui.RESET)
	os.WriteFile(filepath.Join(ConfigDir, "strings/0x409/manufacturer"), []byte(CurrentMan+"\n"), 0644)
	os.WriteFile(filepath.Join(ConfigDir, "strings/0x409/product"), []byte(CurrentProd+"\n"), 0644)
	os.WriteFile(filepath.Join(ConfigDir, "strings/0x409/serialnumber"), []byte(CurrentSer+"\n"), 0644)
	fmt.Printf("%s[OK]%s\n", ui.B_GREEN, ui.RESET)

	// 4. 绑定存储与 Loop 优化
	fmt.Printf("  %s├─ 建立存储函数与 LUN 映射... %s\n", ui.G_TEXT, ui.RESET)
	funcPath := filepath.Join(ConfigDir, "functions/mass_storage.0")
	os.MkdirAll(filepath.Join(funcPath, "lun.0"), 0755)

	os.WriteFile(filepath.Join(funcPath, "lun.0/nofua"), []byte("1\n"), 0644)
	os.WriteFile(filepath.Join(funcPath, "lun.0/removable"), []byte("1\n"), 0644)
	os.WriteFile(filepath.Join(funcPath, "stall"), []byte("1\n"), 0644)
	if _, err := os.Stat(filepath.Join(funcPath, "num_buffers")); err == nil {
		os.WriteFile(filepath.Join(funcPath, "num_buffers"), []byte("4\n"), 0644)
	}

	useLoop := false
	loopDevice := ""

	// Loop 挂载 (优化逻辑)
	if _, err := os.Stat("/dev/loop-control"); err == nil {
		fmt.Printf("  %s│  ├─ 挂载 loop 设备... %s", ui.G_TEXT, ui.RESET)
		// 使用 --show 直接获取挂载后的设备名，避免竞态
		cmd := exec.Command("losetup", "-fP", "--show", target)
		if out, err := cmd.Output(); err == nil {
			dev := strings.TrimSpace(string(out))
			if dev != "" {
				loopDevice = dev
				useLoop = true
				CurrentLoopDevice = dev
				fmt.Printf("%s[OK]%s\n", ui.B_GREEN, ui.RESET)

				// 优化 Loop
				fmt.Printf("  %s│  └─ 优化块设备参数... %s", ui.G_TEXT, ui.RESET)
				loopNum := strings.TrimPrefix(dev, "/dev/")
				os.WriteFile(fmt.Sprintf("/sys/block/%s/queue/logical_block_size", loopNum), []byte("4096\n"), 0644)
				os.WriteFile(fmt.Sprintf("/sys/block/%s/queue/nr_requests", loopNum), []byte("128\n"), 0644)
				os.WriteFile(fmt.Sprintf("/sys/block/%s/queue/scheduler", loopNum), []byte("noop\n"), 0644)
				os.WriteFile(fmt.Sprintf("/sys/block/%s/queue/read_ahead_kb", loopNum), []byte("256\n"), 0644)
				fmt.Printf("%s[DONE]%s\n", ui.B_GREEN, ui.RESET)
			}
		} else {
			// 如果新版 losetup 失败，尝试旧方法
			if out, err := exec.Command("losetup", "-f").Output(); err == nil {
				dev := strings.TrimSpace(string(out))
				if dev != "" && exec.Command("losetup", dev, target).Run() == nil {
					loopDevice = dev
					useLoop = true
					CurrentLoopDevice = dev
					fmt.Printf("%s[OK (legacy)]%s\n", ui.B_GREEN, ui.RESET)
				} else {
					fmt.Printf("%s[FAILED]%s\n", ui.WARN, ui.RESET)
				}
			} else {
				fmt.Printf("%s[FAILED]%s\n", ui.WARN, ui.RESET)
			}
		}
	}

	// 绑定文件
	if useLoop {
		if err := os.WriteFile(filepath.Join(funcPath, "lun.0/file"), []byte(loopDevice+"\n"), 0644); err != nil {
			utils.Log("WARN", "Loop 绑定失败，回退文件模式")
			exec.Command("losetup", "-d", loopDevice).Run()
			CurrentLoopDevice = ""
			useLoop = false
		}
	}
	if !useLoop {
		os.WriteFile(filepath.Join(funcPath, "lun.0/file"), []byte(target+"\n"), 0644)
	}

	// 软链接
	configPath = filepath.Join(ConfigDir, "configs/b.1")
	os.MkdirAll(configPath, 0755)
	os.Symlink(funcPath, filepath.Join(configPath, "f1"))

	// 5. SCSI 标识
	fmt.Printf("  %s├─ SCSI 查询标识注入探测... %s", ui.G_TEXT, ui.RESET)
	
	// 准备基础字符串（去除空格，防止截断问题）
	vendor := strings.ReplaceAll(CurrentMan, " ", "_")
	if len(vendor) > 8 { vendor = vendor[:8] }

	prod := strings.ReplaceAll(CurrentProd, " ", "_")
	if len(prod) > 16 { prod = prod[:16] }
	
	rev := "1.00"

	// 路径定义
	vendorPath := filepath.Join(funcPath, "lun.0/vendor")
	inqPath := filepath.Join(funcPath, "lun.0/inquiry_string")

	if _, err := os.Stat(vendorPath); err == nil {
		// 方案 A: 现代内核，支持独立属性
		// 注意：不加换行符，严格匹配 Shell 的 printf 行为
		os.WriteFile(vendorPath, []byte(vendor), 0644)
		os.WriteFile(filepath.Join(funcPath, "lun.0/model"), []byte(prod), 0644)
		os.WriteFile(filepath.Join(funcPath, "lun.0/rev"), []byte(rev), 0644)
		fmt.Printf("%s[SUCCESS: SPLIT]%s\n", ui.B_GREEN, ui.RESET)
	} else if _, err := os.Stat(inqPath); err == nil {
		// 方案 B: 旧版内核，使用 inquiry_string
		// 格式：Vendor(8) + Product(16) + Rev(4)，左对齐空格填充
		fullInquiry := fmt.Sprintf("%-8s%-16s%-4s", vendor, prod, rev)
		os.WriteFile(inqPath, []byte(fullInquiry), 0644)
		fmt.Printf("%s[SUCCESS: COMBINED]%s\n", ui.B_GREEN, ui.RESET)
	} else {
		// 无法识别的内核接口
		fmt.Printf("%s[SKIPPED]%s\n", ui.WARN, ui.RESET)
		utils.Log("WARN", "无法找到 SCSI 标识注入点 (vendor/inquiry_string 均不存在)")
	}

	// 6. 激活 UDC
	fmt.Printf("  %s└─ 激活 UDC 控制器总线序列... %s", ui.G_TEXT, ui.RESET)
	
	// Check config integrity
	if _, err := os.Lstat(filepath.Join(configPath, "f1")); err != nil {
		fmt.Printf("%s[CONFIG ERROR]%s\n", ui.WARN, ui.RESET)
		utils.Log("ERROR", "ConfigFS 配置不完整")
		return
	}

	if err := os.WriteFile(filepath.Join(ConfigDir, "UDC"), []byte(UdcName+"\n"), 0644); err == nil {
		time.Sleep(2 * time.Second)
		speed := GetUSBSpeed()
		if speed != "UNKNOWN" {
			fmt.Printf("%s[ACTIVE]%s\n", ui.B_GREEN, ui.RESET)
			fmt.Printf("\n  %s✔ 伪装生效！当前速率: %s%s\n", ui.B_CYAN, speed, ui.RESET)
		} else {
			fmt.Printf("%s[NO SPEED]%s\n", ui.WARN, ui.RESET)
			fmt.Printf("\n  %s⚠ UDC 已激活但无法检测速率%s\n", ui.B_CYAN, ui.RESET)
		}
	} else {
		fmt.Printf("%s[FAILED]%s\n", ui.WARN, ui.RESET)
		fmt.Printf("\n  %sUDC 激活失败，可能是未连接 USB 数据线%s\n", ui.G_TEXT, ui.RESET)
	}
	ui.Pause()
}

func DisconnectUSB() {
	ui.PrintLogo(UdcName, GetUSBSpeed(), CurrentMan, CurrentVID, "卸载中...", CurrentLoopDevice != "")
	fmt.Printf("\n  %s▣ 正在执行安全卸载序列...%s\n", ui.P_LIGHT, ui.RESET)

	fmt.Printf("  %s├─ 复位 UDC 控制器连接... %s", ui.G_TEXT, ui.RESET)
	os.WriteFile(filepath.Join(ConfigDir, "UDC"), []byte("\n"), 0644)
	time.Sleep(1 * time.Second)
	fmt.Printf("%s[OK]%s\n", ui.B_GREEN, ui.RESET)

	if CurrentLoopDevice != "" {
		fmt.Printf("  %s├─ 释放 loop 设备... %s", ui.G_TEXT, ui.RESET)
		if err := exec.Command("losetup", "-d", CurrentLoopDevice).Run(); err == nil {
			fmt.Printf("%s[OK]%s\n", ui.B_GREEN, ui.RESET)
			CurrentLoopDevice = ""
		} else {
			fmt.Printf("%s[FAILED]%s\n", ui.WARN, ui.RESET)
		}
	}

	// 尝试清理 mass_storage 链接，帮助系统重置状态
	os.Remove(filepath.Join(ConfigDir, "configs/b.1/f1"))

	RestoreSelinuxState()
	
	fmt.Printf("  %s└─ 恢复系统 MTP 接口... %s", ui.G_TEXT, ui.RESET)
	exec.Command("setprop", "sys.usb.config", "mtp,adb").Run()
	fmt.Printf("%s[SUCCESS]%s\n", ui.B_GREEN, ui.RESET)
	
	ui.Pause()
}

func ReenumerateUSB() {
	ui.PrintLogo(UdcName, GetUSBSpeed(), CurrentMan, CurrentVID, "重枚举...", CurrentLoopDevice != "")
	fmt.Printf("\n  %s▣ 重新枚举 USB 设备%s\n", ui.P_LIGHT, ui.RESET)

	fmt.Printf("  %s├─ 断开当前连接... %s", ui.G_TEXT, ui.RESET)
	os.WriteFile(filepath.Join(ConfigDir, "UDC"), []byte("\n"), 0644)
	fmt.Printf("%s[OK]%s\n", ui.B_GREEN, ui.RESET)

	fmt.Printf("  %s├─ 等待系统识别断开... %s", ui.G_TEXT, ui.RESET)
	time.Sleep(2 * time.Second)
	fmt.Printf("%s[OK]%s\n", ui.B_GREEN, ui.RESET)

	fmt.Printf("  %s└─ 重新激活 UDC... %s", ui.G_TEXT, ui.RESET)
	if err := os.WriteFile(filepath.Join(ConfigDir, "UDC"), []byte(UdcName+"\n"), 0644); err == nil {
		time.Sleep(time.Second)
		fmt.Printf("%s[ACTIVE]%s\n", ui.B_GREEN, ui.RESET)
		fmt.Printf("\n  %s✔ 设备已重新枚举！速率: %s%s\n", ui.B_CYAN, GetUSBSpeed(), ui.RESET)
	} else {
		fmt.Printf("%s[FAILED]%s\n", ui.WARN, ui.RESET)
	}
	ui.Pause()
}

func ShowStatus() {
	ui.PrintLogo(UdcName, GetUSBSpeed(), CurrentMan, CurrentVID, "监控中", CurrentLoopDevice != "")
	fmt.Printf("\n  %s▣ 设备状态%s\n\n", ui.P_LIGHT, ui.RESET)
	
	// ConfigFS Status
	funcPath := filepath.Join(ConfigDir, "functions/mass_storage.0")
	if _, err := os.Stat(funcPath); err == nil {
		fmt.Printf("  %s存储功能:%s %s已激活%s\n", ui.G_TEXT, ui.RESET, ui.B_GREEN, ui.RESET)
		if lunData, err := os.ReadFile(filepath.Join(funcPath, "lun.0/file")); err == nil {
			file := strings.TrimSpace(string(lunData))
			if strings.HasPrefix(file, "/dev/loop") {
				fmt.Printf("  %s挂载模式:%s %sLoop 设备%s %s(%s)%s\n", ui.G_TEXT, ui.RESET, ui.B_GREEN, ui.RESET, ui.B_CYAN, filepath.Base(file), ui.RESET)
			} else {
				fmt.Printf("  %s挂载模式:%s %s直接文件%s %s(%s)%s\n", ui.G_TEXT, ui.RESET, ui.WARN, ui.RESET, ui.B_CYAN, filepath.Base(file), ui.RESET)
			}
		}
	} else {
		fmt.Printf("  %s存储功能:%s %s未激活%s\n", ui.G_TEXT, ui.RESET, ui.WARN, ui.RESET)
	}

	// SELinux
	out, _ := exec.Command("getenforce").Output()
	status := strings.TrimSpace(string(out))
	color := ui.B_CYAN
	if status == "Enforcing" {
		color = ui.B_GREEN
	} else if status == "Permissive" {
		color = ui.WARN
	}
	fmt.Printf("  %sSELinux:%s %s%s%s\n", ui.G_TEXT, ui.RESET, color, status, ui.RESET)

	fmt.Printf("\n  %sUSB 伪装信息:%s\n", ui.G_TEXT, ui.RESET)
	fmt.Printf("  %s  厂商: %s%s%s\n", ui.G_TEXT, ui.B_CYAN, CurrentMan, ui.RESET)
	fmt.Printf("  %s  产品: %s%s%s\n", ui.G_TEXT, ui.B_CYAN, CurrentProd, ui.RESET)
	fmt.Printf("  %s  VID:PID: %s%s:%s%s\n", ui.G_TEXT, ui.B_CYAN, CurrentVID, CurrentPID, ui.RESET)

	ui.Pause()
}

func ShowHelp() {
	ui.PrintLogo(UdcName, GetUSBSpeed(), CurrentMan, CurrentVID, "帮助", CurrentLoopDevice != "")
	fmt.Printf("\n  %s▣ 使用帮助%s\n", ui.P_LIGHT, ui.RESET)
	// Simplified help text
	fmt.Printf("\n  %s[1] 选择镜像挂载%s - 挂载 .img 文件\n", ui.P_LIGHT, ui.RESET)
	fmt.Printf("  %s[2] 新建自定义镜像%s - 创建 exFAT/FAT32/NTFS 镜像\n", ui.P_LIGHT, ui.RESET)
	fmt.Printf("  %s[3] 方案选择%s - 伪装不同厂商\n", ui.P_LIGHT, ui.RESET)
	fmt.Printf("  %s[8] 安全卸载%s - 恢复 MTP 模式 (重要!)\n", ui.P_LIGHT, ui.RESET)
	ui.Pause()
}