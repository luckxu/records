package main

import (
	"encoding/json"
	"flag"
	log "github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"io/ioutil"
	"os"
)

type Config struct {
	Oneshot          bool          `json:"oneshot"`
	ServiceProviders []ServiceInfo `json:"service_providers"`
}

func init() {
	log.SetFormatter(&prefixed.TextFormatter{
		DisableColors: true,
		TimestampFormat : "2006-01-02 15:04:05",
		FullTimestamp:true,
		ForceFormatting: true,
	})
}

func parseArg() *Config {
	lvl := flag.String("log", "info", "log level, can be error, warn, info, debug")
	cfg := flag.String("config", "/etc/cloud.ddns.conf", "set config file path")
	flag.Parse()

	switch *lvl {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	}
	body, err := ioutil.ReadFile(*cfg)
	if err != nil {
		log.Errorf("read configuration file failed, error:%s", err.Error())
		os.Exit(-1)
	}
	config := Config{}
	err = json.Unmarshal(body, &config)
	if err != nil {
		log.Errorf("json.Unmarshal failed, configuration file content must be json type, error:%s", err.Error())
		os.Exit(-2)
	}
	return &config
}

func main() {
	cfg := parseArg()
	cnt := len(cfg.ServiceProviders)
	sig := make(chan byte, cnt)
	for _, v := range cfg.ServiceProviders {
		go v.Handle(sig, cfg.Oneshot)
	}

	for cnt > 0 {
		select {
		case <-sig:
			cnt--
		}
	}
}
