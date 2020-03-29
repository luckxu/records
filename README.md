## 说明

1. web/mysql/grp:基于容器测试Mysql Group Replication部署，运行build.sh [containers count]启动[containers count]个mysql:8.0容器并将第一个设置为GRP主写节点；./clean [containers count]停止并删除创建的容器和数据。
2. web/zookeeper:基础容器测试zookeeper，运行build.sh [containers count]启动[containers count]个zookeeper:3.6 容器节点；./clean [containers count]停止并删除创建的容器和数据。
3. golang/ddns:查询本机的公网IP地址，依托云服务厂商提供的DNS解析管理接口实现DDNS动态域名解析服务，目前支持腾讯云和阿里云。