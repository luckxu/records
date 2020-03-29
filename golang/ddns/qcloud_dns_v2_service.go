package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"math/rand"
	"net/http"
	url2 "net/url"
	"sort"
	"strconv"
	"time"
)

type QcloudCnsV2Service struct {
	info ServiceInfo
}

type QcloudCnsRequest struct {
	svs    QcloudCnsV2Service
	args   map[string]string
	action string
	rec    *DnsRecord
}

type qcloudRecorder struct {
	ID      int64  `json:"id"`
	Value   string `json:"value"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Enabled int    `json:"enabled"`
	Status  string `json:"status"`
}

type qcloudDomain struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type qcloudListData struct {
	Domain  qcloudDomain     `json:"domain"`
	Records []qcloudRecorder `json:"records"`
}

type qcloudListResp struct {
	Code     int            `json:"code"`
	CodeDesc string         `json:"codeDesc"`
	Data     qcloudListData `json:"data"`
}

type qcloudNormalData struct {
	Record qcloudRecorder `json:"record"`
}

type qcloudNormalResp struct {
	Code     int              `json:"code"`
	CodeDesc string           `json:"codeDesc"`
	Data     qcloudNormalData `json:"data"`
}

func (svs QcloudCnsV2Service) newReq(action string, rec *DnsRecord) *QcloudCnsRequest {
	return &QcloudCnsRequest{args: make(map[string]string), action: action, rec: rec, svs: svs}
}

func (req *QcloudCnsRequest) buildUrl() string {
	keys := make([]string, len(req.args))
	i := 0
	for k, _ := range req.args {
		keys[i] = k
		i++
	}

	sort.Strings(keys)

	str := ""
	for _, k := range keys {
		v, _ := req.args[k]
		str += "&" + k + "=" + v
	}
	url := "cns.api.qcloud.com/v2/index.php?"

	h := hmac.New(sha1.New, []byte(req.svs.info.SecretKey))
	url += str[1:]
	sigData := "GET" + url

	h.Write([]byte(sigData))
	url = "https://" + url
	b64 := base64.StdEncoding
	sig := b64.EncodeToString(h.Sum(nil))
	sig = url2.QueryEscape(sig)
	url += "&Signature=" + sig
	log.Debugf("request action:%s, url:%s", req.action, url)
	return url
}

func (req *QcloudCnsRequest) request() []byte {
	req.args["Action"] = req.action
	if len(req.svs.info.Region) > 0 {
		req.args["Region"] = req.svs.info.Region
	}
	req.args["Timestamp"] = strconv.FormatInt(time.Now().Unix(), 10)
	req.args["Nonce"] = strconv.Itoa(rand.Int() & 0x7fffffff)
	req.args["SecretId"] = req.svs.info.SecretID
	req.args["SignatureMethod"] = "HmacSHA1"
	url := req.buildUrl()
	resp, err := http.Get(url)
	if err != nil {
		log.Warnf("Get failed, url:%s, error:%s", url, err.Error())
		return []byte("")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warnf("read response failed, url:%s, error:%s", url, err.Error())
		return []byte("")
	}
	return body
}

//api document:https://cloud.tencent.com/document/product/302/8516
func (svs QcloudCnsV2Service) Create(rec *DnsRecord) bool {
	req := svs.newReq("RecordCreate", rec)
	req.args["domain"] = rec.Name
	req.args["subDomain"] = rec.SubName
	req.args["recordType"] = rec.Type
	req.args["recordLine"] = "默认"
	req.args["value"] = rec.Value
	body := req.request()
	if len(body) == 0 {
		log.Warnf("create record failed, domain:%s.%s", rec.SubName, rec.Name)
		return false
	}
	log.Debugf("create response:%s", string(body))
	resp := qcloudNormalResp{}
	json.Unmarshal(body, &resp)
	if resp.Code == 0 {
		rec.ID = strconv.FormatInt(resp.Data.Record.ID, 10)
		return true
	}
	return false
}

//api document:https://cloud.tencent.com/document/product/302/8514
func (svs QcloudCnsV2Service) Delete(rec *DnsRecord) bool {
	req := svs.newReq("RecordDelete", nil)
	req.args["domain"] = rec.Name
	req.args["recordId"] = rec.ID
	body := req.request()
	if len(body) == 0 {
		return false
	}
	log.Debugf("delete response:%s", string(body))
	resp := qcloudListResp{}
	json.Unmarshal(body, &resp)
	if resp.Code == 0 {
		rec.ID = ""
		return true
	}
	return false
}

//api document:https://cloud.tencent.com/document/product/302/8511
func (svs QcloudCnsV2Service) Update(rec *DnsRecord) bool {
	req := svs.newReq("RecordModify", rec)
	req.args["domain"] = rec.Name
	req.args["recordId"] = rec.ID
	req.args["recordType"] = rec.Type
	req.args["recordLine"] = "默认"
	req.args["subDomain"] = rec.SubName
	req.args["value"] = rec.Value
	body := req.request()
	if len(body) == 0 {
		return false
	}
	log.Debugf("update record:%s", string(body))
	resp := qcloudNormalResp{}
	json.Unmarshal(body, &resp)
	if resp.Code == 0 {
		return true
	}
	return false
}

//api document:https://cloud.tencent.com/document/product/302/8517
func (svs QcloudCnsV2Service) View(rec *DnsRecord) bool {
	rec.ID = ""
	req := svs.newReq("RecordList", rec)
	req.args["domain"] = rec.Name
	req.args["subDomain"] = rec.SubName
	body := req.request()
	if len(body) == 0 {
		log.Warnf("view record failed, domain:%s.%s", rec.SubName, rec.Name)
		return false
	}
	log.Debugf("view record:%s", string(body))
	resp := qcloudListResp{}
	json.Unmarshal(body, &resp)
	for _, v := range resp.Data.Records {
		if v.Type == "A" {
			rec.ID = strconv.FormatInt(v.ID, 10)
			rec.Value = v.Value
			rec.Type = "A"
			if v.Enabled == 1 {
				log.Debugf("record:%v", v)
				return true
			}
		}
	}
	log.Debugf("record:%v", rec)
	return rec.ID != ""
}
