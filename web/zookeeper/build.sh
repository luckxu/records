#!/bin/sh -e

if [ $# -ne 1 ]; then
    echo "usage:$0 <containers count>"
    exit -1
fi

function docker_create()
{
    path=`pwd`/zoo$1
    mkdir $path -p
    rm $path/* -rf
    cp docker-compose.yml $path
    sed -i "s/{id}/$1/g" $path/docker-compose.yml
    cd $path
    docker-compose up -d
    cd -
}

if [ $1 -lt 3 -o $1 -gt 9 ];
then
    echo "error:<containers count> must in [3,9]"
    exit -1;
fi 

docker network create test_vpc --driver=bridge --subnet=10.10.0.0/16 > /dev/null 2>&1 || :

for ((i=1;i<=$1;i++)); do
    docker_create $i
done

