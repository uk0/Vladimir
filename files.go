package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/astaxie/beego/logs"
	"strings"
	"sync"
)

var filesChan = make(chan string)

func WatcherLogDir() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logs.Error(err)
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
				if event.Op&fsnotify.Create == fsnotify.Create {
					logs.Info("create file:", event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logs.Info("error:", err)
			}
		}
	}()
	for dirPath:=range filesChan{
		err = watcher.Add(dirPath)
		if err != nil {
			logs.Error(err)
		}
	}

	<-done
}

func main2() {
	if strings.HasSuffix("/Users/zhangjianxin/home/GO_LIB/src/github.com/uk0/Octopoda/q.log", "/"){
		fmt.Println("dir ")
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		WatcherLogDir()
		wg.Done()
	}()
	go func() {
		filesChan<-"/Users/zhangjianxin/home/GO_LIB/src/github.com/uk0/Octopoda/release/logs/"
		wg.Done()
	}()

	wg.Wait()




}