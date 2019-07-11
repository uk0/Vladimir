package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/fsnotify/fsnotify"
	"github.com/hpcloud/tail"
	"github.com/uk0/Octopoda/Tools"
	"strings"
	"sync"
)

type TailMgr struct {
	//因为我们的agent可能是读取多个日志文件，这里通过存储为一个map
	tailObjMap map[string]*TailObj
	lock       sync.Mutex
}

type TailObj struct {
	//这里是每个读取日志文件的对象
	tail     *tail.Tail
	secLimit *SecondLimit
	offset   int64 //记录当前位置
	//filename string
	logConf  logConfig
	exitChan chan bool
	ktype    string //发送方式 http kafka
}
// 保存采集文件夹的配置
var filesChanWatchMap = map[string]logConfig{}
// 监控那个文件夹[只监控文件夹]
var filesChanWatch = make(chan logConfig)
var tailMgr *TailMgr
var fileOffSets = make(chan OffsetConfig)
// 构建系统协程
var waitGroup sync.WaitGroup

// 创建一个TailManager
func NewTailMgr() (*TailMgr) {
	tailMgr = &TailMgr{
		tailObjMap: make(map[string]*TailObj, 16),
	}
	return tailMgr
}
// 启动新的采集任务（实时监控调用的就是整个方法）
func (t *TailMgr) AddLogFile(conf logConfig, offsetConf OffsetConfig) (err error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	_, ok := t.tailObjMap[conf.LogPath]
	if ok {
		err = fmt.Errorf("duplicate filename:%s  type:%s\n", conf.LogPath, conf.Type)
		return
	}
	tail2, err := tail.TailFile(conf.LogPath, tail.Config{
		ReOpen:    true,
		Follow:    true,
		Location:  &tail.SeekInfo{Offset: offsetConf.Offset, Whence: 1},
		MustExist: false,
		Poll:      true,
	})
	//var offsetNum, _ = tail2.Tell()

	tailobj := &TailObj{
		secLimit: NewSecondLimit(int32(conf.SendRate)),
		logConf:  conf,
		offset:   offsetConf.Offset,
		tail:     tail2,
		ktype:    conf.Type,
		exitChan: make(chan bool, 1),
	}
	t.tailObjMap[conf.LogPath] = tailobj
	logs.Info("map [%s]", t.tailObjMap)
	go tailobj.readLog()
	return
}
// 重新加载采集配置信息 实时加载
func (t *TailMgr) reloadConfig(logConfArr []logConfig) (err error) {
	for _, conf := range logConfArr {
		tailObj, ok := t.tailObjMap[conf.LogPath]
		fmt.Println(fmt.Sprintf("Reload Config On %s Config %v", Tools.GetNowTime(), conf))
		if !ok {

			logs.Info("conf:%v -- tailobj:%v", conf, tailObj)
			if strings.HasSuffix(conf.LogPath,"/"){
				logs.Info("LogPath Dir [%s] [True] start watcher Files... ",conf.LogPath)
				filesChanWatch<-conf
				conf.Ip=""
				filesChanWatchMap[conf.LogPath]=conf;
				continue //忽略文件夹
			}
			filename := strings.Split(conf.LogPath, "/")
			health_key := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s%s", filename[len(filename)-1], conf.Type)))
			OffsetConf := LoadOffSet(health_key)
			logs.Error(" load OffsetConf key  %s", health_key)
			fmt.Println(fmt.Sprintf("Get health FileName  %s   Offset %d", conf.LogPath, OffsetConf.Offset))
			err = t.AddLogFile(conf, OffsetConf)
			if err != nil {
				logs.Error("add log file failed,err:%v", err)
				continue
			}
			continue
		}
		tailObj.logConf = conf
		t.tailObjMap[conf.LogPath] = tailObj
		logs.Info(t.tailObjMap)
	}
	// 处理删除的日志收集配置
	for key, tailObj := range t.tailObjMap {
		var found = false
		for _, newValue := range logConfArr {
			if key == newValue.LogPath {
				found = true
				break
			}
		}
		if found == false {
			logs.Warn("log path:%s is remove", key)
			tailObj.exitChan <- true
			delete(t.tailObjMap, key)
		}
	}

	return
}

func (t *TailMgr) Process() {
	logChan := GetLogConf()
	for conf := range logChan {
		logs.Debug("log conf :%v", conf)
		var logConfArr []logConfig
		err := json.Unmarshal([]byte(conf), &logConfArr)
		if err != nil {
			logs.Error("unmarshal failed,err:%v conf:%v", err, conf)
			continue
		}
		logs.Debug("unmarshal succ conf:%v", logConfArr)
		err = t.reloadConfig(logConfArr)
		if err != nil {
			logs.Error("realod config from etcd failed err:%v", err)
			continue
		}
		logs.Info("[reaload] from etcd success,config:%v", logConfArr)
	}
}
/**
监控文件夹下是否有新的文件生成
**/
func (t *TailMgr)  WatcherLogDir() {
	logs.Info("[watch] start watch dir ...")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logs.Error(err)
	}
	defer watcher.Close()
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					logs.Info("[watch] create file:", event.Name)
					var key = event.Name[0:strings.LastIndex(event.Name,"/")+1] // find `/`
					logs.Info("[watch] Add Tail File %s",key)
					var config = filesChanWatchMap[key];
					config.LogPath=event.Name //modify value
					// todo
					AddConfig(config)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logs.Info("error:", err)
			}
		}
	}()
	for config:=range filesChanWatch{
		err = watcher.Add(config.LogPath)
		if err != nil {
			logs.Error(err)
		}
	}
	waitGroup.Done()
}

func (t *TailObj) readLog() {
	//读取每行日志内容
	fmt.Println("[ReadLog] start readLog now")
	for line := range t.tail.Lines {
		if line.Err != nil {
			logs.Error("[ReadLog] read line failed,err:%v", line.Err)
			continue
		}
		str := strings.TrimSpace(line.Text)
		if len(str) == 0 || str[0] == '\n' {
			continue
		}
		mns, _ := t.tail.Tell()
		// 每次都丢进去Chan
		localIp, _ := getLocalIP();
		filename := strings.Split(t.tail.Filename, "/")
		fileOffSets <- OffsetConfig{
			Offset:    mns,
			Key:       fmt.Sprintf("/pulse/%s/watch/%s/health", localIp[0], base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s%s", filename[len(filename)-1], t.ktype)))),
			Time:      Tools.GetNowTime(),
			Topic:     t.logConf.Topic,
			FilePatch: t.tail.Filename,
		}
		if "kafka" == t.ktype {
			kafkaSender.addMessage(line.Text, t.logConf.Topic)
			t.secLimit.Add(1)
			t.secLimit.Wait()
		}
		if "http" == t.ktype {
			HttpSender.addMessage(line.Text, t.logConf.Topic)
			t.secLimit.Add(1)
			t.secLimit.Wait()
		}
		select {
		case <-t.exitChan:
			logs.Warn("tail obj is exited,config:%v", t.logConf)
			return
		default:
		}

	}
	waitGroup.Done()
}

func RunServer() {
	tailMgr = NewTailMgr()
	go tailMgr.WatcherLogDir() //关注文件夹
	go AutoSaveOffset(fileOffSets)
	tailMgr.Process()
	waitGroup.Wait()
}
