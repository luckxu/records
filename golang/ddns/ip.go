package main

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"regexp"
	"time"
)

var urls = [...]string{
	"http://ident.me",
	"http://ipecho.net/plain",
	"http://whatismyip.akamai.com",
	"http://tnx.nl/ip",
	"http://myip.dnsomatic.com",
	"http://ifconfig.me",
	"http://checkip.dyndns.com",
	"http://myip.ipip.net",
	"http://icanhazip.com",
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func findMyIpAddress() (ip string, ok bool) {
	realIP := ""
	for {
		url := urls[rand.Int()%len(urls)]
		resp, err := http.Get(url)
		if err != nil {
			log.Debugf("http request failed, url:%s, error:%s", url, err.Error())
			continue
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Debugf("read response failed, url:%s, error:%s", url, err.Error())
			continue
		}
		rc, _ := regexp.Compile("[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}")
		log.Debugf("url:%s, response:%s", url, body)
		for _, ip := range rc.FindStringSubmatch(string(body)) {
			if net.ParseIP(ip) != nil {
				log.Infof("my ip address:%s", ip)
				if ip == realIP {
					return ip, true
				} else {
					realIP = ip
				}
			}
		}
	}
	return "", false
}
