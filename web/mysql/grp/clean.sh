#!/bin/sh -e

if [ $# -ne 1 ]; then
	echo "usage:$0 [containers count]"
	exit -1
fi

function clean() {
    path=`pwd`/node$1
    if [ -d $path ]; then
        cd $path
        docker-compose down
        cd -
        rm -rf $path
    fi
}

for ((i=1;i<=$1;i++)); do
    clean $i
done
