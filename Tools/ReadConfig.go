package Tools

import (
	"github.com/BurntSushi/toml"
	"log"
)

type tomlConfig struct {
	Pulse Pulse
}
type Pulse struct {
	KafkaThread   int    `toml:"kafka_thread"`
	KafkaAddress  string `toml:"kafka_addr"`
	LogFiles      string `toml:"log_files"`
	LogPath       string `toml:"log_path"`
	LogLevel      string `toml:"log_level"`
	EtcdAddress   string `toml:"etcd_address"`
	EtcdWatchKey  string `toml:"etcd_watch_key"`
	EtcdHealthKey string `toml:"etcd_health_key"`
	EtcdTimeout   int    `toml:"etcd_timeout"`
	StartType     string `toml:"start_type"`
}

func GetConfig() Pulse {
	var favorites tomlConfig
	if _, err := toml.DecodeFile("conf.toml", &favorites); err != nil {
		log.Fatal(err)
	}
	return favorites.Pulse;
}
