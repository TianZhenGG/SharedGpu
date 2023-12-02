package main

import (
	"context"
	"fmt"
	"image/color"
	"io/ioutil"
	"os"
	"path/filepath"
	"sharedgpu/bdfs"
	"sharedgpu/db"
	"sharedgpu/proxy"
	"sharedgpu/utils"
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
	"golang.org/x/crypto/ssh"
)

// 定义全局变量
var (
	globalFolderPath string
	globalProject    string
	globalEditorVim  *widget.Entry
	globalLeftbottom *fyne.Container
	globalFilePath   string

	// 定义一个全局的 leftline 变量
	leftline = widget.NewMultiLineEntry()

	uuidStr string
)

var importPath string

// 定义 displayArea 为全局变量
var displayArea *widget.Entry

func init() {
	machineModel := "MachineModel123"
	uuidStr = utils.GenerateUUID(machineModel).String()
	//设置中文字体:解决中文乱码问题
	fontPaths := findfont.List()
	for _, path := range fontPaths {
		if strings.Contains(path, "msyh.ttf") || strings.Contains(path, "simhei.ttf") || strings.Contains(path, "simsun.ttc") || strings.Contains(path, "simkai.ttf") {
			os.Setenv("FYNE_FONT", path)
			break
		}
	}
}

func readFile(currentFilePath string, editorVim *widget.Entry) {
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

	// 计算文件的行数
	lines := strings.Split(string(content), "\n")

	// 在 leftline 中显示行号
	lineNumbers := ""
	for i := 1; i <= len(lines); i++ {
		lineNumbers += fmt.Sprintf("%d\n", i)
	}
	leftline.SetText(lineNumbers)

	// 将文件内容显示在编辑器中
	//用canvas.NewText()来设置字体颜色
	contentObj := canvas.NewText(string(content), color.RGBA{255, 255, 255, 255})
	//设置字体大小
	contentObj.TextSize = 20
	contentStr := contentObj.Text
	editorVim.SetText(string(contentStr))
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

	// // 初始化随机数生成器
	// rand.Seed(time.Now().UnixNano())

	// // 打乱 files 切片
	// rand.Shuffle(len(files), func(i, j int) {
	// 	files[i], files[j] = files[j], files[i]
	// })

	// 清空当前的 leftbottom 容器
	leftbottom.Objects = nil
	//files 顺序随机打乱
	// 为每个文件或文件夹创建一个对象，用theme接口创建
	for _, f := range files {
		file := f // 创建一个新的变量来存储当前的文件

		// 捕获当前的文件路径
		currentFilePath := filepath.Join(folderPath, file.Name())

		if file.IsDir() {
			folderButton := widget.NewButtonWithIcon(file.Name(), theme.FolderIcon(), func() {
				// 是文件夹，显示文件夹下的内容
				showFolderContents(currentFilePath, editorVim, leftbottom)
			})
			leftbottom.Add(folderButton)
		} else {
			fileButton := widget.NewButtonWithIcon(file.Name(), theme.FileIcon(), func() {
				// 是文件，显示文件内容
				readFile(currentFilePath, editorVim)
			})
			leftbottom.Add(fileButton)
		}
	}

	// 刷新 leftbottom 容器
	leftbottom.Refresh()
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
	myWindow.Resize(fyne.NewSize(1024, 768))

	// 创建一个 context.Context 对象
	ctx := context.Background()

	rdb, err := db.InitRedis()
	if err != nil {
		fmt.Println("redis init failed")
	}
	// 创建添加和删除机器的按钮，并设置颜色
	addMachineButton := widget.NewButton("租用机器", func() {
		// 在这里添加机器的代码
		// 创建对话框的表单
		form := &widget.Form{}

		// 添加显卡配置的下拉菜单
		gpuOptions := []string{"敬请期待", "RTX 3080 Ti", "敬请期待"}
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

			// gpuselect值来模糊匹配gpu型号
			fmt.Println("gpuSelect:", gpuSelect.Selected)
			// 查询redis下所有value记录并打印出来
			rAll, err := db.GetAllValues()
			if err != nil {
				fmt.Println("failed to get all values:", err)
			}
			fmt.Println("rAll:", rAll)
			var result string

			for _, v := range rAll {
				switch value := v.(type) {
				case string:
					if strings.Contains(value, gpuSelect.Selected) {
						result = value
						break
					}
				case map[string]string:
					// 处理哈希类型的值
					for _, hashValue := range value {
						if strings.Contains(hashValue, gpuSelect.Selected) {
							result = hashValue
							break
						}
					}
				default:
					fmt.Printf("unsupported value type: %T\n", v)
				}
			}
			// 如果result不是空，说明匹配到了机器
			if result != "" {
				fmt.Println("匹配到机器:", result)
				dialog.ShowInformation("租用成功", "租用成功", myWindow)

				// 创建 SSH 客户端配置
				config := &ssh.ClientConfig{
					User: "tian",
					Auth: []ssh.AuthMethod{
						ssh.Password("tian"),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				}

				// 在 goroutine 外部定义一个变量来存储初次连接成功的时间
				var startTime time.Time

				go func() {
					for range time.Tick(time.Second) {
						client, err := ssh.Dial("tcp", "127.0.0.1:3333", config)
						if err != nil {
							displayArea.SetText("连接失败: " + err.Error())
						} else {
							// 如果是初次连接成功，获取当前时间
							if startTime.IsZero() {
								startTime = time.Now()
							}

							// 计算时间差
							duration := time.Since(startTime)

							// 计算天、小时、分钟和秒
							days := int(duration.Hours()) / 24
							hours := int(duration.Hours()) % 24
							minutes := int(duration.Minutes()) % 60
							seconds := int(duration.Seconds()) % 60

							// 显示连接成功和时间差还有gpu相关信息
							displayArea.SetText(fmt.Sprintf("连接成功，已连接：%d天%d小时%d分钟%d秒，\nGPU:%s", days, hours, minutes, seconds, result))

							client.Close()
						}
					}
				}()

			} else {
				fmt.Println("没有匹配到机器")
				dialog.ShowInformation("暂无资源", "暂无资源", myWindow)
			}

		}, myWindow)

		dialog.Show()
	})
	addMachineButton.Importance = widget.HighImportance

	// 创建租用机器和管理数据集的按钮，并设置颜色
	rentMachineButton := widget.NewButton("出租机器", func() {

		// 根据 uuid 查询是否存在
		exists, err := rdb.Exists(ctx, uuidStr).Result()
		if err != nil {
			panic(err)
		}

		if exists == 1 {
			// 如果存在，直接返回信息，机器已共享
			println("Machine is already shared.")
		} else {

			// 获取机器的 CPU、内存、显卡型号和个数
			cpuInfo, memoryInfo, gpuInfo, nil := utils.GetSystemInfo()
			if err != nil {
				panic(err)
			}

			fmt.Println("cpuInfo:", cpuInfo, "memoryInfo:", memoryInfo, "gpuInfo:", gpuInfo)

			// 将 uuid 和机器的 CPU、内存、显卡型号存入 redis 想变成json形式
			err = rdb.HSet(ctx, uuidStr, "cpu", cpuInfo, "memory", memoryInfo, "gpu", gpuInfo).Err()
			if err != nil {
				fmt.Println("failed to set info to redis :", err)
			}

		}

		// 在新的 goroutine 中运行 proxy.StartSShServer()
		go func() {
			errChan := proxy.StartSShServer()
			_ = <-errChan
			utils.CreateNewWindow(rdb, uuidStr)
		}()
	})

	rentMachineButton.Importance = widget.LowImportance
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
		// 如果当前目录是导入代码的目录，或者为空，则不执行回退操作
		if globalFolderPath == "" || globalFolderPath == importPath {
			return
		}

		// 获取上级目录
		parentPath := filepath.Dir(globalFolderPath)

		// 如果当前目录已经是顶级目录，则不执行回退操作
		if parentPath == importPath {
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
	// leftSplit里面新建个容器叫做leftbottom,支持滚动
	leftbottom := container.NewVBox()
	// 设置文本对齐方式
	alignment := fyne.TextAlignCenter

	// 创建一个新的文本样式
	textStyle := fyne.TextStyle{Bold: true, Italic: true}

	// 创建一个新的标签,可以滚动
	label := widget.NewLabelWithStyle("你的文本", alignment, textStyle)

	// 将标签添加到 leftbottom 容器
	leftbottom.Add(label)
	//设置leftbottom的字体大小

	output := widget.NewMultiLineEntry()
	output.SetPlaceHolder("键入命令")
	output.Wrapping = fyne.TextWrapWord
	output.Enable()

	// 创建中间的编辑器
	editorVim := widget.NewMultiLineEntry()
	editorVim.SetPlaceHolder("")
	editorVim.Wrapping = fyne.TextWrapWord

	// 创建新的按钮
	debugButton := widget.NewButton("GPT", func() {
		// 按钮的点击事件处理函数
	})

	executeButton := widget.NewButton("执行", func() {
		// 按钮的点击事件时将globalproject压缩并上传到网盘
		// 压缩文件夹
		bdfs.Zipit(globalProject, globalProject+".zip")
		err = bdfs.CreateDir(uuidStr)
		if err != nil {
			fmt.Println("failed to create dir:", err)
		}
		// 上传文件夹
		err = bdfs.Upload(globalProject+".zip", uuidStr)
		if err != nil {
			fmt.Println("failed to upload file:", err)
		}
		// 删除本地压缩文件
		err = os.Remove(globalProject + ".zip")
		if err != nil {
			fmt.Println("failed to remove file:", err)
		}
		// 下载文件夹
		//get folder from path
		projectFolder := filepath.Base(globalProject)
		fmt.Println("globalProject:", projectFolder)
		err = bdfs.Download(uuidStr, projectFolder+".zip")
		if err != nil {
			fmt.Println("failed to download file:", err)
		}

	})
	// 竖直布局
	buttonBox := container.NewVBox(debugButton, executeButton)

	newBottom := container.NewHSplit(
		output,
		buttonBox,
	)

	newBottom.Offset = 0.9 // 设置 output 和 rightButton 的大小比例为 9:1

	// 和editorVim按1：9的比例合并
	leftConn := container.NewHSplit(leftline, editorVim)
	leftConn.Offset = 0.01
	//监听滚动事件
	// 创建一个支持滚动的容器，然后将 editorVim 添加到这个容器中
	scrollableEditorVim := container.NewHScroll(leftConn)

	// 创建新的显示区域,可以滚动但是不能编辑
	// 创建一个新的显示区域
	label = widget.NewLabel("输出区域")

	// 创建一个可以滚动的容器
	displayArea := container.NewVScroll(label)

	// 将scrollableEditorVim和displayArea添加到新的HSplit中
	HSplit := container.NewVSplit(
		scrollableEditorVim,
		displayArea,
	)
	HSplit.Offset = 0.7 // 设置 scrollableEditorVim 和 displayArea 的大小比例为 9:1
	editorVimSplit := container.NewVSplit(
		HSplit,
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

					// 更新全局变量为选择的路径
					globalFolderPath = uri.Path()
					globalProject = uri.Path()
					globalEditorVim = editorVim
					//空fyne.Container
					// 从 leftSplit 中移除旧的 leftbottom 容器
					leftSplit.Remove(leftbottom)
					// 创建一个新的 leftbottom 容器来替换旧的
					leftbottom = fyne.NewContainerWithLayout(layout.NewVBoxLayout())
					globalLeftbottom = leftbottom
					importPath = uri.Path()

					showFolderContents(globalFolderPath, globalEditorVim, globalLeftbottom)

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
