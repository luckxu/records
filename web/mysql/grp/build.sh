#!/bin/sh -e


function docker_create()
{
    path=`pwd`/node$1
    port=1000$1
    rpath=`echo "$path" | sed 's#\/#\\\/#g'`
    rm $path -rf
    mkdir $path -p
    cp docker-compose.yml $path
    sed -i "s/{id}/$1/g" $path/docker-compose.yml
    sed -i "s/{port}/$port/g" $path/docker-compose.yml
    sed -i "s/{path}/$rpath/g" $path/docker-compose.yml
    cd $path
    mkdir -p $path/mysql/conf.d
    echo -e "[mysqld]\nskip-name-resolve=1\nrelay_log=node-relay-bin\n" > $path/mysql/conf.d/mysql.cnf
    docker-compose up -d
    cd -
}

function mysql_init() {
    path=`pwd`/node$1
    port=1000$1
    mysql -u root -p12345678 -h 127.0.0.1 -P $port < init.sql
    \cp mysql.cnf $path/mysql/conf.d/mysql.cnf
    sed -i "s/{id}/$1/g" $path/mysql/conf.d/mysql.cnf
    cd $path && docker-compose restart && cd -
    sleep 5
    sed -i "s/group_replication_start_on_boot=off/group_replication_start_on_boot=on/g" $path/mysql/conf.d/mysql.cnf
    if [ $1 -eq 1 ]; then
        mysql -u root -p12345678 -h 127.0.0.1 -P $port < master.sql
        sed -i "s/group_replication_bootstrap_group=off/group_replication_bootstrap_group=on/g" $path/mysql/conf.d/mysql.cnf
    else
        mysql -u root -p12345678 -h 127.0.0.1 -P $port < slave.sql
    fi
}


if [ $1 -lt 3 -o $1 -gt 9 ];
then
    echo "error:<node count> must in [3,9]"
    exit -1;
fi 

for ((i=1;i<=$1;i++)); do
    docker_create $i
done

sleep 15

for ((i=1;i<=$1;i++)); do
    mysql_init $i
done
