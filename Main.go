package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

type SystemInfo struct {
	Hostname    string `json:"hostname"`
	IP          string `json:"ip"`
	CPUModel    string `json:"cpu_model"`
	TotalMemory string `json:"total_memory"`
	UsedMemory  string `json:"used_memory"`
	Uptime      string `json:"uptime"`
	WiFi        string `json:"wifi"`
	Battery     string `json:"battery"`
	SSHInfo     string `json:"ssh_info"`
	Timestamp   string `json:"timestamp"`
}

type SystemInfoWrapper struct {
	System1Info SystemInfo `json:"Thangavi_info"`
}

func getCPUModel() (string, error) {
	cpuInfo, err := cpu.Info()
	if err != nil {
		return "", err
	}
	if len(cpuInfo) > 0 {
		return cpuInfo[0].ModelName, nil
	}
	return "Unknown", nil
}

func getMemoryInfo() (string, string, error) {
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return "", "", err
	}
	totalMem := fmt.Sprintf("%.2f GB", float64(vmStat.Total)/(1024*1024*1024))
	usedMem := fmt.Sprintf("%.2f GB (%.2f%%)", float64(vmStat.Used)/(1024*1024*1024), vmStat.UsedPercent)
	return totalMem, usedMem, nil
}

func getUptime() (string, error) {
	uptime, err := host.Uptime()
	if err != nil {
		return "", err
	}
	uptimeDuration := time.Duration(uptime) * time.Second
	return uptimeDuration.String(), nil
}

func getWiFiInfo() (string, error) {
	out, err := exec.Command("iwconfig").Output()
	if err != nil {
		return "", fmt.Errorf("WiFi information command not found")
	}
	wifiInfo := string(out)
	if strings.Contains(wifiInfo, "ESSID") {
		lines := strings.Split(wifiInfo, "\n")
		for _, line := range lines {
			if strings.Contains(line, "ESSID") || strings.Contains(line, "Signal level") {
				return line, nil
			}
		}
	}
	return "No WiFi info available", nil
}

func getBatteryInfo() (string, error) {
	out, err := exec.Command("upower", "-i", "/org/freedesktop/UPower/devices/battery_BAT0").Output()
	if err != nil {
		return "", fmt.Errorf("Battery information command not found")
	}

	lines := strings.Split(string(out), "\n")
	var batteryInfo []string
	var state, timeToEmpty, timeToFull, percentage string

	for _, line := range lines {
		line = strings.TrimSpace(line) // Clean up whitespace
		if strings.Contains(line, "state") {
			state = strings.ReplaceAll(line, " ", "") // Remove spaces
		} else if strings.Contains(line, "time to empty") {
			timeToEmpty = strings.TrimSpace(line)
		} else if strings.Contains(line, "time to full") {
			timeToFull = strings.TrimSpace(line)
		} else if strings.Contains(line, "percentage") {
			percentage = strings.ReplaceAll(line, " ", "") // Remove spaces
		}
	}

	batteryInfo = append(batteryInfo, state)
	if timeToFull != "" {
		batteryInfo = append(batteryInfo, timeToFull)
	} else if timeToEmpty != "" {
		batteryInfo = append(batteryInfo, timeToEmpty)
	}
	if percentage != "" {
		batteryInfo = append(batteryInfo, percentage)
	}

	if len(batteryInfo) > 0 {
		return strings.Join(batteryInfo, ", "), nil
	}
	return "No Battery info available", nil
}

func getSSHInfo() (string, error) {
	out, err := exec.Command("ss", "-tuna").Output()
	if err != nil {
		return "", fmt.Errorf("SSH command not found")
	}

	ssOut := string(out)
	var connectedIPs []string
	lines := strings.Split(ssOut, "\n")

	for _, line := range lines {
		if strings.Contains(line, ":22") && strings.Contains(line, "ESTAB") {
			parts := strings.Fields(line)
			if len(parts) >= 5 {
				remoteAddress := parts[len(parts)-1]
				ip := strings.Split(remoteAddress, ":")[0]
				if ip != "127.0.0.1" && ip != "localhost" {
					connectedIPs = append(connectedIPs, ip)
				}
			}
		}
	}

	if len(connectedIPs) > 0 {
		return fmt.Sprintf("Active SSH connections detected. Connected IPs: %s", strings.Join(connectedIPs, ", ")), nil
	}
	return "No active SSH connections", nil
}

// Function to get the system IP address
func getIPAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagLoopback == 0 {
			addrs, err := iface.Addrs()
			if err != nil {
				return "", err
			}
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				if ip != nil && ip.To4() != nil {
					return ip.String(), nil
				}
			}
		}
	}
	return "No IP address found", nil
}

func main() {
	// WebSocket connection setup
	wsURL := "ws://localhost:8080/WebSockCon/serverws" // Replace with your actual WebSocket server URL
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Fatal("Error connecting to WebSocket:", err)
	}
	defer conn.Close()

	// Infinite loop to update system information every 1 minute
	for {
		// Get Hostname
		hostInfo, err := host.Info()
		if err != nil {
			log.Fatal("Error getting host info:", err)
		}
		hostname := hostInfo.Hostname

		// Get IP Address
		ipAddress, err := getIPAddress()
		if err != nil {
			log.Fatal("Error getting IP address:", err)
		}

		// Get CPU model
		cpuModel, err := getCPUModel()
		if err != nil {
			log.Fatal("Error getting CPU info:", err)
		}

		// Get Memory info
		totalMem, usedMem, err := getMemoryInfo()
		if err != nil {
			log.Fatal("Error getting memory info:", err)
		}

		// Get System Uptime
		uptime, err := getUptime()
		if err != nil {
			log.Fatal("Error getting uptime info:", err)
		}

		// Get WiFi info
		wifi, err := getWiFiInfo()
		if err != nil {
			log.Println(err)
			wifi = "No WiFi info"
		}

		// Get Battery info
		battery, err := getBatteryInfo()
		if err != nil {
			log.Println(err)
			battery = "No Battery info"
		}

		// Get SSH info
		ssh, err := getSSHInfo()
		if err != nil {
			log.Println(err)
			ssh = "No SSH info"
		}

		// Get Current Time
		currentTime := time.Now().Format(time.RFC3339)

		// Create a struct for system information
		sysInfo := SystemInfo{
			Hostname:    hostname,
			IP:          ipAddress,
			CPUModel:    cpuModel,
			TotalMemory: totalMem,
			UsedMemory:  usedMem,
			Uptime:      uptime,
			WiFi:        wifi,
			Battery:     battery,
			SSHInfo:     ssh,
			Timestamp:   currentTime,
		}

		// Wrap system info inside SystemInfoWrapper with the "system1_info" key
		wrappedInfo := SystemInfoWrapper{
			System1Info: sysInfo,
		}

		// Convert system information to JSON
		jsonData, err := json.Marshal(wrappedInfo)
		if err != nil {
			log.Fatal("Error marshalling system info to JSON:", err)
		}

		// Send the JSON data over WebSocket
		err = conn.WriteMessage(websocket.TextMessage, jsonData)
		if err != nil {
			log.Println("Error sending message:", err)
		}

		// Wait for 1 minute before sending the next update
		time.Sleep(1 * time.Second)
	}
}
