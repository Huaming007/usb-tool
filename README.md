# USB Tool v9.0 (Go Edition)

[![Go Version](https://img.shields.io/github/go-mod/go-version/Huaming007/usb-tool)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Android-green)](https://android.com)

**USB Tool** 是一个为 Android 设备（Rooted）设计的高性能 USB Mass Storage (UMS) 伪装与挂载工具。

相比传统的 Shell 脚本版本，v9.0 采用 **Go 语言**重写，带来了毫秒级的启动速度、更安全的资源管理以及自动化的环境初始化功能。

## ✨ 核心特性

- **🚀 极致性能**: 静态编译的 Go 二进制文件，零依赖，启动瞬间完成。
- **🛡️ 自动环境初始化**: 首次运行自动创建 `/sdcard/ISO` 工作目录，无需手动配置。
- **⚡ USB 3.1 协议**: 强制开启 USB 3.1 协议协商，榨干传输带宽。
- **💾 多格式镜像支持**: 支持挂载和创建 exFAT (推荐)、FAT32、NTFS 格式镜像。
- **🎭 硬件伪装**: 一键模拟 Kingston, SanDisk, Samsung, Sony 等品牌指纹。
- **🔧 智能 Loop 挂载**: 自动优化 IO 调度器 (noop) 和预读缓存 (256KB)，大幅降低 CPU 占用。
- **📱 完美 UI 适配**: 针对 ADB Shell 和 Termux 进行屏幕宽度自适应和中文对齐优化。

## 📦 快速开始

### 前置要求

*   Android 7.0+ 设备
*   **Root 权限** (必须)
*   内核支持 ConfigFS (绝大多数现代手机都支持)

### 安装与运行

1.  从 [Releases](../../releases) 下载最新的 `usb_tool_android_arm64`。
2.  推送到手机（建议路径 `/data/local/tmp`）：

    ```bash
    adb push usb_tool_android_arm64 /data/local/tmp/usb_tool
    adb shell chmod +x /data/local/tmp/usb_tool
    ```

3.  启动工具：

    ```bash
    adb shell su -c "/data/local/tmp/usb_tool"
    ```

### 首次使用

1.  工具启动后，会自动检测并创建 `/sdcard/ISO` 目录。
2.  将你的 `.img` 镜像文件放入该目录。
3.  或者，使用菜单中的 `[2] 新建自定义镜像` 直接在手机上创建。

## 🛠️ 编译指南

如果你想自己编译本项目：

```bash
# 克隆仓库
git clone https://github.com/Huaming007/usb-tool.git
cd usb-tool

# 交叉编译 (Android ARM64)
CGO_ENABLED=0 GOOS=android GOARCH=arm64 go build -ldflags="-s -w" -o usb_tool_android_arm64

# 编译 (Linux AMD64)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o usb_tool_linux_amd64
```

## 📝 目录结构

*   `core/`: 核心逻辑 (ConfigFS 操作, 挂载, 镜像管理)
*   `ui/`: 终端界面渲染与交互
*   `utils/`: 通用工具 (日志, 锁文件)
*   `main.go`: 程序入口

## 🤝 贡献

欢迎提交 Issue 或 Pull Request 来改进这个项目！

## 👥 开发团队

- **开发者/维护者**: [Huaming007](https://github.com/Huaming007)
- **人工智能助手**: Gemini CLI Agent (由 Google 提供技术支持)

## 📄 许可证

本项目采用 [MIT 许可证](LICENSE)。
