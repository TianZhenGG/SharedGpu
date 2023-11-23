package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sharedgpu/proxy"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/flopp/go-findfont"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

// 定义全局变量
var (
	globalFolderPath string
	globalEditorVim  *widget.Entry
	globalLeftbottom *fyne.Container
	globalFilePath   string
)

// 连接到 Redis
// client := redis.NewClient(&redis.Options{
//     Addr:     "localhost:6379",
//     Password: "", // 如果没有密码，就留空
//     DB:       0,  // 使用默认的 DB
// })

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

func readFile(currentFilePath string, editorVim *widget.Entry, fileButton *LeftAlignedButton) {
	// 检查文件是否存在
	if _, err := os.Stat(currentFilePath); os.IsNotExist(err) {
		fyne.LogError("文件不存在", err)
		return
	}

	// 更新 globalFilePath
	globalFilePath = currentFilePath

	// 如果是文件，打开文件
	f, err := os.Open(currentFilePath)
	if err != nil {
		fyne.LogError("无法打开文件", err)
		return
	}
	defer f.Close()

	// 读取文件内容
	content, err := ioutil.ReadAll(f)
	if err != nil {
		fyne.LogError("无法读取文件", err)
		return
	}

	// 将文件内容显示在编辑器中
	editorVim.SetText(string(content))

	// 添加横杠来表示当前选中的文件
	fileButton.Text = "-" + fileButton.Text
	fileButton.Refresh()
}

func showFolderContents(folderPath string, editorVim *widget.Entry, leftbottom *fyne.Container) {

	// 更新全局变量
	// 更新全局变量
	globalFolderPath = folderPath
	globalEditorVim = editorVim
	globalLeftbottom = leftbottom
	// 获取文件或文件夹的信息
	info, err := os.Stat(folderPath)
	if err != nil {
		fyne.LogError("无法获取文件或文件夹信息", err)
		return
	}

	if !info.IsDir() {
		fmt.Println("folderPath:", folderPath) // 打印出 folderPath 的值
		fyne.LogError("路径不是一个文件夹", err)
		return
	}

	// 读取文件夹内容
	files, err := ioutil.ReadDir(folderPath)
	if err != nil {
		fyne.LogError("无法读取文件夹", err)
		return
	}

	// 清空当前的 leftbottom 容器
	leftbottom.Objects = nil

	// 为每个文件或文件夹创建一个按钮
	for _, f := range files {
		file := f // 创建一个新的变量来存储当前的文件
		fileButton := NewLeftAlignedButton(file.Name(), nil)

		// 捕获当前的文件路径
		currentFilePath := filepath.Join(folderPath, file.Name())

		// 设置点击事件处理函数
		fileButton.OnTapped = func() {
			if file.IsDir() {
				// 是文件夹，显示文件夹下的内容
				showFolderContents(currentFilePath, editorVim, leftbottom)
			} else {
				// 是文件，读取并显示文件内容
				readFile(currentFilePath, editorVim, fileButton)
			}
		}

		// 将按钮添加到 leftbottom 容器
		leftbottom.Add(fileButton)
	}

	// 刷新 leftbottom 容器
	leftbottom.Refresh()
}

func getSystemUsage() (cpuUsage, memoryUsage, diskUsage, networkUsage string, err error) {
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		return "", "", "", "", err
	}
	cpuUsage = fmt.Sprintf("%.2f%%", cpuPercent[0])

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return "", "", "", "", err
	}
	memoryUsage = fmt.Sprintf("%.2f%%", memInfo.UsedPercent)

	diskInfo, err := disk.Usage("/")
	if err != nil {
		return "", "", "", "", err
	}
	diskUsage = fmt.Sprintf("%.2f%%", diskInfo.UsedPercent)

	// 获取初始网络接口信息
	netIOs1, err := net.IOCounters(true)
	if err != nil {
		return "", "", "", "", err
	}

	// 等待一段时间
	time.Sleep(1 * time.Second)

	// 获取结束时的网络接口信息
	netIOs2, err := net.IOCounters(true)
	if err != nil {
		return "", "", "", "", err
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

	return cpuUsage, memoryUsage, diskUsage, networkUsage, nil
}

type LeftAlignedButton struct {
	widget.BaseWidget
	Text     string
	OnTapped func()
}

func NewLeftAlignedButton(text string, tapped func()) *LeftAlignedButton {
	b := &LeftAlignedButton{
		Text:     text,
		OnTapped: tapped,
	}
	b.ExtendBaseWidget(b)
	return b
}

func (b *LeftAlignedButton) CreateRenderer() fyne.WidgetRenderer {
	label := canvas.NewText(b.Text, theme.ForegroundColor())
	label.Alignment = fyne.TextAlignLeading
	return widget.NewSimpleRenderer(label)
}

func (b *LeftAlignedButton) Tapped(*fyne.PointEvent) {
	if b.OnTapped != nil {
		b.OnTapped()
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

			//定义一个"" 的 result
			result := ""
			if result == "asda" {
				dialog.ShowError(fmt.Errorf("failed"), myWindow)
			} else {
				dialog.ShowInformation("租用成功", "租用成功", myWindow)
			}

		}, myWindow)

		dialog.Show()
	})
	addMachineButton.Importance = widget.HighImportance

	// 创建租用机器和管理数据集的按钮，并设置颜色
	rentMachineButton := widget.NewButton("出租机器", func() {
		// 创建并显示一个圆形进度条
		progress := widget.NewProgressBarInfinite()
		progressDialog := dialog.NewCustom("正在出租机器...", "", progress, myWindow)
		progressDialog.Show()

		// 在新的 goroutine 中运行 proxy.StartSShServer()
		go func() {
			errChan := proxy.StartSShServer()
			err := <-errChan
			progressDialog.Hide()
			if err != nil {
				dialog.ShowError(err, myWindow)
			} else {
				//打印出租成功2s后关闭

				// 创建一个新的窗口来显示仪表盘
				dashboardWindow := fyne.CurrentApp().NewWindow("仪表盘")
				cpuLabel := widget.NewLabel("CPU: ")
				memoryLabel := widget.NewLabel("内存: ")
				diskLabel := widget.NewLabel("磁盘: ")
				networkLabel := widget.NewLabel("网络: ")
				gpuMemoryLabel := widget.NewLabel("GPU 内存: ")
				dashboardWindow.SetContent(container.NewVBox(cpuLabel, memoryLabel, diskLabel, networkLabel, gpuMemoryLabel))
				dashboardWindow.Show()

				// 在新的 goroutine 中定期更新仪表盘
				go func() {
					for {
						// 这里需要你自己的函数来获取 CPU、内存、磁盘、网络和 GPU 内存的占用情况
						cpu, memory, disk, network, gpuMemory := getSystemUsage()
						cpuLabel.SetText(fmt.Sprintf("CPU: %s", cpu))
						memoryLabel.SetText(fmt.Sprintf("内存: %s", memory))
						diskLabel.SetText(fmt.Sprintf("磁盘: %s", disk))
						networkLabel.SetText(fmt.Sprintf("网络: %s", network))
						gpuMemoryLabel.SetText(fmt.Sprintf("GPU 内存: %s", gpuMemory))
						time.Sleep(time.Second)
					}
				}()
			}
		}()
	})
	rentMachineButton.Importance = widget.LowImportance
	manageDatasetButton := widget.NewButton("数据集", func() {
		// 在这里管理数据集的代码
	})
	manageDatasetButton.Importance = widget.HighImportance
	// 创建一个新的容器
	leftSplit := container.NewVBox()

	// 创建一个新的水平容器
	buttonContainer := container.NewHBox()

	// 创建 "新建" 按钮
	newButton := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		// 获取当前窗口
		win := fyne.CurrentApp().Driver().AllWindows()[0]

		// 创建一个输入字段用于输入文件或文件夹的名称
		nameEntry := widget.NewEntry()

		// 创建一个对话框，包含 "创建文件" 和 "创建文件夹" 两个按钮和一个输入字段
		dialog.ShowCustomConfirm("新建", "创建文件", "创建文件夹", fyne.NewContainerWithLayout(layout.NewVBoxLayout(), nameEntry), func(createFile bool) {
			// 获取输入的名称
			name := nameEntry.Text

			if createFile {
				// 在全局变量路径下创建文件
				newFilePath := filepath.Join(globalFolderPath, name)
				_, err := os.Create(newFilePath)
				if err != nil {
					fyne.LogError("无法创建文件", err)
					return
				}
			} else {
				// 在全局变量路径下创建文件夹
				newFolderPath := filepath.Join(globalFolderPath, name)
				err := os.Mkdir(newFolderPath, 0755)
				if err != nil {
					fyne.LogError("无法创建文件夹", err)
					return
				}
			}

			// 刷新文件列表
			showFolderContents(globalFolderPath, globalEditorVim, globalLeftbottom)
		}, win)
	})

	// 创建 "保存" 按钮
	saveButton := widget.NewButtonWithIcon("", theme.DocumentSaveIcon(), func() {
		// "保存" 按钮的点击事件
		//如果当前没有打开任何文件，则不做任何操作
		if globalFilePath == "" {
			return
		}
		// 获取编辑器中的文本
		content := globalEditorVim.Text

		// 保存到文件
		err := ioutil.WriteFile(globalFilePath, []byte(content), 0644)
		if err != nil {
			// 如果保存失败，显示错误信息
			fyne.LogError("无法保存文件", err)
			return
		}

		// 如果保存成功，显示成功信息
		dialog.ShowInformation("保存", "文件已成功保存", myWindow)
	})

	// 创建 "回退" 按钮
	backButton := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		// 当点击按钮的时候，获取当前文件的路径，如果是第一级目录或者没有文件路径，则不做任何操作，否则返回上一级目录，刷新左侧的文件列表
		// 如果当前目录为空，则不执行回退操作
		if globalFolderPath == "" {
			return
		}

		// 获取上级目录
		parentPath := filepath.Dir(globalFolderPath)

		// 如果当前目录已经是顶级目录，则不执行回退操作
		if parentPath == globalFolderPath {
			return
		}

		// 显示上级目录的内容
		showFolderContents(parentPath, globalEditorVim, globalLeftbottom)

		// 更新当前路径
		globalFolderPath = parentPath
	})
	buttonContainer.Add(newButton)
	buttonContainer.Add(saveButton)
	buttonContainer.Add(backButton)

	leftSplit.Add(buttonContainer)
	// leftSplit里面新建个容器叫做leftbottom
	leftbottom := container.NewVBox()

	// 创建底部的输出面板，宽度是左侧菜单栏的90%
	output := widget.NewMultiLineEntry()
	output.SetPlaceHolder("输出面板")
	output.Wrapping = fyne.TextWrapWord
	output.Disable()

	// 创建中间的编辑器
	editorVim := widget.NewMultiLineEntry()
	editorVim.TextStyle = fyne.TextStyle{Monospace: true, Bold: true, Italic: true}
	// 创建新的按钮
	rightButton := widget.NewButton("执行", func() {
		// 按钮的点击事件处理函数
	})

	newBottom := container.NewHSplit(
		output,
		rightButton,
	)

	newBottom.Offset = 0.9 // 设置 output 和 rightButton 的大小比例为 9:1

	// 创建一个支持滚动的容器，然后将 editorVim 添加到这个容器中
	scrollableEditorVim := container.NewHScroll(editorVim)
	editorVimSplit := container.NewVSplit(
		scrollableEditorVim,
		newBottom,
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

					// 更新全局变量为选择的路径
					globalFolderPath = uri.Path()
					globalEditorVim = editorVim
					globalLeftbottom = leftbottom
					// 清空 leftSplit
					leftbottom.Objects = nil

					// 将文件名显示在leftSplit中,并且为每个文件名添加一个点击事件，点击不同的文件名时，显示不同的文件内容
					for _, file := range files {
						// 捕获当前的文件路径
						currentFilePath := file.Path()

						// 获取文件的信息
						info, err := os.Stat(currentFilePath)
						if err != nil {
							fyne.LogError("无法获取文件信息", err)
							return
						}

						// 捕获当前的文件是否是一个目录
						isDir := info.IsDir()

						if isDir {
							// 如果是目录，则创建一个新的按钮
							_ = NewLeftAlignedButton(file.Name(), func() {

							})
						}

						// 在这里定义和初始化 fileButton
						fileButton := NewLeftAlignedButton(file.Name(), nil)

						fileButton.OnTapped = func() {
							// 在这里，currentFilePath 已经被捕获，所以我们可以直接使用它

							if isDir {
								// 显示新文件夹的内容
								showFolderContents(currentFilePath, editorVim, leftbottom)

							} else {
								readFile(currentFilePath, editorVim, fileButton)
							}
						}
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
	menuSplit.Offset = 0.9 // 调整宽度，使左侧菜单更窄

	// 创建一个新的 Split 来包含 menuSplit 和 editorVimSplit
	mainSplit := container.NewHSplit(menuSplit, editorVimSplit)
	mainSplit.Offset = 0.2 // 调整位置

	myWindow.SetContent(mainSplit)
	myWindow.ShowAndRun()
}
