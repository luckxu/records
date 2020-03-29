## 说明

1. web/mysql/grp: 基于容器实现Mysql Group Replication部署测试，运行build.sh <node count>启动<node count>个docker容器并将第一个设置为GRP主写节点；./clean <node count>停止并删除创建的容器和数据。