## DDNS动态域名解析

一般家用路由器都有DDNS动态域名解析服务，在路由器不支持DDNS服务或DDNS服务不佳时可以考虑使用本程序调用公有云服务厂商提供的DNS解析服务实现动态公网IP地址获取和解析。

## 配置文件

Linux环境默认配置文件为`/etc/cloud.ddns.conf`，也可通过-config参数指定配置文件，配置文件格式：
```json
{
	"oneshot": false,
	"service_providers": [
        {
            "provider": "qcloud or aliyun",
            "domain": "example.com",
            "sub_domain": "www",
            "secret_id": "fill real secret_id",
            "secret_key": "fill real secret_key",
            "region": "file real region"
        },
        ...
    ]
}
```

- oneshot：单次执行标识，可选true或false。域名解析配置成功后程序退出，用于crontab等单次调度环境
- provider：DNS解析服务提供商，目前支持qcloud(腾讯云)和aliyun(阿里云)
- domain：域名
- sub_domain: 主机记录
- secret_id和secret_key：接口调用凭证，阿里云从`https://usercenter.console.aliyun.com/#/manage/ak`获取，腾讯云从`https://console.cloud.tencent.com/cam/capi`获取
- region：分区，aliyun需要配置成`cn-hangzhou`, 腾讯云配置为空字符串即可

## 部署

首先将ddns程序上传至服务器，赋予可执行权限，填写正确的配置文。假定程序放置`/usr/sbin/ddns`，配置文件放置`/etc/cloud.ddns.conf`，可参考如下方式配置运行：

* crontab

    crontab能够定时、周期性执行脚本或程序，因此需要将oneshot配置为false。执行`vi /etc/crontab`，添加行`*/1 * * * * root /usr/sbin/ddns&`
   
* supervisor

    未安装supervisor需要先安装，随后在`/etc/supervisor/conf.d/`目录下新增`ddns.conf`文件，内容如下:
    ```
    ; /etc/supervisor/conf.d/ddns.conf
    
    [program:ddns]
    command         = /usr/sbin/ddns
    user            = root
    stdout_logfile  = /var/log/ddns
    autorestart     = true
    ```
    执行supervisorctl reload 加载新的配置。
    
## 其它说明

1. 当云解析的主机A记录有多条时优先选择启用状态的A记录；如果所有A记录都未启用，则随机选择一条A记录并将其记录值修改成公网IP地址；当不存在任何A记录时，则创建新的A记录。
2. 程序每60秒检查一次外网IP，每600秒检查一次DNS存储的A记录，当外网IP变更或与A记录不一致时，更新或创建域名解析A记录。