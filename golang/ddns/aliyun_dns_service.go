package main

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	log "github.com/sirupsen/logrus"
	"strings"
)

type AliyunDnsService struct {
	info ServiceInfo
}

//api document:https://help.aliyun.com/document_detail/29772.html?spm=a2c4g.11186623.6.657.404d3b59g6iQth
func (svs AliyunDnsService) Create(rec *DnsRecord) bool {
	client, err := alidns.NewClientWithAccessKey(svs.info.Region, svs.info.SecretID, svs.info.SecretKey)
	if err != nil {
		log.Warnf("create aliyun client failed, error:%s", err.Error())
		return false
	}
	request := alidns.CreateAddDomainRecordRequest()
	request.SetContentType("application/json")
	request.Scheme = "https"
	request.DomainName = rec.Name
	request.RR = rec.SubName
	request.Type = rec.Type
	request.Value = rec.Value
	resp, err := client.AddDomainRecord(request)
	if err != nil {
		log.Warnf("create record failed, domain:%s.%s, error:%s", rec.SubName, rec.Name, err.Error())
	}
	if resp.IsSuccess() {
		rec.ID = resp.RecordId
		return true
	}
	return false
}

//api document:https://help.aliyun.com/document_detail/29773.html?spm=a2c4g.11186623.6.658.7f672846vRx63n
func (svs AliyunDnsService) Delete(rec *DnsRecord) bool {
	client, err := alidns.NewClientWithAccessKey(svs.info.Region, svs.info.SecretID, svs.info.SecretKey)
	if err != nil {
		log.Warnf("create aliyun client failed, error:%s", err.Error())
		return false
	}
	request := alidns.CreateDeleteDomainRecordRequest()
	request.SetContentType("application/json")
	request.Scheme = "https"
	request.RegionId = rec.ID
	resp, err := client.DeleteDomainRecord(request)
	if err != nil {
		log.Warnf("delete record failed, domain:%s.%s, error:%s", rec.SubName, rec.Name, err.Error())
		return false
	}
	if resp.IsSuccess() {
		rec.ID = ""
		return true
	}
	return false
}

//api document:https://help.aliyun.com/document_detail/29774.html?spm=a2c4g.11186623.6.659.b7dd3192TKPBLR
func (svs AliyunDnsService) Update(rec *DnsRecord) bool {
	client, err := alidns.NewClientWithAccessKey(svs.info.Region, svs.info.SecretID, svs.info.SecretKey)
	if err != nil {
		log.Warnf("create aliyun client failed, error:%s", err.Error())
		return false
	}
	request := alidns.CreateUpdateDomainRecordRequest()
	request.SetContentType("application/json")
	request.Scheme = "https"
	request.RecordId = rec.ID
	request.RR = rec.SubName
	request.Type = rec.Type
	request.Value = rec.Value
	resp, err := client.UpdateDomainRecord(request)
	if err != nil {
		log.Warnf("update record failed, domain:%s.%s, error:%s", rec.SubName, rec.Name, err.Error())
		return false
	}
	if resp.IsSuccess() {
		return true
	}
	return false
}

//api document:https://help.aliyun.com/document_detail/29778.html?spm=a2c4g.11186623.6.656.65d31cebfd36FR
func (svs AliyunDnsService) View(rec *DnsRecord) bool {
	rec.ID = ""
	client, err := alidns.NewClientWithAccessKey(svs.info.Region, svs.info.SecretID, svs.info.SecretKey)
	if err != nil {
		log.Warnf("create aliyun client failed, error:%s", err.Error())
		return false
	}
	request := alidns.CreateDescribeSubDomainRecordsRequest()
	request.SetContentType("application/json")
	request.SubDomain = rec.SubName + "." + rec.Name
	resp, err := client.DescribeSubDomainRecords(request)
	if err != nil {
		log.Warnf("view record failed, domain:%s.%s, error:%s", rec.SubName, rec.Name, err.Error())
	}
	for _, v := range resp.DomainRecords.Record {
		if v.Type == "A" {
			rec.ID = v.RecordId
			rec.Value = v.Value
			rec.Type = "A"
			if strings.ToLower(v.Status) == "enable" {
				log.Debugf("record:%v", v)
				return true
			}
		}
	}
	log.Debugf("record:%v", rec)
	return rec.ID != ""
}
