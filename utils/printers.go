package utils

import (
	"bytes"
	"fmt"
	"os/exec"
	"print-server/global"
	"runtime"
	"strings"
)

// PrintDocument ：跨平台打印文件（Windows/macOS/Linux）
func PrintDocument(filePath string) error {
	var cmd *exec.Cmd
	switch strings.ToLower(runtime.GOOS) {
	case "windows":
		cmd = exec.Command("print", filePath) // Windows打印命令
	case "darwin": // macOS
		cmd = exec.Command("lpr", filePath) // macOS打印命令
	default: // Linux
		//   -d <打印机名称>：指定要使用的打印机。
		//   -n <副本数>：指定打印份数。
		//   -o <选项>：指定打印选项，如双面打印、彩色打印等。
		//   -q <队列名称>：将打印任务添加到指定的打印队列。
		fmt.Printf("打印文件 %s,大小 %s，方向 %s 打印机 %s", filePath, global.PageSize, global.Orientation, global.Printer)
		cmd = exec.Command("lp", "-d", "Virtual_PDF_Printer", filePath) // Linux打印命令
	}
	return cmd.Run()
}

// GetPrinters 获取系统打印机列表
func GetPrinters() ([]string, error) {
	switch strings.ToLower(runtime.GOOS) {
	case "windows":
		return getWindowsPrinters()
	case "darwin": // macOS
		return getMacPrinters()
	default: // Linux
		return getLinuxPrinters()
	}
}

// getWindowsPrinters 获取Windows打印机列表
func getWindowsPrinters() ([]string, error) {
	cmd := exec.Command("wmic", "printer", "get", "name")
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	// 解析输出
	lines := strings.Split(out.String(), "\n")
	var printers []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && line != "Name" {
			printers = append(printers, line)
		}
	}

	return printers, nil
}

// getMacPrinters 获取macOS打印机列表
func getMacPrinters() ([]string, error) {
	cmd := exec.Command("lpstat", "-p")
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	// 解析输出
	lines := strings.Split(out.String(), "\n")
	var printers []string
	for _, line := range lines {
		if strings.HasPrefix(line, "printer") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				// 移除 "printer" 前缀和状态信息
				printerName := parts[1]
				printers = append(printers, printerName)
			}
		}
	}

	return printers, nil
}

// getLinuxPrinters 获取Linux打印机列表
func getLinuxPrinters() ([]string, error) {
	cmd := exec.Command("lpstat", "-a")
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	// 解析输出
	lines := strings.Split(out.String(), "\n")
	var printers []string
	for _, line := range lines {
		if line != "" {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				printers = append(printers, parts[0])
			}
		}
	}

	return printers, nil
}
