package utils

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sharedgpu/bdfs"
	"strings"
	"time"

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

func ExecCommand(execType string, bottomInput *widget.Entry, bottomPart *widget.Label, globalProject string, uuidStr string, rdb *redis.Client) {
	var inputText string
	ctx := rdb.Context()
	// 根据execType判断是执行本地还是执行远程
	if execType == "local" {
		inputText = bottomInput.Text
	} else {
		// 从redis中获取命令
		cmd, err := rdb.HGet(ctx, uuidStr, "cmd").Result()
		if err != nil {
			//打入redis log
			err = rdb.HSet(ctx, uuidStr, "log", err.Error()).Err()
			if err != nil {
				fmt.Println("failed to set log:", err)
			}
		}
		fmt.Println("get cmd", cmd)
		inputText = cmd
	}
	// 如果输入框为空，则不执行任何操作
	if inputText == "" {
		bottomPart.SetText("请输入命令。。。")
		return
	}
	//清空bottomInput
	bottomInput.SetText("")

	//解析输入的文本，如果是python或者是python3，改成miniconda/python.exe
	// 解析输入的文本，如果是python或者是python3，改成miniconda/python.exe
	inputText = strings.Replace(inputText, "python ", "miniconda/python.exe ", -1)
	inputText = strings.Replace(inputText, "python3 ", "miniconda/python.exe ", -1)
	inputText = strings.Replace(inputText, "pip ", "miniconda/python.exe -m pip ", -1)
	inputText = strings.Replace(inputText, "pip3 ", "miniconda/python.exe -m pip ", -1)

	//如果本地没有miniconda，则下载miniconda
	if _, err := os.Stat("miniconda"); os.IsNotExist(err) {
		if execType == "local" {
			bottomPart.SetText("正在配置环境，请稍等。。。")
		} else {
			//打入redis log
			err = rdb.HSet(ctx, uuidStr, "log", "正在配置环境，请稍等。。。").Err()
			if err != nil {
				fmt.Println("failed to set log:", err)
			}
		}
		// 执行本地机器的代码
		err = bdfs.Download("miniconda", "miniconda.zip", "./")
		if err != nil {
			fmt.Println("failed to download file:", err)
		}
		//project目录下有没有.BaiduPCS-Go-downloading结尾的文件，如果有则等待，如果没有则解压文件
		for {

			files, err := ioutil.ReadDir(globalProject)
			if err != nil {
				fmt.Println("failed to read dir:", err)
			}
			downloading := false
			for _, f := range files {
				if strings.HasSuffix(f.Name(), ".BaiduPCS-Go-downloading") {
					time.Sleep(time.Second * 2)
					downloading = true
					break
				}
			}
			if downloading {
				continue
			}

			// 解压文件
			fmt.Println("解压文件", globalProject)
			err = bdfs.Unzip("miniconda.zip", globalProject)
			if err != nil {
				fmt.Println("failed to unzip file:", err)
			}
			// 删除压缩包
			err = os.Remove("miniconda.zip")
			if err != nil {
				fmt.Println("failed to remove file:", err)
			}
			break
		}

	}

	// 切分 bottomInput.Text
	args := strings.Fields(inputText)
	fmt.Println("args", args)
	cmd := exec.Command(args[0], args[1:]...)
	// 获取命令的输出
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		if execType == "local" {
			bottomPart.SetText(err.Error())
		} else {
			//打入redis log
			err = rdb.HSet(ctx, uuidStr, "log", err.Error()).Err()
			if err != nil {
				fmt.Println("failed to set log:", err)
			}
		}
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		if execType == "local" {
			bottomPart.SetText(err.Error())
		} else {
			//打入redis log
			err = rdb.HSet(ctx, uuidStr, "log", err.Error()).Err()
			if err != nil {
				fmt.Println("failed to set log:", err)
			}
		}
	}

	// 创建一个新的 scanner 来读取命令的输出
	outScanner := bufio.NewScanner(stdout)
	errScanner := bufio.NewScanner(stderr)

	// 使用一个 goroutine 来读取命令的输出
	go func() {
		for outScanner.Scan() {
			if execType == "local" {
				bottomPart.SetText(bottomPart.Text + outScanner.Text() + "\n")
			} else {
				//打入redis log
				err = rdb.HSet(ctx, uuidStr, "log", outScanner.Text()).Err()
				if err != nil {
					fmt.Println("failed to set log:", err)
				}
			}
		}
	}()
	go func() {
		for errScanner.Scan() {
			if execType == "local" {
				bottomPart.SetText(bottomPart.Text + errScanner.Text() + "\n")
			} else {
				//打入redis log
				err = rdb.HSet(ctx, uuidStr, "log", errScanner.Text()).Err()
				if err != nil {
					fmt.Println("failed to set log:", err)
				}
			}
		}
	}()

	// 运行命令
	err = cmd.Start()
	if err != nil {
		if execType == "local" {
			bottomPart.SetText(err.Error())
		} else {
			//打入redis log
			err = rdb.HSet(ctx, uuidStr, "log", err.Error()).Err()
			if err != nil {
				fmt.Println("failed to set log:", err)
			}
		}
	}
	err = cmd.Wait()
	if err != nil {
		if execType == "local" {
			bottomPart.SetText(err.Error())
		} else {
			//打入redis log
			err = rdb.HSet(ctx, uuidStr, "log", err.Error()).Err()
			if err != nil {
				fmt.Println("failed to set log:", err)
			}
		}
	}

}
