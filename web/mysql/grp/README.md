## Mysql Group Replication

参考：
- [Mysql Group Replication#18.2 Getting Started](https://dev.mysql.com/doc/refman/8.0/en/group-replication-getting-started.html)
- [MySQL InnoDB Cluster : avoid split-brain while forcing quorum](https://lefred.be/content/mysql-innodb-cluster-avoid-split-brain-while-forcing-quorum/)
- [Docker#Mysql](https://hub.docker.com/_/mysql)
- [Overview of Docker Compose](https://docs.docker.com/compose/)
- [Docker#Networking Tutorials](https://docs.docker.com/network/network-tutorial-standalone/)


## 创建网络

`docker network create test_vpc --driver=bridge --subnet=10.10.0.0/16`

## 创建并配置主节点[单主模式]

1. 编辑mysql配置文件`/path/to/mode1/mysql/conf.d/mysql.cnf`:
```mysql
[mysqld]

# resolv problem: IP address 'xx.xx.xx.xx' could not be resolved: Name or service not known
skip-name-resolve=1

relay_log=node-relay-bin
```

2. 编辑`/path/to/node1/docker-compose.yml`：
```
version: "3"
services:
    mysql:
        image: mysql:8.0
        volumes:
            - /path/to/node1/mysql/data:/var/lib/mysql
            - /path/to/node1/mysql/conf.d:/etc/mysql/conf.d
        ports:
            - 13306:3306
        command: mysqld --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci --init-connect='SET NAMES utf8mb4;' --innodb-flush-log-at-trx-commit=0
        networks:
            vpc:
                ipv4_address: 10.10.0.11
        environment:
            MYSQL_ROOT_PASSWORD: 12345678
        restart: always
networks:
    vpc:
        external:
            name: test_vpc

```

* /path/to/node1指向文件存储目录
* 10.10.0.11为节点IP地址
* 13306为映射的主机端口

3. 执行`docker-compose up`即会自动初始化数据库
4. 登录mysql后执行用户创建、授权、安装插件等
```
set sql_log_bin=0;
create user 'grp'@'%';
alter user 'grp'@'%' identified by 'grp12345678';
grant REPLICATION SLAVE on *.* to 'grp'@'%';
grant BACKUP_ADMIN on *.* to 'grp'@'%';
flush privileges;
set sql_log_bin=1;
change master to master_user='grp', master_password='grp12345678' for channel 'group_replication_recovery';
install plugin group_replication soname 'group_replication.so';
show plugins;
```
5. 编辑`path/to/mode1/mysql/conf.d/mysql.cnf`
```
[mysqld]

# resolv problem: IP address 'xx.xx.xx.xx' could not be resolved: Name or service not known
skip-name-resolve=1
log_error_verbosity=3

disabled_storage_engines="MyISAM,BLACKHOLE,FEDERATED,ARCHIVE,MEMORY"
server_id=1
gtid_mode=ON
enforce_gtid_consistency=ON
binlog_checksum=NONE

relay_log=node-relay-bin

# plugin_load_add='group_replication.so'
group_replication_group_name="7a73ff80-70df-11ea-94c2-0242ac120004"
group_replication_start_on_boot=off
group_replication_local_address="10.10.0.11:33061"
group_replication_group_seeds="10.10.0.11:33061,10.10.0.12:33061,10.10.0.13:33061"
group_replication_ip_whitelist="10.10.0.0/16"
group_replication_bootstrap_group=off
group_replication_recovery_get_public_key=on
#如果多主模式请使能下面两行
#group_replication_single_primary_mode=OFF
#group_replication_enforce_update_everywhere_checks=ON
```
6. 执行`docker-compose restart`重启容器
7. 登录mysql后启动GRP并创建测试库和表
```
SET global group_replication_bootstrap_group=ON;
START GROUP_REPLICATION;

# 创建测试库和表
CREATE DATABASE test;
USE test;
CREATE TABLE t1 (c1 INT PRIMARY KEY, c2 TEXT NOT NULL);
INSERT INTO t1 VALUES (1, 'Luis');
```
8. 编辑`path/to/mode1/mysql/conf.d/mysql.cnf`, `group_replication_start_on_boot`值变更为`on`；`group_replication_bootstrap_group`变更为`off`。运行`docker-compose restart`重启

## 创建并配置从节点

步骤与创建主写节点相同，不同之处：
1. `docker-compose.yml`文件中存储目录、IP地址、端口变更
2. 第5步`mysql.cnf`文件文件中`group_replication_start_on_boot`值为`on`；`group_replication_local_address`需要与`docker-compose.yml`文件配置的IP地址一致；server_id值根据创建的节点依次排序。
3. 不执行第7、8步

## build.sh 和 clean.sh

- build.sh用于自动创建GRP集群，脚本需要一个参数指示集群的节点数量，数量必须不少于3且不大于9。
- clean.sh用于清除创建的GRP集群及数据，脚本需要一个参数指示当前集群的节点数量，与build.sh脚本参数一致。

## 其它

* 查看集群成员: `SELECT * FROM performance_schema.replication_group_members;`


* 脑列问题：当primary节点因故离开集群后，内部将重新推举新的primary，此时若primary重启且group_replication_bootstrap_group=on+group_replication_start_on_boot=on则会重新组成集群，原集群脑列为两个集群。