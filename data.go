package main

type logConfig struct {
	Type     string `json:"type"`     // 采集方式
	Topic    string `json:"topic"`    // 主题
	LogPath  string `json:"log_path"` //采集地址
	Service  string `json:"service"`
	Regexp     string `json:"regexp"`      //日志正则表达式
	SendRate int    `json:"send_rate"` // 速率
	Ip       string `json:"host"`      // 主机
	Dir      bool   `json:"is_dir"`    //是否是文件夹
}

type OffsetConfig struct {
	Offset    int64
	FilePatch string
	Key       string
	Time      string
	Topic     string
	TPS       int64
	QPS       int64
}
