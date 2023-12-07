package main

import (
	"context"
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sharedgpu/bdfs"
	"sharedgpu/db"
	"sharedgpu/proxy"
	"sharedgpu/utils"
	"strings"
	"sync/atomic"
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
	leftline          = widget.NewMultiLineEntry()
	labelout          *widget.Label
	bottomPart        *widget.Label
	bottomInput       *widget.Entry
	uuidStr           string
	isOccupied        bool
	isShared          int32
	mountedMachine    []string
	selectedValue     string
	selectmachineName string
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
	ctxTask, cancelTask := context.WithCancel(ctx)

	rdb, err := db.InitRedis(ctx)
	if err != nil {
		fmt.Println("redis init failed")
	}

	bduss, err := rdb.Get(ctx, "bduss").Result()
	if err != nil {
		fmt.Println("failed to get bduss:", err)
	}
	err = bdfs.LoginBd(bduss)
	if err != nil {
		fmt.Println("failed to login bd:", err)
	}
	// 挂载本地机器
	mountedMachine = append(mountedMachine, "local")

	// 创建添加和删除机器的按钮，并设置颜色
	addMachineButton := widget.NewButton("租用机器", func() {
		// 在这里添加机器的代码
		// 重置取消gorutine
		ctxTask, cancelTask = context.WithCancel(ctx)
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

			// 检查是否处于占用状态
			if isOccupied {
				return
			}

			var startTime time.Time

			if startTime.IsZero() {
				startTime = time.Now()
			}

			//想把下面的程序改成0.5s执行一次去redis里面查询是否有匹配的机器，如果有则显示连接成功，如果没有则显示暂无资源

			var ConnKey string

			go func(ctx context.Context) {
				ticker := time.NewTicker(500 * time.Millisecond)
				defer ticker.Stop()

				for {
					select {
					case <-ctxTask.Done():
						fmt.Println("Goroutine cancelled")
						// 重置 context 和 cancel 函数
						ctxTask, cancelTask = context.WithCancel(ctx)
						// 更新 isOccupied 的值
						isOccupied = false
						return
					case <-ticker.C:
						// 从 Redis 中获取所有包含 gpuSelect.Selected 的键
						result, err := db.HgetallByValue(ctx, rdb, "gpu", gpuSelect.Selected)
						if err != nil {
							labelout.SetText("没有匹配的机器。。。")
						}
						//如果result中有key则看一下是不是跟ConnKey一致，如果一致且stauts是1则跳出循环，如果是0，则匹配此机器，将状态置为1
						if len(result) == 0 {
							if ConnKey == "" {
								labelout.SetText("没有合适机器")
							} else {
								status, err := rdb.HGet(ctx, ConnKey, "status").Result()
								if err != nil {
									startTime = time.Now()

									labelout.SetText("暂无机器可以使用。。。")
									continue
								}
								if status == "1" {
									continue
								} else {
									labelout.SetText("机器重连失败")
								}

							}
						} else {
							for _, key := range result {
								// 获取 status 字段的值
								status, err := rdb.HGet(ctx, key, "status").Result()
								if err != nil {
									labelout.SetText("暂无机器可以使用。。。")
									continue
								}
								// 如果键与 ConnKey 一致
								if key == ConnKey {
									// 如果 status 为 1，跳出循环
									if status == "1" {
										// 计算时间差
										duration := time.Since(startTime)

										// 计算天、小时、分钟和秒
										days := int(duration.Hours()) / 24
										hours := int(duration.Hours()) % 24
										minutes := int(duration.Minutes()) % 60
										seconds := int(duration.Seconds()) % 60

										// 显示连接成功和时间差还有gpu相关信息
										labelout.SetText(fmt.Sprintf("连接成功，已连接：%d天%d小时%d分钟%d秒，\nGPU:%s", days, hours, minutes, seconds, gpuSelect.Selected))
										mountedMachine = append(mountedMachine, gpuSelect.Selected)
										//uuidStr更新为result中的key
										uuidStr = key
										continue
									}
									// 如果 status 为 0，匹配此机器，并将状态置为 1
									if status == "0" {
										err := rdb.HSet(ctx, key, "status", "1").Err()
										if err != nil {

											labelout.SetText("重新匹配机器成功。。。")
											uuidStr = key
											mountedMachine = append(mountedMachine, gpuSelect.Selected)

										}
										ConnKey = key

										// 创建 SSH 客户端配置
										config := &ssh.ClientConfig{
											User: "tian",
											Auth: []ssh.AuthMethod{
												ssh.Password("tian"),
											},
											HostKeyCallback: ssh.InsecureIgnoreHostKey(),
										}

										go func() {
											for range time.Tick(time.Second) {
												client, err := ssh.Dial("tcp", "127.0.0.1:3333", config)
												if err != nil {
													fmt.Println("连接失败: ", err)
													labelout.SetText("连接失败: " + err.Error())
												} else {
													client.Close()
												}
											}
										}()
									}
								} else {
									// 如果 status 为 0，匹配此机器，并将状态置为 1
									if status == "0" {
										err := rdb.HSet(ctx, key, "status", "1").Err()
										if err != nil {
											fmt.Println("failed to set status:", err)
										}
										ConnKey = key

										// 创建 SSH 客户端配置
										config := &ssh.ClientConfig{
											User: "tian",
											Auth: []ssh.AuthMethod{
												ssh.Password("tian"),
											},
											HostKeyCallback: ssh.InsecureIgnoreHostKey(),
										}

										go func() {
											for range time.Tick(time.Second) {
												client, err := ssh.Dial("tcp", "127.0.0.1:3333", config)
												if err != nil {
													fmt.Println("连接失败: ", err)
													labelout.SetText("连接失败: " + err.Error())
												} else {

													client.Close()
												}
											}
										}()
										uuidStr = key
										labelout.SetText("匹配机器成功。。。")
										mountedMachine = append(mountedMachine, gpuSelect.Selected)
									}
									if status == "1" {
										startTime = time.Now()
										labelout.SetText("暂无机器可以挂载...")
									}
								}
							}

						}

						isOccupied = true
					}
				}
			}(ctxTask)

		}, myWindow)

		dialog.Show()
	})
	addMachineButton.Importance = widget.HighImportance

	// 创建租用机器和管理数据集的按钮，并设置颜色
	rentMachineButton := widget.NewButton("出租机器", func() {

		if atomic.LoadInt32(&isShared) == 1 {
			return
		} else {
			// 在新的 goroutine 中运行 proxy.StartSShServer()
			go func() {
				errChan := proxy.StartSShServer()
				err = <-errChan
				if err != nil {
					fmt.Println("failed to start ssh server:", err)
				}

			}()
		}

		machineModel := "MachineModel123"
		uuidStr = utils.GenerateUUID(machineModel).String()

		// 先清空redis里面的uuidStr的信息
		err := rdb.Del(ctx, uuidStr).Err()
		if err != nil {
			fmt.Println("failed to del uuidStr:", err)
		}

		// 获取机器的 CPU、内存、显卡型号和个数
		cpuInfo, memoryInfo, gpuInfo, nil := utils.GetSystemInfo()
		if err != nil {
			panic(err)
		}

		fmt.Println("cpuInfo:", cpuInfo, "memoryInfo:", memoryInfo, "gpuInfo:", gpuInfo)
		//gpuinfo  from NVIDIA GeForce RTX 3080 Ti: 616 to RTX 3080 Ti
		gpuInfo = strings.Split(gpuInfo, ":")[0]
		fmt.Println("gpuInfo:", gpuInfo)
		// 将 uuid 和机器的 CPU、内存、显卡型号存入 redis 想变成json形式
		// 新加个字段status 0表示没有任务需要执行，1表示有任务需要执行，2表示任务执行完成
		// 新加个字段submitTime 用于记录任务提交时间
		// 新加个字段log 用于记录任务执行日志
		err = rdb.HSet(ctx, uuidStr, "cpu", cpuInfo, "memory", memoryInfo, "gpu", gpuInfo, "status", "0", "taskStatus", "0", "submitTime", time.Now().Format("2006-01-02 15:04:05"), "log", "testting").Err()
		if err != nil {
			fmt.Println("failed to set info to redis :", err)
		}

		var dashboardWindow fyne.Window
		// 创建一个新的窗口来显示仪表盘
		if atomic.LoadInt32(&isShared) == 1 {
			fmt.Println("仪表盘已创建")
		} else {
			atomic.StoreInt32(&isShared, 1)
			dashboardWindow = fyne.CurrentApp().NewWindow("仪表盘")
		}

		go func() {
			cpuLabel := widget.NewLabel("CPU: ")
			memoryLabel := widget.NewLabel("内存: ")
			diskLabel := widget.NewLabel("磁盘: ")
			networkLabel := widget.NewLabel("网络: ")
			gpuMemoryLabel := widget.NewLabel("GPU 内存: ")

			// 创建一个可以被取消的 context
			panCtx, panCancel := context.WithCancel(context.Background())
			// 在新的 goroutine 中定期更新仪表盘
			go func() {
				for {
					select {
					case <-panCtx.Done():
						// 如果 context 被取消，停止更新仪表盘
						return
					default:
						// 如果 isShared 为 0，停止更新仪表盘
						if atomic.LoadInt32(&isShared) == 0 {
							return
						}
						// 更新仪表盘
						// 这里需要你自己的函数来获取 CPU、内存、磁盘、网络和 GPU 内存的占用情况
						cpu, memory, disk, network, gpuMemory, err := utils.GetSystemUsage()
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

			unshareMachineButton := widget.NewButton("取消共享", func() {
				// 删除 Redis 中的 uuid
				err := rdb.Del(panCtx, uuidStr).Err()
				if err != nil {
					fmt.Println(fmt.Errorf("failed to delete uuid from redis: %w", err))
				}
				println("Machine is no longer shared.")

				// 关闭窗口
				dashboardWindow.Close()

				// 停止更新仪表盘
				panCancel()
				atomic.StoreInt32(&isShared, 0)

			})

			// 将 unshareMachineButton 添加到窗口的内容中
			dashboardWindow.SetContent(container.NewVBox(cpuLabel, memoryLabel, diskLabel, networkLabel, gpuMemoryLabel, unshareMachineButton))
			dashboardWindow.SetOnClosed(func() {
				// 删除 Redis 中的 uuid
				err := rdb.Del(panCtx, uuidStr).Err()
				if err != nil {
					fmt.Println(fmt.Errorf("failed to delete uuid from redis: %w", err))
				}
				println("Machine is no longer shared.")

				// 关闭窗口
				dashboardWindow.Close()

				// 停止更新仪表盘
				panCancel()
				atomic.StoreInt32(&isShared, 0)
			})

			dashboardWindow.Show()

		}()

		//创建一个新的循环线程去不停的轮询uuidStr下的任务状况,当taskStatus为1时则执行运行代码
		go func() {
			for {
				// 轮询 uuidStr 下的任务状态
				taskStatus, err := rdb.HGet(ctx, uuidStr, "taskStatus").Result()
				if err != nil {
					// 处理错误
					fmt.Println(err)
					continue
				}

				// 当 taskStatus 为 "1" 时，执行运行代码
				if taskStatus == "1" {

					//获取当前目录
					currentDir, err := os.Getwd()
					// 这里添加你的运行代码
					// 先下载代码和配置环境到本地tmp目录下面
					// 从网盘uuidStr目录下下载代码和环境
					// 执行本地机器的代码
					// 从redis中获取updateTime 和 submitTime
					updateTime, err := rdb.HGet(ctx, uuidStr, "updateTime").Result()
					if err != nil {
						fmt.Println("failed to get updateTime:", err)
					}
					submitTime, err := rdb.HGet(ctx, uuidStr, "submitTime").Result()
					if err != nil {
						fmt.Println("failed to get submitTime:", err)
					}
					//计算时间差
					updateTimeObj, err := time.Parse("2006-01-02 15:04:05", updateTime)
					if err != nil {
						fmt.Println("failed to parse updateTime:", err)
					}
					submitTimeObj, err := time.Parse("2006-01-02 15:04:05", submitTime)
					if err != nil {
						fmt.Println("failed to parse submitTime:", err)
					}
					duration := updateTimeObj.Sub(submitTimeObj)

					err = bdfs.Download(uuidStr, "", "./")
					if err != nil {
						fmt.Println("failed to download file:", err)
					}

					//新建文件夹currentdir
					// err = os.MkdirAll(currentDir, 0755)
					// if err != nil {
					// 	log.Fatal(err)
					// }

					//如果currentDir没有miniconda 则下载
					minicondaPath := filepath.Join(currentDir, "miniconda")
					_, err = os.Stat(minicondaPath)
					if os.IsNotExist(err) {
						err = bdfs.Download("miniconda", "miniconda.zip", currentDir)
						if err != nil {
							fmt.Println("failed to download file:", err)
						}

						if err != nil {
							fmt.Println("failed to get current dir:", err)
						}

					} else if err != nil {
						// 其他错误
						log.Fatal(err)
					}

					for {

						files, err := ioutil.ReadDir(currentDir)
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

						err = filepath.Walk(currentDir+"/"+uuidStr, func(path string, info os.FileInfo, err error) error {
							if err != nil {
								return err
							}
							if !info.IsDir() && strings.HasSuffix(info.Name(), ".zip") {
								// 使用完整路径解压文件
								err = bdfs.Unzip(path, currentDir)
								if err != nil {
									fmt.Println("failed to unzip file:", err)
								}
							}
							return nil
						})
						if err != nil {
							log.Fatal(err)
						}

						//删除miniconda.zip
						err = os.RemoveAll(filepath.Join(currentDir, "miniconda.zip"))
						if err != nil {
							fmt.Println("failed to remove miniconda.zip:", err)
						}
						break
					}

					//删除本地uuidStr下的压缩包
					err = os.RemoveAll(filepath.Join(currentDir, uuidStr))
					if err != nil {
						fmt.Println("failed to remove dir:", err)
					}

					// 如果一致则不需要更新，如果不一致则需要更新
					if duration.Seconds() == 0 {
						//删除uuidStr下的压缩包
						err = bdfs.DeleteDir(uuidStr)
						if err != nil {
							fmt.Println("failed to delete dir:", err)
						}
					}

					utils.ExecCommand(selectedValue, bottomInput, bottomPart, globalProject, uuidStr, rdb)

					//将taskStatus置为0
					err = rdb.HSet(ctx, uuidStr, "taskStatus", "0", "updateTime", time.Now().Format("2006-01-02 15:04:05")).Err()
					if err != nil {
						fmt.Println("failed to set taskStatus:", err)
					}
				}
			}
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

	bottomInput = widget.NewMultiLineEntry()
	bottomInput.SetPlaceHolder("键入命令")
	bottomInput.Wrapping = fyne.TextWrapWord
	bottomInput.Enable()

	// 创建中间的编辑器
	editorVim := widget.NewMultiLineEntry()
	editorVim.SetPlaceHolder("")
	editorVim.Wrapping = fyne.TextWrapWord

	// 创建新的按钮
	debugButton := widget.NewButton("GPT", func() {
		// 按钮的点击事件处理函数
	})

	executeButton := widget.NewButton("执行", func() {
		if selectedValue == "" {
			labelout.SetText("请选择机器")
			return
		}

		if selectedValue == "local" {
			utils.ExecCommand(selectedValue, bottomInput, bottomPart, globalProject, uuidStr, rdb)
		} else {
			err := bdfs.DeleteDir(uuidStr)
			if err != nil {
				fmt.Println("failed to delete dir:", err)
			}
			bottomPart.SetText("清空网盘任务。。。")
			// 压缩文件夹
			bdfs.Zipit(globalProject, globalProject+".zip")
			err = bdfs.CreateDir(uuidStr)
			if err != nil {
				fmt.Println("failed to create dir:", err)
			}
			bottomPart.SetText("压缩文件夹。。。")

			// 上传文件夹
			err = bdfs.Upload(globalProject+".zip", uuidStr)
			if err != nil {
				fmt.Println("failed to upload file:", err)
			}
			bottomPart.SetText("上传文件夹。。。")
			// 删除本地压缩文件
			err = os.Remove(globalProject + ".zip")
			if err != nil {
				fmt.Println("failed to remove file:", err)
			}
			// 更新redis uuid 下的任务状态为有任务需要执行，并将提交时间更新为当前时间
			err = rdb.HSet(ctx, uuidStr, "cmd", bottomInput.Text, "taskStatus", "1", "log", "", "updateTime", time.Now().Format("2006-01-02 15:04:05"), "submitTime", time.Now().Format("2006-01-02 15:04:05")).Err()
			if err != nil {
				fmt.Println("failed to set info to redis :", err)
			}
			bottomPart.SetText("任务已创建")
			//不停的轮询redis uuid 下的任务状态，如果为2，则下载文件夹
			for {
				// 获取任务状态
				status, err := rdb.HGet(ctx, uuidStr, "taskStatus").Result()
				if err != nil {
					fmt.Println("failed to get status:", err)
				}

				log, err := rdb.HGet(ctx, uuidStr, "log").Result()
				if err != nil {
					fmt.Println("failed to get status:", err)
				}
				bottomPart.SetText(log)

				if status == "0" {
					bottomPart.SetText("任务执行完成，获取结果。。。")
					// 获取执行日志
					log, err := rdb.HGet(ctx, uuidStr, "log").Result()
					if err != nil {
						fmt.Println("failed to get log:", err)
					}
					bottomPart.SetText(log)
					break
				}
			}
		}
	})
	// 竖直布局
	buttonBox := container.NewVBox(debugButton, executeButton)

	newBottom := container.NewHSplit(
		bottomInput,
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
	labelout = widget.NewLabel("输出区域")
	// 创建一个空的部件作为下部分
	bottomPart = widget.NewLabel("")
	// 创建一个按钮
	button := widget.NewButton("取消挂载机器", func() {
		//点击取消挂载机器的时候，将轮询redis的任务关掉
		cancelTask()
		// 将 Redis 中的 status 置为 0
		err := rdb.HSet(ctx, uuidStr, "status", "0").Err()
		if err != nil {
			// 处理错误
			fmt.Println(err)

		}
		// 如果 mountedMachine 的长度大于1，才执行删除操作
		if len(mountedMachine) > 1 {
			//将mountmachine中的机器去掉
			for i, machine := range mountedMachine {
				if machine == selectmachineName {
					// 删除这个机器
					mountedMachine = append(mountedMachine[:i], mountedMachine[i+1:]...)
					break
				}
			}
		}
		isOccupied = false
		labelout.SetText("机器挂载已取消")
	})

	// 将 labelout 和 button 添加到一个新的 HSplit 中
	topPart := container.NewHSplit(labelout, button)
	topPart.Offset = 0.9 // 设置 labelout 和 button 的大小比例为 9:1

	// 将 topPart 和 bottomPart 添加到一个新的 VSplit 中
	labeloutSplit := container.NewVSplit(topPart, bottomPart)
	labeloutSplit.Offset = 0.1 // 设置 topPart 和 bottomPart 的大小比例为 1:9

	// 创建一个可以滚动的容器
	displayArea := container.NewVScroll(labeloutSplit)

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

	button1 := widget.NewButtonWithIcon("选择机器", theme.ConfirmIcon(), nil)

	button1.OnTapped = func() {
		// 创建一个存储机器名的切片
		machineNames := make([]string, len(mountedMachine))
		for i, machine := range mountedMachine {
			machineNames[i] = machine // 直接使用 machine 作为机器名
		}

		// 创建一个 RadioGroup
		radio := widget.NewRadioGroup(machineNames, func(machineName string) {
			if machineName != "" {
				// 在这里处理选中的机器名
				//将选中的机器名赋值给selectedValue
				selectedValue = machineName
				selectmachineName = machineName
			}
		})

		// 创建一个新的弹出覆盖式窗口
		popUpContent := fyne.NewContainerWithLayout(layout.NewVBoxLayout(), radio) // 使用 RadioGroup 作为弹出窗口的内容
		canvas := fyne.CurrentApp().Driver().CanvasForObject(button1)
		popUp := widget.NewPopUp(popUpContent, canvas)

		// 设置弹出窗口的大小和位置
		popUp.Resize(fyne.NewSize(200, canvas.Size().Height)) // 设置弹出窗口的宽度为200，高度为画布的高度
		popUp.Move(fyne.NewPos(0, 0))                         // 将弹出窗口移动到画布的左上角

		// 显示弹出覆盖式窗口
		popUp.Show()
	}

	button2 := widget.NewButtonWithIcon("数据集上传", theme.UploadIcon(), func() {
		// 在这里处理用户点击 "Button 2" 的事件
	})

	// 创建一个新的 VBox 容器，包含你的菜单
	menu := container.NewVBox(
		addMachineButton,
		rentMachineButton,
		importButton,          // 新添加的按钮
		widget.NewSeparator(), // 添加一个分隔符
	)

	// 创建一个新的 VBox 容器，包含你的按钮
	buttons := container.NewVBox(
		button1, // 添加快捷按钮
		button2, // 添加快捷按钮
	)

	// 使用 container.NewBorder 创建一个新的容器，将菜单放在顶部，将按钮放在底部
	leftMenu := container.NewBorder(menu, buttons, nil, nil)

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
