package utils

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
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

func CreateNewWindow(rdb *redis.Client, uuidStr string) {
	// 创建一个新的窗口来显示仪表盘
	dashboardWindow := fyne.CurrentApp().NewWindow("仪表盘")
	cpuLabel := widget.NewLabel("CPU: ")
	memoryLabel := widget.NewLabel("内存: ")
	diskLabel := widget.NewLabel("磁盘: ")
	networkLabel := widget.NewLabel("网络: ")
	gpuMemoryLabel := widget.NewLabel("GPU 内存: ")

	// 创建一个可以被取消的 context
	ctx, cancel := context.WithCancel(context.Background())

	// 在新的 goroutine 中定期更新仪表盘
	go func() {
		for {
			select {
			case <-ctx.Done():
				// 如果 context 被取消，停止更新仪表盘
				return
			default:
				// 这里需要你自己的函数来获取 CPU、内存、磁盘、网络和 GPU 内存的占用情况
				cpu, memory, disk, network, gpuMemory, err := GetSystemUsage()
				if err != nil {
					fmt.Println(fmt.Errorf("failed to get system usage: %w", err))
				}
				// 更新标签的文本
				cpuLabel.SetText(fmt.Sprintf("CPU: %s", cpu))
				memoryLabel.SetText(fmt.Sprintf("内存: %s", memory))
				diskLabel.SetText(fmt.Sprintf("磁盘: %s", disk))
				networkLabel.SetText(fmt.Sprintf("网络: %s", network))
				gpuMemoryLabel.SetText(fmt.Sprintf("GPU 内存: %s", gpuMemory))

				// 等待一段时间再更新
				time.Sleep(time.Second * 1)
			}
		}
	}()

	// 创建一个新的按钮
	unshareMachineButton := widget.NewButton("取消共享", func() {
		// 删除 Redis 中的 uuid
		err := rdb.Del(ctx, uuidStr).Err()
		if err != nil {
			fmt.Println(fmt.Errorf("failed to delete uuid from redis: %w", err))
		}
		println("Machine is no longer shared.")

		// 关闭窗口
		dashboardWindow.Close()

		// 停止更新仪表盘
		cancel()

	})

	// 将 unshareMachineButton 添加到窗口的内容中
	dashboardWindow.SetContent(container.NewVBox(cpuLabel, memoryLabel, diskLabel, networkLabel, gpuMemoryLabel, unshareMachineButton))
	dashboardWindow.Show()
}
