package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"os/exec"
	"strings"
	"time"
)

func GenerateUUID(machineModel string) uuid.UUID {
	hasher := md5.New()
	hasher.Write([]byte(machineModel))
	md5String := hex.EncodeToString(hasher.Sum(nil))

	namespace := uuid.Must(uuid.Parse(md5String))
	return uuid.NewSHA1(namespace, []byte(machineModel))
}

func GetSystemUsage() (cpuUsage, memoryUsage, diskUsage, networkUsage string, gpuMemUsage string, err error) {
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		return "", "", "", "", "", err
	}
	cpuUsage = fmt.Sprintf("%.2f%%", cpuPercent[0])

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return "", "", "", "", "", err
	}
	memoryUsage = fmt.Sprintf("%.2f%%", memInfo.UsedPercent)

	diskInfo, err := disk.Usage("/")
	if err != nil {
		return "", "", "", "", "", err
	}
	diskUsage = fmt.Sprintf("%.2f%%", diskInfo.UsedPercent)

	// 获取初始网络接口信息
	netIOs1, err := net.IOCounters(true)
	if err != nil {
		return "", "", "", "", "", err
	}

	// 等待一段时间
	time.Sleep(1 * time.Second)

	// 获取结束时的网络接口信息
	netIOs2, err := net.IOCounters(true)
	if err != nil {
		return "", "", "", "", "", err
	}

	// 计算所有接口的总发送和接收字节数的差值，然后转换为兆字节/秒
	var totalMBsSent, totalMBsRecv float64
	for i, netIO := range netIOs1 {
		totalMBsSent += float64(netIOs2[i].BytesSent-netIO.BytesSent) / 1048576.0
		totalMBsRecv += float64(netIOs2[i].BytesRecv-netIO.BytesRecv) / 1048576.0
	}

	// 格式化网络使用情况
	networkUsage = fmt.Sprintf("Upload: %.2f MB/s, Download: %.2f MB/s", totalMBsSent, totalMBsRecv)

	// 获取 GPU 内存的占用情况可能需要特定的库或 API，这取决于你的环境和需求
	gpuMemoryUsage, err := GetGPUMemoryUsage()
	if err != nil {
		return "", "", "", "", "", err
	}
	var gpuMemoryUsageStr string
	for gpu, memory := range gpuMemoryUsage {
		gpuMemoryUsageStr += fmt.Sprintf("%s: %s\n", gpu, memory)
	}
	return cpuUsage, memoryUsage, diskUsage, networkUsage, gpuMemoryUsageStr, nil
}

// 获取显卡信息
func GetGPUMemoryUsage() (map[string]string, error) {
	cmd := exec.Command("nvidia-smi", "--query-gpu=name,memory.used", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run nvidia-smi: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	gpuMemoryUsage := make(map[string]string)
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ", ")
		gpuName := parts[0]
		memoryUsed := parts[1]
		gpuMemoryUsage[gpuName] = memoryUsed
	}

	return gpuMemoryUsage, nil
}

// 获取本机的cpu，内存，显卡型号和数量信息
func GetSystemInfo() (cpuInfo, memoryInfo, gpuInfo string, err error) {
	// 获取 CPU 信息
	cpuStats, err := cpu.Info()
	if err != nil {
		return "", "", "", err
	}
	cpuInfo = cpuStats[0].ModelName

	// 获取内存信息
	memStats, err := mem.VirtualMemory()
	if err != nil {
		return "", "", "", err
	}
	memoryInfo = fmt.Sprintf("Total: %.2f GB", float64(memStats.Total)/1024/1024/1024)

	// 获取显卡信息
	gpuStats, err := GetGPUMemoryUsage()
	if err != nil {
		return "", "", "", err
	}
	for gpu, memory := range gpuStats {
		gpuInfo += fmt.Sprintf("%s: %s\n", gpu, memory)
	}
	return cpuInfo, memoryInfo, gpuInfo, nil
}
