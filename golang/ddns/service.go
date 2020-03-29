package main

import (
	log "github.com/sirupsen/logrus"
	"time"
)

type DnsRecord struct {
	ID      string
	SubName string
	Name    string
	Value   string
	Type    string
}

type DnsService interface {
	Create(rec *DnsRecord) bool
	Delete(rec *DnsRecord) bool
	Update(rec *DnsRecord) bool
	View(rec *DnsRecord) bool
}

type ServiceInfo struct {
	Domain    string `json:"domain"`
	SubDomain string `json:"sub_domain"`
	Provider  string `json:"provider"`
	SecretID  string `json:"secret_id"`
	SecretKey string `json:"secret_key"`
	Region    string `json:"region"`
}

func handle(sig chan byte, svs DnsService, rec *DnsRecord, onshot bool) {
	lastViewAt := time.Now().Unix()
	for {
		ip, ok := findMyIpAddress()
		if !ok {
			time.Sleep(5 * time.Second)
			continue
		}
		if rec.ID == "" || lastViewAt+600 < time.Now().Unix() {
			svs.View(rec)
		}
		if rec.ID != "" {
			if rec.Value != ip {
				rec.Value = ip
				if ok := svs.Update(rec); !ok {
					rec.ID = ""
				}
			}
		} else {
			rec.Value = ip
			rec.Type = "A"
			svs.Create(rec)
		}
		log.Infof("record info, name:%s.%s, value:%s", rec.SubName, rec.Name, rec.Value)
		if onshot && ip == rec.Value {
			break
		}
		time.Sleep(60 * time.Second)
	}
	sig <- '0'
}

func (info ServiceInfo) Handle(sig chan byte, oneshot bool) {
	rec := &DnsRecord{Name: info.Domain, SubName: info.SubDomain}
	switch info.Provider {
	case "qcloud":
		svs := QcloudCnsV2Service{info: info}
		go handle(sig, svs, rec, oneshot)
	case "aliyun":
		svs := AliyunDnsService{info: info}
		go handle(sig, svs, rec, oneshot)
	default:
		return
	}
}
