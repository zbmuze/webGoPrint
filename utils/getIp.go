package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

// GetLocalIP ：获取本机非loopback IPv4地址
func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil { // 只返回IPv4
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("未找到有效IPv4地址")
}

func GetAllIPInfo() (map[string]string, error) {
	result := make(map[string]string)

	// 获取公网IPv4
	if v4, err := GetPublicIP(); err == nil {
		result["ipv4"] = v4
	}

	// 获取公网IPv6
	if v6, err := GetPublicIPv6JSON(); err == nil {
		result["ipv6"] = v6
	}

	return result, nil
}

// GetPublicIPv6JSON 获取公网IPv6（使用JSON API）
func GetPublicIPv6JSON() (string, error) {
	type Response struct {
		IP string `json:"ip"`
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api64.ipify.org?format=json")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.IP, nil
}

func GetPublicIP() (string, error) {
	// 使用一个提供纯文本IP返回的可靠服务
	// 常见的选择有：
	urls := []string{
		"https://api.ipify.org", // 最流行的选择
		"https://ident.me",
		"http://myexternalip.com/raw",
		"http://ipinfo.io/ip",
	}

	// 为HTTP客户端设置一个合理的超时时间
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var lastErr error
	// 尝试多个服务，直到一个成功为止
	for _, url := range urls {
		resp, err := client.Get(url)
		if err != nil {
			lastErr = err
			continue // 这个失败了，试试下一个
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				fmt.Printf("error closing body: %s\n", err)
			}
		}(resp.Body)

		ip, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}

		// 成功获取到IP
		return string(ip), nil
	}

	// 所有服务都尝试失败了
	return "", fmt.Errorf("所有获取公网IP的API请求均失败，最后一条错误信息: %v", lastErr)
}
