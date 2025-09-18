package utils

import (
	"bytes"
	"os/exec"
	"runtime"
	"strings"
)

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
