package main

import (
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"
)

var changedFile []string

func ListenFsNotify(globalProject string, changedFile []string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					// 检查 event.Name 是否已经存在于 changedFile 中
					exists := false
					for _, file := range changedFile {
						if file == event.Name {
							exists = true
							break
						}
					}

					// 如果 event.Name 不存在于 changedFile 中，将其添加到 changedFile
					if !exists {
						changedFile = append(changedFile, event.Name)
					}
					fmt.Println(changedFile)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(globalProject)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func main() {
	globalProject := "./"
	ListenFsNotify(globalProject, changedFile)
}
