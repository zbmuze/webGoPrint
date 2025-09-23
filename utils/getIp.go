package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// GetAllIPs 获取所有可用IP的调试函数
func GetAllIPs() ([]string, error) {
	localip, err := GetLocalIP()
	ipv4, _ := GetPublicIP()
	ipv6, _ := GetPublicIPv6JSON()
	return []string{localip, ipv4, ipv6}, err
}

// GetLocalIP ：获取本机非loopback IPv4地址
func GetLocalIP() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	// 定义优先级：有线 > 无线 > 其他
	var ethernetIP, wifiIP, otherIP string

	for _, iface := range interfaces {
		// 跳过回环接口和未启用的接口
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
				ip := ipnet.IP.String()

				// 过滤虚拟IP和特殊地址
				if isVirtualIP(ip) {
					continue
				}

				// 根据接口类型分类
				switch getInterfaceType(iface.Name) {
				case "ethernet":
					if ethernetIP == "" {
						ethernetIP = ip
					}
				case "wifi":
					if wifiIP == "" {
						wifiIP = ip
					}
				default:
					if otherIP == "" {
						otherIP = ip
					}
				}
			}
		}
	}

	// 按优先级返回IP
	if ethernetIP != "" {
		return ethernetIP, nil
	}
	if wifiIP != "" {
		return wifiIP, nil
	}
	if otherIP != "" {
		return otherIP, nil
	}

	return "", fmt.Errorf("未找到有效IPv4地址")
}

func getInterfaceType(name string) string {
	name = strings.ToLower(name)

	// Windows 接口名称匹配
	if strings.Contains(name, "ethernet") ||
		strings.Contains(name, "eth") ||
		strings.Contains(name, "本地连接") ||
		strings.Contains(name, "lan") {
		return "ethernet"
	}

	// WiFi 接口名称匹配
	if strings.Contains(name, "wi-fi") ||
		strings.Contains(name, "wifi") ||
		strings.Contains(name, "wireless") ||
		strings.Contains(name, "wlan") ||
		strings.Contains(name, "无线网络") {
		return "wifi"
	}

	// Linux 接口名称匹配
	if strings.HasPrefix(name, "eth") ||
		strings.HasPrefix(name, "en") {
		return "ethernet"
	}
	if strings.HasPrefix(name, "wlan") ||
		strings.HasPrefix(name, "wl") ||
		strings.HasPrefix(name, "wlp") {
		return "wifi"
	}

	return "other"
}

func isVirtualIP(ip string) bool {
	// 排除虚拟网络、链路本地、Docker等特殊地址
	virtualPrefixes := []string{
		"169.254.", // 链路本地地址 (APIPA)
		"192.168.100.",
		"198.18.0.",
		"192.168.122.",    // libvirt
		"172.17.",         // Docker默认
		"172.18.",         // Docker网络
		"172.19.",         // Docker网络
		"172.20.",         // Docker网络
		"10.0.",           // 常见虚拟网络
		"0.0.0.0",         // 无效地址
		"255.255.255.255", // 广播地址
	}

	// 检查私有地址段（但不过滤所有私有地址，因为正常网络也在私有段）
	for _, prefix := range virtualPrefixes {
		if strings.HasPrefix(ip, prefix) {
			return true
		}
	}

	return false
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
		"http://ipinfo.io/ip",
		"https://api.ipify.org", // 最流行的选择
		"https://ident.me",
		"http://myexternalip.com/raw",
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
