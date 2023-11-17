package main

import (
	"os"
	"strings"
	"fmt"
	"time"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/flopp/go-findfont"
)

func init() {
	//设置中文字体:解决中文乱码问题
	fontPaths := findfont.List()
	for _, path := range fontPaths {
		if strings.Contains(path, "msyh.ttf") || strings.Contains(path, "simhei.ttf") || strings.Contains(path, "simsun.ttc") || strings.Contains(path, "simkai.ttf") {
			os.Setenv("FYNE_FONT", path)
			break
		}
	}
}

func main() {
	myApp := app.NewWithID("myApp")
	myApp.Settings().SetTheme(theme.DarkTheme())

	myWindow := myApp.NewWindow("Client")
	myWindow.Resize(fyne.NewSize(800, 600))

	// 创建添加和删除机器的按钮，并设置颜色
	addMachineButton := widget.NewButton("新加机器", func() {
		// 在这里添加机器的代码
		// 创建对话框的表单
		form := &widget.Form{}

		// 添加显卡配置的下拉菜单
		gpuOptions := []string{"GPU1", "GPU2", "GPU3"}
		gpuSelect := widget.NewSelect(gpuOptions, func(value string) {
			// 在这里处理用户选择的显卡配置
		})
		form.Append("显卡配置", gpuSelect)
	
		// 添加内存配置的下拉菜单
		memoryOptions := []string{"8GB", "16GB", "32GB"}
		memorySelect := widget.NewSelect(memoryOptions, func(value string) {
			// 在这里处理用户选择的内存配置
		})
		form.Append("内存配置", memorySelect)
	
		// 添加 CPU 配置的下拉菜单
		cpuOptions := []string{"CPU1", "CPU2", "CPU3"}
		cpuSelect := widget.NewSelect(cpuOptions, func(value string) {
			// 在这里处理用户选择的 CPU 配置
		})
		form.Append("CPU 配置", cpuSelect)
	
		// 创建对话框
		dialog := dialog.NewCustomConfirm("新加机器", "确定", "取消", form, func(ok bool) {
			if !ok {
				return
			}
	
			// 创建一个进度条对话框
			progressDialog := dialog.NewProgress("正在匹配机器", "请稍等...", myWindow)

			// 启动一个 goroutine 来更新进度条
			go func() {
				for i := 0.0; i <= 1.0; i += 0.1 {
					time.Sleep(500 * time.Millisecond) // 模拟匹配机器的过程
					progressDialog.SetValue(i)         // 更新进度条
				}
				progressDialog.Hide() // 隐藏进度条对话框
				
				// 判断是否成功
				success := true
				// 如果成功将设备信息显示在主界面上
				if success {
					// 在这里显示设备信息
					fmt.Println("匹配成功")
				} else {
					// 如果失败则显示失败信息
					dialog.ShowError(fmt.Errorf("匹配失败"), myWindow)
				}

			}()
		
			// 显示进度条对话框
			progressDialog.Show()
		}, myWindow)
	
		dialog.Show()
	})
	addMachineButton.Importance = widget.HighImportance

	deleteMachineButton := widget.NewButton("删除机器", func() {
		// 在这里删除机器的代码
	})
	deleteMachineButton.Importance = widget.MediumImportance

	// 创建租用机器和管理数据集的按钮，并设置颜色
	rentMachineButton := widget.NewButton("租用机器", func() {
		// 在这里租用机器的代码
	})
	rentMachineButton.Importance = widget.LowImportance

	manageDatasetButton := widget.NewButton("数据集", func() {
		// 在这里管理数据集的代码
	})
	manageDatasetButton.Importance = widget.HighImportance

	// 创建左侧菜单栏并添加按钮
	leftMenu := container.NewHBox(
		container.NewVBox(
			addMachineButton,
			deleteMachineButton,
			rentMachineButton,
			manageDatasetButton,
		),
		widget.NewSeparator(),
		container.NewMax(),
	)

	// 创建主容器并添加左侧菜单栏
	content := container.NewBorder(leftMenu, nil, nil, nil)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}
