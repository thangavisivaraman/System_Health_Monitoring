package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

type SystemInfo struct {
	DiskUsage      string `json:diskusage`
	Bluetoothuse   string `json:bluetoothuse`
	OsName         string `json:"OperatingSystem"`
	HardwareModel  string `json:"HardwareModel"`
	HardwareVendor string `json:HardwareVendor`
	Firewallstatus string `json:firewallstatus`
	NmapScan       string `json:"nmap_scan"`
	Hostname       string `json:"hostname"`
	IP             string `json:"ip"`
	CPUModel       string `json:"cpu_model"`
	TotalMemory    string `json:"total_memory"`
	UsedMemory     string `json:"used_memory"`
	Uptime         string `json:"uptime"`
	WiFi           string `json:"wifi"`
	Battery        string `json:"battery"`
	SSHInfo        string `json:"ssh_info"`
	Timestamp      string `json:"timestamp"`
}

type SystemInfoWrapper struct {
	System1Info SystemInfo `json:"Thangavi"`
}

func gethardwareModel() (string, error) {
	out, err := exec.Command("sudo", "dmidecode", "-s", "system-product-name").Output()
	if err != nil {
		return "", fmt.Errorf("hardwareModel information command not found")
	}
	hardwareModel := strings.TrimSpace(string(out))
	return hardwareModel, nil

}

func getOSType() (string, error) {
	out, err := exec.Command("hostnamectl").Output()
	if err != nil {
		return "", fmt.Errorf("OS information command not found")
	}

	osdetail := string(out)
	if strings.Contains(osdetail, "Operating System") {
		lines := strings.Split(osdetail, "\n")
		for _, line := range lines {
			if strings.Contains(line, "Operating System") {
				parts := strings.Split(line, ":")
				//fmt.Printf(parts[0])
				//fmt.Printf(parts[1])
				if len(parts) > 1 {
					// Return only the value part, trimming any extra spaces
					return strings.TrimSpace(parts[1]), nil
				}
			}
		}
	}

	return "No OS info available", nil
}
func getfirewall() (string, error) {
	out, err := exec.Command("sudo", "ufw", "status").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("firewall not giving status: %v", err)
	}

	status := string(out)
	if strings.Contains(status, "Status") {
		// Split the output into lines
		lines := strings.Split(status, "\n")
		for _, line := range lines {
			// Check for the line containing "Status"
			if strings.HasPrefix(line, "Status:") {
				// Split the line by ":"
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					return strings.TrimSpace(parts[1]), nil // Trim spaces around the value
				}
			}
		}
	}

	return "No status info available", nil
}

func getVendorType() (string, error) {
	out, err := exec.Command("sudo", "dmidecode", "-s", "system-manufacturer").Output()
	if err != nil {
		return "", fmt.Errorf("vendortype information command not found")
	}
	vendortype := strings.TrimSpace(string(out))
	return vendortype, nil
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
func getblueusage() (string, error) {
	// Run the hciconfig command to check the Bluetooth status
	cmd := exec.Command("hciconfig")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "Error checking Bluetooth", err
	}

	// Convert output to string
	output := out.String()

	// Check if there is a Bluetooth device present (look for hci interface)
	if !strings.Contains(output, "hci") {
		return "NO BLUETOOTH DEVICE", nil // No Bluetooth device found
	}

	// Check if Bluetooth is DOWN (off)
	if strings.Contains(output, "DOWN") {
		return "OFF", nil // Bluetooth is off
	}

	// If not DOWN, consider Bluetooth as ON
	return "ON", nil // Bluetooth is on
}
func getDiskUsage() (string, error) {
	// Execute the df -h --total command
	out, err := exec.Command("df", "-h", "--total").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get disk usage: %v", err)
	}

	// Convert the output to a string
	output := string(out)

	// Split the output into lines
	lines := strings.Split(output, "\n")

	// Look for the line that contains 'total'
	for _, line := range lines {
		if strings.Contains(line, "total") {
			// Split the line into fields by spaces
			fields := strings.Fields(line)

			// Check if the output has at least 5 fields (Size, Used, Avail, Use%)
			if len(fields) >= 5 {
				totalSpace := fields[1]
				usedSpace := fields[2]
				availableSpace := fields[3]
				percentageUsed := fields[4]

				// Format the output without newlines
				result := fmt.Sprintf("Total Space: %s | Used Space: %s | Available Space: %s | Percentage of Use: %s",
					totalSpace, usedSpace, availableSpace, percentageUsed)

				// Remove any newlines if they exist (though they shouldn't in this format)
				result = strings.ReplaceAll(result, "\n", "")

				return result, nil
			}
		}
	}

	return "", fmt.Errorf("total disk usage information not found")
}

func getMemoryInfo() (string, string, error) {
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return "", "", err
	}
	totalMem := fmt.Sprintf("%.2f GB", float64(vmStat.Total)/(1024*1024*1024)) // covert gb
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
	// Execute the `ip addr` command to get network interface information
	out, err := exec.Command("ip", "addr").Output()
	if err != nil {
		return "", fmt.Errorf("Failed to retrieve network information")
	}

	networkInfo := string(out)

	// Search for the Ethernet interface (enp0s3) and link/ether
	if strings.Contains(networkInfo, "enp0s3") {
		lines := strings.Split(networkInfo, "\n")
		for i, line := range lines {
			// Find the line with "enp0s3"
			if strings.Contains(line, "enp0s3") {
				// Search for the "link/ether" information on the following lines
				for j := i; j < len(lines); j++ {
					if strings.Contains(lines[j], "link/ether") {
						// Return the exact string "wifi":"link/ether"
						return " enp0s3:link/ether", nil
					}
				}
			}
		}
	}

	return "", fmt.Errorf("No Ethernet info available")
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
			state = line
		} else if strings.Contains(line, "time to empty") {
			timeToEmpty = line
		} else if strings.Contains(line, "time to full") {
			timeToFull = line
		} else if strings.Contains(line, "percentage") {
			percentage = line
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

// Check if nmap is installed

// Function to perform vulnerability scan

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
func checkAndInstallNmap() error {
	_, err := exec.LookPath("nmap")
	if err != nil {
		fmt.Println("Nmap is not installed. Installing...")
		cmd := exec.Command("sudo", "apt", "install", "-y", "nmap")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install nmap: %w", err)
		}
		fmt.Println("Nmap has been installed.")
	}
	return nil
}
func getNmapScan() (string, error) {
	cmd := exec.Command("sudo", "nmap", "-sT", "-O", "localhost")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Error executing Nmap: %v", err)
	}

	// Use a scanner to read the Nmap output line by line
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	var nmapResults []string
	re := regexp.MustCompile(`^([0-9]+)/tcp\s+(\S+)\s+(\S+)`)

	for scanner.Scan() {
		line := scanner.Text()
		if matches := re.FindStringSubmatch(line); matches != nil {
			// Extract port, state, and service
			port := matches[1]
			state := matches[2]
			service := matches[3]
			nmapResults = append(nmapResults, fmt.Sprintf("Port: %s, State: %s, Service: %s", port, state, service))
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("Error reading Nmap output: %v", err)
	}

	if len(nmapResults) == 0 {
		return "No open ports detected", nil
	}
	return strings.Join(nmapResults, "; "), nil
}

func main() {
	// Check and install nmap if not installed
	if err := checkAndInstallNmap(); err != nil {
		log.Fatal(err)
	}
	// WebSocket connection setup
	wsURL := "ws://192.168.11.194:8080/SystemMonitoring/serverws" // Replace with your actual WebSocket server URL
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Fatal("Error connecting to WebSocket:", err)
	}
	defer conn.Close()

	// Infinite loop to update system information every 1 minute
	for {

		nmap, err := getNmapScan()
		fmt.Println(nmap)
		if err != nil {
			log.Fatal("Error getting host info:", err)
		}

		bluetoothuse, err := getblueusage()
		if err != nil {
			panic(err)
		}
		println("Bluetooth Status:", bluetoothuse)

		// Now, you can use the bluetoothuse variable
		//println("Bluetooth Status:", bluetoothuse)
		diskuse, err := getDiskUsage()
		if err != nil {
			log.Fatal("Error getting host info:", err)
		}

		typc, err := gethardwareModel()
		if err != nil {
			log.Fatal("Error getting host info:", err)
		}

		firewalstatus, err := getfirewall()
		if err != nil {
			log.Fatal("Error getting host info:", err)
		}
		vendor, err := getVendorType()
		if err != nil {
			log.Fatal("Error getting host info:", err)
		}

		typcc, err := getOSType()
		if err != nil {
			log.Fatal("Error getting host info:", err)
		}

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
		fmt.Println(totalMem)
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
			Hostname:       hostname,
			IP:             ipAddress,
			CPUModel:       cpuModel,
			TotalMemory:    totalMem,
			UsedMemory:     usedMem,
			Uptime:         uptime,
			WiFi:           wifi,
			Battery:        battery,
			SSHInfo:        ssh,
			Timestamp:      currentTime,
			OsName:         typcc,
			HardwareModel:  typc,
			HardwareVendor: vendor,
			Firewallstatus: firewalstatus,
			NmapScan:       nmap,
			Bluetoothuse:   bluetoothuse,
			DiskUsage:      diskuse,
		}

		// Wrap system info inside SystemInfoWrapper with the "system1_info" key
		wrappedInfo := SystemInfoWrapper{
			System1Info: sysInfo,
		}

		// Convert system information to JSON
		jsonData, err := json.Marshal(wrappedInfo)
		//fmt.Println(jsonData)
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
