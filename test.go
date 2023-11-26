package main

import (
	"fmt"
	"os/exec"
	"strings"
)

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

func main() {
	gpuMemoryUsage, err := GetGPUMemoryUsage()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for gpuName, memoryUsed := range gpuMemoryUsage {
		fmt.Printf("GPU %s memory used: %s\n", gpuName, memoryUsed)
	}
}
