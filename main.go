package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
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
	addMachineButton := widget.NewButton("租用机器", func() {
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
		dialog := dialog.NewCustomConfirm("租用机器", "确定", "取消", form, func(ok bool) {
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
	rentMachineButton := widget.NewButton("出租机器", func() {
		// 在这里出租机器的代码

		// 创建并显示一个圆形进度条
		progress := widget.NewProgressBarInfinite()
		progressDialog := dialog.NewCustom("正在出租机器...", "", progress, myWindow)
		progressDialog.Show()
	})
	rentMachineButton.Importance = widget.LowImportance
	manageDatasetButton := widget.NewButton("数据集", func() {
		// 在这里管理数据集的代码
	})
	manageDatasetButton.Importance = widget.HighImportance
	// 创建一个新的容器
	leftSplit := container.NewVBox()

	// 定义 currentFilePath 变量
	var currentFilePath string

	// 定义 textEditor 变量
	var textEditor *widget.Entry

	// 创建一个新的水平容器
	buttonContainer := container.NewHBox()

	// 创建 "新建" 按钮
	newButton := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		// "新建" 按钮的点击事件
		// 创建新的文件
		f, err := os.Create("newfile.txt")
		if err != nil {
			fyne.LogError("无法创建新的文件", err)
			return
		}
		defer f.Close()

		// 更新当前打开的文件的路径
		currentFilePath = "newfile.txt"

		// 清空 textEditor 的内容
		textEditor.SetText("")
	})

	// 创建 "保存" 按钮
	saveButton := widget.NewButtonWithIcon("", theme.DocumentSaveIcon(), func() {
		// "保存" 按钮的点击事件
		// 打开文件
		f, err := os.OpenFile(currentFilePath, os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			fyne.LogError("无法打开文件", err)
			return
		}
		defer f.Close()

		// 将 textEditor 中的内容写入到文件
		_, err = f.WriteString(textEditor.Text)
		if err != nil {
			fyne.LogError("无法写入文件", err)
			return
		}
	})

	// 创建 "删除" 按钮
	deleteButton := widget.NewButtonWithIcon("", theme.ContentRemoveIcon(), func() {
		// "删除" 按钮的点击事件
		// 删除当前打开的文件
		err := os.Remove(currentFilePath)
		if err != nil {
			fyne.LogError("无法删除文件", err)
			return
		}

		// 清空 textEditor 的内容
		textEditor.SetText("")
	})
	// 在 buttonContainer 中添加 "新建"、"保存" 和 "删除" 三个按钮
	buttonContainer.Add(newButton)
	buttonContainer.Add(saveButton)
	buttonContainer.Add(deleteButton)

	leftSplit.Add(buttonContainer)
	// leftSplit里面新建个容器叫做leftbottom
	leftbottom := container.NewVBox()

	// 创建底部的输出面板
	output := widget.NewMultiLineEntry()
	output.Disable()

	// 创建中间的编辑器
	editorVim := widget.NewMultiLineEntry()

	// 创建右侧的按钮
	rightButton := widget.NewButton("普通按钮", nil)
	// 创建主容器并添加左侧菜单栏、编辑器、输出面板和属性面板
	editorVimSplit := container.NewVSplit(
		editorVim,
		container.NewHSplit(
			output,
			rightButton,
		),
	)
	importButton := widget.NewButton("导入代码", func() {
		// 创建一个新的列表来显示文件名
		fileList := widget.NewList(
			func() int { return 0 },                                 // 初始时列表为空
			func() fyne.CanvasObject { return widget.NewLabel("") }, // 创建新的标签来显示文件名
			func(id widget.ListItemID, item fyne.CanvasObject) {},   // 初始时列表为空，所以这个函数不做任何事
		)

		var customDialog *widget.PopUp

		localImportButton := widget.NewButton("本地导入", func() {
			dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
				if err == nil && uri != nil {
					// 获取文件夹下的所有文件
					files, err := uri.List()
					if err != nil {
						fyne.LogError("无法获取文件列表", err)
						return
					}

					// 清空 leftSplit
					leftbottom.Objects = nil

					// 将文件名显示在leftSplit中,并且为每个文件名添加一个点击事件，点击不同的文件名时，显示不同的文件内容
					for _, file := range files {
						// 捕获当前的文件路径
						currentFilePath := file.Path()

						fileButton := widget.NewButton(file.Name(), func() {
							// 打开文件
							f, err := os.Open(currentFilePath)
							if err != nil {
								fyne.LogError("无法打开文件", err)
								return
							}
							defer f.Close()

							// 读取文件内容
							fileContent, err := ioutil.ReadAll(f)
							if err != nil {
								fyne.LogError("无法读取文件", err)
								return
							}

							// 将文件内容显示在 editorVim 中
							editorVim.SetText(string(fileContent))
						})

						leftbottom.Add(fileButton)
					}
					leftSplit.Add(leftbottom)
					// 导入完成后，隐藏对话框
					customDialog.Hide()
				} else {
					customDialog.Hide()
				}
			}, myWindow)
		})

		githubImportButton := widget.NewButton("GitHub 导入", func() {
			// 在这里添加导入 GitHub 代码的代码
		})

		// 创建一个自定义的对话框并添加两个新的按钮和列表
		customDialog = widget.NewModalPopUp(container.NewVBox(
			container.NewHBox(localImportButton, githubImportButton),
			fileList,
		), myWindow.Canvas())
		customDialog.Show()
	})

	// 在左侧菜单栏添加新的按钮
	leftMenu := container.NewHBox(
		container.NewVBox(
			addMachineButton,
			deleteMachineButton,
			rentMachineButton,
			manageDatasetButton,
			importButton, // 新添加的按钮
		),
		widget.NewSeparator(),
		container.NewMax(),
	)

	// 调整中间编辑器的位置
	editorVimSplit.Offset = 0.9

	// 创建一个新的 Split 来包含 leftMenu 和 leftSplit
	menuSplit := container.NewHSplit(leftMenu, leftSplit)
	menuSplit.Offset = 0.1 // 调整宽度

	// 创建一个新的 Split 来包含 menuSplit 和 editorVimSplit
	mainSplit := container.NewHSplit(menuSplit, editorVimSplit)
	mainSplit.Offset = 0.2 // 调整位置

	myWindow.SetContent(mainSplit)
	myWindow.ShowAndRun()
}
